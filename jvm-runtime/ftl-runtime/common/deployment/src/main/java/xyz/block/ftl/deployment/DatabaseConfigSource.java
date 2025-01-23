package xyz.block.ftl.deployment;

import java.util.HashMap;
import java.util.Map;
import java.util.Properties;
import java.util.Set;

import org.eclipse.microprofile.config.spi.ConfigSource;

public class DatabaseConfigSource implements ConfigSource {

    private final Map<String, String> properties;

    public DatabaseConfigSource() {
        properties = new HashMap<>();
        var ds = Thread.currentThread().getContextClassLoader().getResourceAsStream("META-INF/ftl-sql-databases.txt");
        if (ds == null) {
            return;
        }
        try (var in = ds) {
            Properties p = new Properties();
            p.load(in);
            for (var name : p.stringPropertyNames()) {
                properties.put("quarkus.datasource." + name + ".db-kind", p.getProperty(name));
            }
            if (properties.size() == 1) {
                try {
                    // Temporary implicit support for hibernate orm
                    Thread.currentThread().getContextClassLoader()
                            .loadClass("io.quarkus.hibernate.orm.runtime.HibernateOrmRecorder");
                    properties.put("quarkus.hibernate-orm.datasource", p.entrySet().iterator().next().getKey().toString());
                } catch (ClassNotFoundException ignore) {

                }
            }
        } catch (Exception e) {
            throw new RuntimeException("Failed to load database properties", e);
        }

    }

    @Override
    public Set<String> getPropertyNames() {
        return properties.keySet();
    }

    @Override
    public String getValue(String propertyName) {
        return properties.get(propertyName);
    }

    @Override
    public String getName() {
        return "FTL Database Config Source";
    }
}
