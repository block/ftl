package xyz.block.ftl.runtime;

import java.util.Map;

import jakarta.inject.Singleton;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.arc.Arc;

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

    // private Object executeQuery(String verb, String module, Object message, Class<?> returnType, boolean listReturnType,
    //         boolean mapReturnType, MetadataSQLQuery sqlQuery) throws Exception {
    //     // Convert parameters to JSON array
    //     String paramsJson = message != null ? mapper.writeValueAsString(message) : "[]";

    //     // Create result columns based on the return type
    //     List<ResultColumn> resultColumns = new ArrayList<>();
    //     // TODO: Add logic to extract column names from return type using reflection
    //     // This will require getting field names and SQL column mappings

    //     ExecuteQueryRequest.Builder reqBuilder = ExecuteQueryRequest.newBuilder()
    //         .setRawSql(sqlQuery.getQuery())
    //         .setParametersJson(paramsJson);

    //     // Set command type based on return type and SQL command
    //     if (sqlQuery.getCommand().equalsIgnoreCase("exec")) {
    //         reqBuilder.setCommandType(CommandType.COMMAND_TYPE_EXEC);
    //     } else if (listReturnType) {
    //         reqBuilder.setCommandType(CommandType.COMMAND_TYPE_MANY);
    //     } else {
    //         reqBuilder.setCommandType(CommandType.COMMAND_TYPE_ONE);
    //     }

    //     // Add result columns
    //     for (ResultColumn col : resultColumns) {
    //         reqBuilder.addResultColumns(col);
    //     }

    //     // Get query service endpoint from environment
    //     String queryEndpoint = System.getenv("FTL_QUERY_ADDRESS_" + module);
    //     if (queryEndpoint == null) {
    //         throw new RuntimeException("Query service endpoint not found for module " + module);
    //     }

    //     // Execute query
    //     List<ExecuteQueryResponse> responses = queryClient.executeQuery(queryEndpoint, reqBuilder.build());

    //     // Process responses
    //     if (responses.isEmpty()) {
    //         return null;
    //     }

    //     // For EXEC commands
    //     if (sqlQuery.getCommand().equalsIgnoreCase("exec")) {
    //         return null; // Or return affected rows count if needed
    //     }

    //     // For ONE/MANY commands
    //     List<Object> results = new ArrayList<>();
    //     for (ExecuteQueryResponse response : responses) {
    //         if (response.hasRowResults() && response.getRowResults().getJsonRows() != null
    //             && !response.getRowResults().getJsonRows().isEmpty()) {
    //             Object result = mapper.readValue(response.getRowResults().getJsonRows(), returnType);
    //             results.add(result);
    //         }
    //     }

    //     if (listReturnType) {
    //         return results;
    //     } else if (!results.isEmpty()) {
    //         return results.get(0);
    //     }
    //     return null;
    // }

    public static VerbClientHelper instance() {
        return Arc.container().instance(VerbClientHelper.class).get();
    }
}
