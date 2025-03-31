package xyz.block.ftl.hotreload;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * This is a bit of a hack that allows for the hot reload process to wait for an async code generation process to complete
 *
 */
public class CodeGenNotification {

    private static final Map<Path, Long> lastModified = new HashMap<>();
    private static List<Path> schemaDirs = new ArrayList<>();

    public static synchronized void updateLastModified(List<Path> sd, List<Path> path) {
        for (Path p : path) {
            lastModified.put(p, p.toFile().lastModified());
        }
        schemaDirs = new ArrayList<>(sd);
        CodeGenNotification.class.notifyAll();
    }

    public static synchronized void waitForCodeGen(boolean schemaChanged) {
        boolean wait = schemaChanged && schemaDirs.isEmpty();
        for (Map.Entry<Path, Long> p : lastModified.entrySet()) {
            if (p.getKey().toFile().lastModified() != p.getValue()) {
                wait = true;
                break;
            }
        }
        for (Path p : schemaDirs) {
            if (Files.isDirectory(p)) {
                try (var stream = Files.newDirectoryStream(p)) {
                    for (Path path : stream) {
                        if (path.getFileName().toString().endsWith(".pb")) {
                            if (!lastModified.containsKey(path)) {
                                // New schema file
                                wait = true;
                                break;
                            }
                        }
                    }
                } catch (IOException e) {
                    // ignore
                }
            }
        }
        if (wait) {
            try {
                CodeGenNotification.class.wait();
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }
    }

}
