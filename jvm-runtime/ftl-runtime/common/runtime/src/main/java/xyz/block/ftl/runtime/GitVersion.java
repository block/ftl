package xyz.block.ftl.runtime;

import java.io.IOException;
import java.util.Properties;

import org.jboss.logging.Logger;

public class GitVersion {

    private static final Logger log = Logger.getLogger(GitVersion.class);

    private static volatile boolean logged = false;

    public static void logVersion() {
        if (logged) {
            return;
        }
        synchronized (GitVersion.class) {
            if (logged) {
                return;
            }
            logged = true;
            Properties properties = new Properties();
            try {
                properties.load(GitVersion.class.getClassLoader().getResourceAsStream("git.properties"));
                String commit = properties.getProperty("git.commit.id.abbrev");
                String dirty = properties.getProperty("git.dirty");
                if ("true".equals(dirty)) {
                    log.infof("FTL Git Commit: %s (dirty)", commit);
                } else {
                    log.infof("FTL Git Commit: %s", commit);
                }
            } catch (IOException e) {
                log.errorf("Failed to load git.properties");
            }
        }
    }
}
