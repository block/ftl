package xyz.block.ftl.hotreload;

import java.nio.file.Path;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * This is a bit of a hack that allows for the hot reload process to wait for an async code generation process to complete
 *
 */
public class CodeGenNotification {

    private static final Map<Path, Long> lastModified = new HashMap<>();

    public static synchronized void updateLastModified(List<Path> path) {
        for (Path p : path) {
            lastModified.put(p, p.toFile().lastModified());
        }
        CodeGenNotification.class.notifyAll();
    }

    public static synchronized void waitForCodeGen() {
        boolean wait = false;
        for (Map.Entry<Path, Long> p : lastModified.entrySet()) {
            if (p.getKey().toFile().lastModified() != p.getValue()) {
                wait = true;
                break;
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
