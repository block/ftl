package xyz.block.ftl.deployment;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;

import org.jboss.logging.Logger;

import io.quarkus.deployment.util.FileUtil;

public class PackageOutput implements AutoCloseable {

    private static final Logger log = Logger.getLogger(PackageOutput.class);

    private final Path outputDir;
    private final String packageName;
    private final Map<String, StringBuilder> files = new HashMap<>();

    public PackageOutput(Path outputDir, String packageName) {
        this.outputDir = outputDir;
        this.packageName = packageName.replaceAll("\\.", "/");
    }

    public StringBuilder writeKotlin(String fileName) throws IOException {
        StringBuilder value = new StringBuilder();
        files.put(packageName + "/" + fileName + ".kt", value);
        return value;
    }

    public StringBuilder writeJava(String fileName) throws IOException {
        StringBuilder value = new StringBuilder();
        files.put(fileName, value);
        return value;
    }

    @Override
    public void close() {
        try {
            Path packageDir = outputDir.resolve(packageName);
            if (Files.exists(packageDir)) {
                try (var s = Files.list(packageDir)) {
                    s.forEach(path -> {
                        if (!files.containsKey(path.getFileName().toString())) {
                            try {
                                FileUtil.deleteIfExists(path);
                            } catch (IOException e) {
                                log.errorf(e, "Failed to delete path: %s", path);
                            }
                        }
                    });
                }
            }
            Files.createDirectories(packageDir);
            for (var e : files.entrySet()) {
                Path target = outputDir.resolve(e.getKey());
                byte[] fileBytes = e.getValue().toString().getBytes(StandardCharsets.UTF_8);
                if (Files.exists(target)) {
                    var existing = Files.readAllBytes(target);
                    if (Arrays.equals(fileBytes, existing)) {
                        continue;
                    }
                }
                Files.write(target, fileBytes);
            }
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }
}
