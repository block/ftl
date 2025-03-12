package xyz.block.ftl.runtime;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import jakarta.inject.Singleton;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.arc.Arc;

@Singleton
public class VerbClientHelper {

    final ObjectMapper mapper;

    public VerbClientHelper(ObjectMapper mapper) {
        this.mapper = mapper;
    }

    public static VerbClientHelper instance() {
        return Arc.container().instance(VerbClientHelper.class).get();
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

    public Object executeQuery(Object message, String dbName, String command, String rawSQL, String[] fields,
            String[] colToFieldName, Class<?> returnType) throws Exception {
        return executeQuery(mapper.writeValueAsBytes(message), dbName, command, rawSQL, fields, colToFieldName,
                returnType);
    }

    public Object executeQuery(byte[] request, String dbName, String command, String rawSQL, String[] fields,
            String[] colToFieldName, Class<?> returnType) throws Exception {
        if (dbName == null || dbName.isEmpty()) {
            throw new IllegalArgumentException("Database name must be provided in the SQLQueryClient annotation");
        }

        String paramsJson = getParamsJson(request, fields);

        String cmd = command.toLowerCase();
        if (cmd.equals("exec")) {
            executeExec(dbName, rawSQL, paramsJson);
            return null;
        }

        Object result = switch (cmd) {
            case "many" -> {
                yield executeMany(dbName, rawSQL, paramsJson, colToFieldName, returnType);
            }
            case "one" -> {
                yield executeOne(dbName, rawSQL, paramsJson, colToFieldName, returnType);
            }
            default -> {
                throw new IllegalArgumentException("Unknown command type: " + command);
            }
        };

        return result;
    }

    private Object executeOne(String dbName, String query, String paramsJson, String[] colToFieldName,
            Class<?> returnType) throws Exception {
        String jsonResult = FTLController.instance().executeQueryOne(dbName, query, paramsJson, colToFieldName);
        if (jsonResult == null || jsonResult.isEmpty()) {
            return null;
        }
        return mapper.readValue(jsonResult, returnType);
    }

    private List<Object> executeMany(String dbName, String query, String paramsJson, String[] colToFieldName,
            Class<?> returnType) throws Exception {
        List<String> results = FTLController.instance().executeQueryMany(dbName, query, paramsJson, colToFieldName);
        List<Object> resultList = new ArrayList<>();
        for (String result : results) {
            var r = mapper.readValue(result, returnType);
            resultList.add(r);
        }
        return resultList;
    }

    private void executeExec(String dbName, String query, String paramsJson) throws Exception {
        FTLController.instance().executeQueryExec(dbName, query, paramsJson);
    }

    private String getParamsJson(byte[] request, String[] fields) throws Exception {
        if (fields == null || fields.length == 0 || request == null || request.toString().isEmpty()) {
            return "[]";
        }
        Map<String, Object> requestJsonMap = mapper.readValue(request, new TypeReference<Map<String, Object>>() {
        });
        List<Object> params = new ArrayList<>();
        for (String field : fields) {
            if (requestJsonMap.containsKey(field)) {
                params.add(requestJsonMap.get(field));
            } else {
                params.add(null);
            }
        }
        return mapper.writeValueAsString(params);
    }
}
