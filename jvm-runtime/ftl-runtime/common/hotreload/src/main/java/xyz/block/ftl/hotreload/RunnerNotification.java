package xyz.block.ftl.hotreload;

import java.util.concurrent.atomic.AtomicReference;

public class RunnerNotification {

    private static volatile RunnerCallback callback;
    private static volatile RunnerInfo info;
    private static final AtomicReference<String> currentDeploymentKey = new AtomicReference<>();

    public static String getDeploymentKey() {
        return currentDeploymentKey.get();
    }

    public static void newDeploymentKey(String key) {

        currentDeploymentKey.set(key);
        RunnerCallback callback;
        synchronized (RunnerNotification.class) {
            callback = RunnerNotification.callback;
        }
        if (callback != null) {
            callback.newRunnerDeployment(key);
        }
    }

    public static void setCallback(RunnerCallback callback) {
        RunnerInfo info;
        String runnerVersion;
        synchronized (RunnerNotification.class) {
            RunnerNotification.callback = callback;
            info = RunnerNotification.info;
            RunnerNotification.info = null;
            runnerVersion = RunnerNotification.currentDeploymentKey.get();
        }
        if (runnerVersion != null) {
            callback.newRunnerDeployment(runnerVersion);
        }
        if (info != null) {
            callback.runnerDetails(info);
        }
    }

    public static void reloadStarted() {
        RunnerCallback cb;
        synchronized (RunnerNotification.class) {
            cb = RunnerNotification.callback;
        }
        if (cb != null) {
            cb.reloadStarted();
        }
    }

    public static boolean setRunnerInfo(RunnerInfo info) {
        RunnerCallback callback;
        boolean outdated;
        synchronized (RunnerNotification.class) {
            outdated = !info.deployment().equals(currentDeploymentKey.get()) && currentDeploymentKey.get() != null;
            callback = RunnerNotification.callback;
            if (callback == null) {
                RunnerNotification.info = info;
            }
        }
        if (callback != null) {
            callback.runnerDetails(info);
        }
        return outdated;
    }

    public interface RunnerCallback {
        void runnerDetails(RunnerInfo info);

        void reloadStarted();

        void newRunnerDeployment(String deploymentKey);
    }
}
