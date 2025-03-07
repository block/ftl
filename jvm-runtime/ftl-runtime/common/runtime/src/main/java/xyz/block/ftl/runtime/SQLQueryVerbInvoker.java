package xyz.block.ftl.runtime;

import java.util.Arrays;

import org.jboss.logging.Logger;

import com.google.protobuf.ByteString;

import io.quarkus.arc.Arc;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;

/**
 * A VerbInvoker implementation that handles SQL query verbs.
 */
public class SQLQueryVerbInvoker implements VerbInvoker {

    private static final Logger log = Logger.getLogger(SQLQueryVerbInvoker.class);

    private final String dbName;
    private final String command;
    private final String rawSQL;
    private final String[] fields;
    private final String[] colToFieldName;
    private final Class<?> returnType;
    private final Class<?> paramType;

    // Lazily initialized
    private volatile VerbClientHelper verbClientHelper;

    public SQLQueryVerbInvoker(String dbName, String command, String rawSQL,
            String[] fields, String[] colToFieldName, Class<?> returnType, Class<?> paramType) {
        this.dbName = dbName;
        this.command = command;
        this.rawSQL = rawSQL;
        this.fields = fields;
        this.colToFieldName = colToFieldName;
        this.returnType = returnType;
        this.paramType = paramType;
    }

    private VerbClientHelper getVerbClientHelper() {
        if (verbClientHelper == null) {
            synchronized (this) {
                if (verbClientHelper == null) {
                    if (Arc.container() == null) {
                        throw new IllegalStateException("Arc container is not initialized");
                    }
                    verbClientHelper = Arc.container().instance(VerbClientHelper.class).get();
                }
            }
        }
        return verbClientHelper;
    }

    @Override
    public CallResponse handle(CallRequest request) {
        try {
            log.debugf("Executing SQL query: %s, command: %s, dbName: %s", rawSQL, command, dbName);

            // Parse the request body if there is a parameter type
            Object param = null;
            if (paramType != null && request.getBody() != null && !request.getBody().isEmpty()) {
                param = getVerbClientHelper().mapper.readValue(request.getBody().toByteArray(), paramType);
            }

            // Execute the query
            Object result = getVerbClientHelper().executeQuery(dbName, param, returnType, command, rawSQL, fields,
                    colToFieldName);

            // Convert the result to JSON and return it
            byte[] responseBytes = getVerbClientHelper().mapper.writeValueAsBytes(result);
            return CallResponse.newBuilder()
                    .setBody(ByteString.copyFrom(responseBytes))
                    .build();

        } catch (Exception e) {
            log.errorf(e, "Error executing SQL query: %s", e.getMessage());
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
                ", paramType=" + paramType +
                '}';
    }
}
