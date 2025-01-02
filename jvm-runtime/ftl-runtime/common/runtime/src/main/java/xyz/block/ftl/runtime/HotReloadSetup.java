package xyz.block.ftl.runtime;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Collections;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicBoolean;

import org.jboss.logging.Logger;

import io.quarkus.dev.spi.HotReplacementContext;
import io.quarkus.dev.spi.HotReplacementSetup;

public class HotReloadSetup implements HotReplacementSetup {

    static final Set<Path> existingMigrations = Collections.newSetFromMap(new ConcurrentHashMap<>());

    static volatile HotReplacementContext context;
    private static volatile String errorOutputPath;
    private static final String ERRORS_OUT = "errors.pb";

    @Override
    public void setupHotDeployment(HotReplacementContext hrc) {
        context = hrc;
        for (var dir : context.getSourcesDir()) {
            Path migrations = dir.resolve("db");
            if (Files.isDirectory(migrations)) {
                try (var stream = Files.walk(migrations)) {
                    stream.forEach(p -> {
                        existingMigrations.add(p);

                    });
                } catch (IOException e) {
                    throw new RuntimeException(e);
                }
            }
        }
    }

    static void doScan(boolean force) {
        if (context != null) {
            try {
                AtomicBoolean newForce = new AtomicBoolean();
                for (var dir : context.getSourcesDir()) {
                    Path migrations = dir.resolve("db");
                    if (Files.isDirectory(migrations)) {
                        try (var stream = Files.walk(migrations)) {
                            stream.forEach(p -> {
                                if (p.getFileName().toString().endsWith(".sql")) {
                                    if (existingMigrations.add(p)) {
                                        newForce.set(true);
                                    }
                                }
                            });
                        }
                    }
                }
                context.doScan(force || newForce.get());
            } catch (Exception e) {
                Logger.getLogger(HotReloadSetup.class).error("Failed to scan for changes", e);
            }
        }
    }
}
