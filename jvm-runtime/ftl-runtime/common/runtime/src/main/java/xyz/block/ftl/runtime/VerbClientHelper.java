package xyz.block.ftl.runtime;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import jakarta.inject.Singleton;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.arc.Arc;
import xyz.block.ftl.schema.v1.MetadataSQLQuery;

@Singleton
public class VerbClientHelper {

    final ObjectMapper mapper;

    public VerbClientHelper(ObjectMapper mapper) {
        this.mapper = mapper;
    }

    public Object call(String verb, String module, Object message, Class<?> returnType, boolean listReturnType,
            boolean mapReturnType) {
        try {
            if (message == null) {
                //Unit must be an empty map
                //TODO: what about optional?
                message = Map.of();
            }

            var result = FTLController.instance().callVerb(verb, module, mapper.writeValueAsBytes(message));
            if (result == null) {
                return null;
            } else if (listReturnType) {
                return mapper.readerForArrayOf(returnType).readValue(result);
            } else if (mapReturnType) {
                return mapper.readerForMapOf(returnType).readValue(result);
            } else if (returnType == JsonNode.class) {
                return mapper.readTree(result);
            }
            return mapper.readerFor(returnType).readValue(result);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    /**
     * Execute a SQL query and return the results.
     *
     * @param dbName The database name
     * @param message The query parameters
     * @param returnType The type to deserialize the results to
     * @param listReturnType Whether the return type is a list
     * @param mapReturnType Whether the return type is a map
     * @param sqlQuery The SQL query metadata
     * @return The query results
     * @throws Exception If an error occurs
     */
    public Object executeQuery(String dbName, Object message, Class<?> returnType, boolean listReturnType,
            boolean mapReturnType, MetadataSQLQuery sqlQuery) throws Exception {
        // Convert parameters to JSON
        String paramsJson = message != null ? mapper.writeValueAsString(message) : "[]";

        // Create result columns map
        Map<String, String> resultColumns = new HashMap<>();
        // TODO: Add logic to extract column names from return type using reflection
        // This will require getting field names and SQL column mappings

        // Execute query based on command type
        if (sqlQuery.getCommand().equalsIgnoreCase("exec")) {
            // For EXEC commands
            long rowsAffected = FTLController.instance().executeQueryExec(dbName, sqlQuery.getQuery(), paramsJson);
            return rowsAffected;
        } else if (listReturnType) {
            // For MANY commands
            List<String> jsonResults = FTLController.instance().executeQueryMany(dbName, sqlQuery.getQuery(), paramsJson,
                    resultColumns);
            if (jsonResults.isEmpty()) {
                return listReturnType ? List.of() : null;
            }

            List<Object> results = new ArrayList<>();
            for (String json : jsonResults) {
                Object result = mapper.readValue(json, returnType);
                results.add(result);
            }
            return results;
        } else {
            // For ONE commands
            String jsonResult = FTLController.instance().executeQueryOne(dbName, sqlQuery.getQuery(), paramsJson,
                    resultColumns);
            if (jsonResult == null || jsonResult.isEmpty()) {
                return null;
            }
            return mapper.readValue(jsonResult, returnType);
        }
    }

    public static VerbClientHelper instance() {
        return Arc.container().instance(VerbClientHelper.class).get();
    }
}
