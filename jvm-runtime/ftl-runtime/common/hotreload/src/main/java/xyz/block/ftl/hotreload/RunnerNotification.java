package xyz.block.ftl.hotreload;

public class RunnerNotification {

    private static volatile RunnerCallback callback;
    private static volatile RunnerInfo info;
    private static volatile long runnerVersion = 1;

    public static long getRunnerVersion() {
        return runnerVersion;
    }

    public static void setCallback(RunnerCallback callback) {
        RunnerInfo info;
        long runnerVersion;
        synchronized (RunnerNotification.class) {
            RunnerNotification.callback = callback;
            info = RunnerNotification.info;
            RunnerNotification.info = null;
            runnerVersion = RunnerNotification.runnerVersion;
        }
        if (runnerVersion != -1) {
            callback.newRunnerVersion(runnerVersion);
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

    public static void setRunnerVersion(long version) {
        RunnerCallback callback;
        synchronized (RunnerNotification.class) {
            runnerVersion = version;
            callback = RunnerNotification.callback;
        }
        if (callback != null) {
            callback.newRunnerVersion(version);
        }
    }

    public static boolean setRunnerInfo(RunnerInfo info) {
        RunnerCallback callback;
        boolean outdated;
        synchronized (RunnerNotification.class) {
            outdated = info.version() < runnerVersion;
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

        void newRunnerVersion(long version);
    }
}
