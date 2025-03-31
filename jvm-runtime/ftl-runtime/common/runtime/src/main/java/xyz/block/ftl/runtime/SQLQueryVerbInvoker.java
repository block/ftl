package xyz.block.ftl.runtime;

import java.nio.charset.StandardCharsets;
import java.util.Arrays;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.ByteString;

import io.quarkus.arc.Arc;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;

public class SQLQueryVerbInvoker implements VerbInvoker {

    final ObjectMapper mapper;

    private final String dbName;
    private final String command;
    private final String rawSQL;
    private final String[] fields;
    private final String[] colToFieldName;
    private final Class<?> returnType;

    // Lazily initialized
    private volatile VerbClientHelper verbClientHelper;

    public SQLQueryVerbInvoker(String dbName, String command, String rawSQL,
            String[] fields, String[] colToFieldName, Class<?> returnType) {
        this.dbName = dbName;
        this.command = command;
        this.rawSQL = rawSQL;
        this.fields = fields;
        this.colToFieldName = colToFieldName;
        this.returnType = returnType;
        this.mapper = new ObjectMapper();
    }

    private VerbClientHelper getVerbClientHelper() {
        if (verbClientHelper == null) {
            synchronized (this) {
                if (verbClientHelper == null) {
                    if (Arc.container() == null) {
                        throw new IllegalStateException("Arc container is not initialized");
                    }
                    verbClientHelper = VerbClientHelper.instance();
                }
            }
        }
        return verbClientHelper;
    }

    @Override
    public CallResponse handle(CallRequest request) {
        try {
            String transactionId = null;
            for (var pair : request.getMetadata().getValuesList()) {
                if (pair.getKey().equals("ftl-transaction")) {
                    transactionId = pair.getValue();
                }
            }
            Object result = getVerbClientHelper().executeQuery(request.getBody().toByteArray(), dbName, command, rawSQL,
                    fields, colToFieldName, returnType, transactionId);
            if (result == null) {
                if (command.equals("exec")) {
                    return CallResponse.newBuilder().setBody(ByteString.copyFrom("{}", StandardCharsets.UTF_8)).build();
                } else {
                    return CallResponse.newBuilder().setError(CallResponse.Error.newBuilder()
                            .setMessage("No results found").build()).build();
                }
            }
            var mappedResponse = getVerbClientHelper().mapper.writer().writeValueAsBytes(result);
            return CallResponse.newBuilder().setBody(ByteString.copyFrom(mappedResponse)).build();
        } catch (Exception e) {
            return CallResponse.newBuilder()
                    .setError(CallResponse.Error.newBuilder()
                            .setMessage("Error executing SQL query: " + e.getMessage())
                            .build())
                    .build();
        }
    }

    @Override
    public String toString() {
        return "SQLQueryVerbInvoker{" +
                "dbName='" + dbName + '\'' +
                ", command='" + command + '\'' +
                ", rawSQL='" + rawSQL + '\'' +
                ", fields=" + Arrays.toString(fields) +
                ", colToFieldName=" + Arrays.toString(colToFieldName) +
                ", returnType=" + returnType +
                '}';
    }

}
