package xyz.block.ftl.runtime;

import java.io.Closeable;
import java.time.Duration;
import java.util.List;
import java.util.Map;

import xyz.block.ftl.LeaseFailedException;
import xyz.block.ftl.LeaseHandle;
import xyz.block.ftl.v1.GetDeploymentContextResponse;

public interface FTLRunnerConnection extends Closeable {
    String getEndpoint();

    byte[] getSecret(String secretName);

    byte[] getConfig(String secretName);

    byte[] callVerb(String name, String module, byte[] payload);

    void publishEvent(String topic, String callingVerbName, byte[] event, String key);

    /**
     * Execute a SQL query that returns a single row.
     *
     * @param dbName The database name
     * @param sql The SQL query
     * @param params The query parameters as a JSON array string
     * @param resultColumns Map of SQL column names to result field names
     * @return The query result as a JSON string, or null if no results
     */
    String executeQueryOne(String dbName, String sql, String paramsJson, Map<String, String> resultColumns);

    /**
     * Execute a SQL query that returns multiple rows.
     *
     * @param dbName The database name
     * @param sql The SQL query
     * @param params The query parameters as a JSON array string
     * @param resultColumns Map of SQL column names to result field names
     * @return List of query results as JSON strings
     */
    List<String> executeQueryMany(String dbName, String sql, String paramsJson, Map<String, String> resultColumns);

    /**
     * Execute a SQL statement that doesn't return results (INSERT, UPDATE, DELETE, etc.).
     *
     * @param dbName The database name
     * @param sql The SQL statement
     * @param params The statement parameters as a JSON array string
     * @return Number of affected rows, or -1 if not available
     */
    long executeQueryExec(String dbName, String sql, String paramsJson);

    LeaseHandle acquireLease(Duration duration, String... keys) throws LeaseFailedException;

    GetDeploymentContextResponse getDeploymentContext();

    @Override
    void close();
}
