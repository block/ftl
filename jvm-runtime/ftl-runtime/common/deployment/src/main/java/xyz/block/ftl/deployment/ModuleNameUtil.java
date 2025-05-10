package xyz.block.ftl.deployment;

import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class ModuleNameUtil {
    private static final String FILE = "ftl.toml";
    private static final String PROJECT_FILE = "ftl-project.toml";
    private static final String MODULE_PATTERN = "module\\s*=\\s*\"([^\"]+)\"";
    private static final String PROJECT_PATTERN = "name\\s*=\\s*\"([^\"]+)\"";

    public static String getModuleName() {
        Path current = Paths.get("");
        while (current != null && Files.isDirectory(current)) {
            Path file = current.resolve(FILE);
            if (Files.exists(file)) {
                try {
                    String content = Files.readString(file);
                    Matcher matcher = Pattern.compile(MODULE_PATTERN).matcher(content);
                    if (matcher.find()) {
                        return matcher.group(1);
                    }
                } catch (Exception e) {
                    throw new RuntimeException("Failed to read module name from " + file, e);
                }
            }
            current = current.getParent();
        }
        return "unknown";
    }

    public static String getRealmName() {
        Path current = Paths.get("");
        while (current != null && Files.isDirectory(current)) {
            Path file = current.resolve(PROJECT_FILE);
            if (Files.exists(file)) {
                try {
                    String content = Files.readString(file);
                    Matcher matcher = Pattern.compile(PROJECT_PATTERN).matcher(content);
                    if (matcher.find()) {
                        return matcher.group(1);
                    }
                } catch (Exception e) {
                    throw new RuntimeException("Failed to read project name from " + file, e);
                }
            }
            current = current.getParent();
        }
        return "ftl";
    }
}
