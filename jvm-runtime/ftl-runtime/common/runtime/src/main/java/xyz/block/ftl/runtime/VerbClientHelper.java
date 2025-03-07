package xyz.block.ftl.runtime;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import jakarta.inject.Singleton;

import org.jboss.logging.Logger;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.arc.Arc;

@Singleton
public class VerbClientHelper {

    final ObjectMapper mapper;
    private static final Logger log = Logger.getLogger(VerbClientHelper.class);

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
     * Call a verb with automatic detection of return type.
     *
     * @param verb The verb name
     * @param module The module name
     * @param message The message to send
     * @param returnType The return type
     * @return The result of the verb call
     */
    public Object call(String verb, String module, Object message, Class<?> returnType) {
        // Determine if the return type is a list or map
        boolean listReturnType = List.class.isAssignableFrom(returnType);
        boolean mapReturnType = Map.class.isAssignableFrom(returnType);

        return call(verb, module, message, returnType, listReturnType, mapReturnType);
    }

    /**
     * Execute a SQL query and return the results.
     *
     * @param dbName The database name
     * @param message The query parameters
     * @param returnType The type to deserialize the results to
     * @param command The SQL command type (one, many, exec)
     * @param rawSQL The raw SQL query
     * @param fields The field names from the SQLQueryClient annotation
     * @param colToFieldName The column to field name mappings from the SQLQueryClient annotation
     * @return The query results
     * @throws Exception If an error occurs
     */
    public Object executeQuery(String dbName, Object message, Class<?> returnType,
            String command, String rawSQL, String[] fields, String[] colToFieldName) throws Exception {

        // Convert parameters to JSON
        String paramsJson = convertParamsToJson(message);

        // Create result columns map
        Map<String, String> resultColumns = buildResultColumnsMap(returnType, fields, colToFieldName);

        // Get the database name from the query metadata if not provided
        String effectiveDbName = dbName;
        if (effectiveDbName == null || effectiveDbName.isEmpty()) {
            throw new IllegalArgumentException("Database name must be provided in the SQLQueryClient annotation");
        }

        // Execute query based on command type
        String lowerCommand = command.toLowerCase();
        switch (lowerCommand) {
            case "exec":
                return executeExec(effectiveDbName, rawSQL, paramsJson);
            case "many":
                return executeMany(effectiveDbName, rawSQL, paramsJson, resultColumns, returnType);
            case "one":
                return executeOne(effectiveDbName, rawSQL, paramsJson, resultColumns, returnType);
            default:
                throw new IllegalArgumentException("Unknown command type: " + command);
        }
    }

    /**
     * Convert parameters to JSON.
     *
     * @param message The parameters object
     * @return JSON string representation of parameters
     * @throws Exception If an error occurs
     */
    private String convertParamsToJson(Object message) throws Exception {
        if (message == null) {
            return "[]";
        } else if (message instanceof List) {
            // If the message is already a list, convert it directly
            return mapper.writeValueAsString(message);
        } else {
            // For single parameter objects, extract fields and create a parameter array
            List<Object> params = new ArrayList<>();
            for (var field : message.getClass().getDeclaredFields()) {
                field.setAccessible(true);
                params.add(field.get(message));
            }
            return mapper.writeValueAsString(params);
        }
    }

    /**
     * Build a map of column names to field names for the query results.
     *
     * @param returnType The return type class
     * @param fields The field names from the SQLQueryClient annotation
     * @param colToFieldName The column to field name mappings from the SQLQueryClient annotation
     * @return A map of column names to field names
     */
    private Map<String, String> buildResultColumnsMap(Class<?> returnType, String[] fields, String[] colToFieldName) {
        Map<String, String> resultColumns = new HashMap<>();

        // Process column to field name mappings if provided
        if (colToFieldName != null && colToFieldName.length > 0) {
            for (String mapping : colToFieldName) {
                String[] parts = mapping.split(",", 2);
                if (parts.length == 2) {
                    resultColumns.put(parts[0], parts[1]);
                }
            }
        }
        // Process fields if provided
        else if (fields != null && fields.length > 0) {
            for (String field : fields) {
                resultColumns.put(field, field);
            }
        }
        // Fall back to reflection if no mappings provided
        else if (returnType != null && returnType != Void.class) {
            for (var field : returnType.getDeclaredFields()) {
                resultColumns.put(field.getName(), field.getName());
            }
        }

        return resultColumns;
    }

    private Object executeOne(String dbName, String query, String paramsJson, Map<String, String> resultColumns,
            Class<?> returnType) throws Exception {
        String jsonResult = FTLController.instance().executeQueryOne(dbName, query, paramsJson, resultColumns);
        if (jsonResult == null || jsonResult.isEmpty()) {
            return null;
        }
        return mapper.readValue(jsonResult, returnType);
    }

    /**
     * Execute a SQL query that returns multiple rows.
     * Similar to the Go client's Many function.
     *
     * @param dbName The database name
     * @param query The SQL query
     * @param paramsJson The query parameters as a JSON array string
     * @param resultColumns Map of SQL column names to result field names
     * @param returnType The type to deserialize the results to
     * @return List of query results
     * @throws Exception If an error occurs
     */
    private List<Object> executeMany(String dbName, String query, String paramsJson, Map<String, String> resultColumns,
            Class<?> returnType) throws Exception {
        List<String> jsonResults = FTLController.instance().executeQueryMany(dbName, query, paramsJson, resultColumns);
        if (jsonResults.isEmpty()) {
            return List.of();
        }

        // If the return type is a List, we need to use the component type for deserialization
        Class<?> componentType = returnType;
        if (List.class.isAssignableFrom(returnType)) {
            // This is a simplification - in a real implementation, you'd use reflection
            // to get the generic type parameters
            // For now, we'll just use Object as the component type
            componentType = Object.class;
        }

        // Combine all JSON results into a single JSON array
        StringBuilder jsonArrayBuilder = new StringBuilder("[");
        for (int i = 0; i < jsonResults.size(); i++) {
            jsonArrayBuilder.append(jsonResults.get(i));
            if (i < jsonResults.size() - 1) {
                jsonArrayBuilder.append(",");
            }
        }
        jsonArrayBuilder.append("]");
        String jsonArray = jsonArrayBuilder.toString();

        // Use readerForArrayOf for better handling of collections
        return (List<Object>) mapper.readerForArrayOf(componentType).readValue(jsonArray);
    }

    private long executeExec(String dbName, String query, String paramsJson) throws Exception {
        return FTLController.instance().executeQueryExec(dbName, query, paramsJson);
    }

    public static VerbClientHelper instance() {
        return Arc.container().instance(VerbClientHelper.class).get();
    }
}
