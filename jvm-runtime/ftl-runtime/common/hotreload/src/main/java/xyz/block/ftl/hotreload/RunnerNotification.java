package xyz.block.ftl.hotreload;

import java.util.concurrent.atomic.AtomicLong;

public class RunnerNotification {

    private static volatile RunnerCallback callback;
    private static volatile RunnerInfo info;
    private static final AtomicLong runnerVersion = new AtomicLong(0);

    public static long getRunnerVersion() {
        return runnerVersion.get();
    }

    public static long incrementRunnerVersion() {
        long incremented = runnerVersion.incrementAndGet();
        RunnerCallback callback;
        synchronized (RunnerNotification.class) {
            callback = RunnerNotification.callback;
        }
        if (callback != null) {
            callback.newRunnerVersion(incremented);
        }
        return incremented;
    }

    public static void setCallback(RunnerCallback callback) {
        RunnerInfo info;
        long runnerVersion;
        synchronized (RunnerNotification.class) {
            RunnerNotification.callback = callback;
            info = RunnerNotification.info;
            RunnerNotification.info = null;
            runnerVersion = RunnerNotification.runnerVersion.get();
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

    public static boolean setRunnerInfo(RunnerInfo info) {
        RunnerCallback callback;
        boolean outdated;
        synchronized (RunnerNotification.class) {
            outdated = info.version() < runnerVersion.get();
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
