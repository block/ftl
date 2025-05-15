package xyz.block.ftl.hotreload;

import java.util.concurrent.atomic.AtomicLong;

public class RunnerNotification {

    private static volatile RunnerCallback callback;
    private static volatile RunnerInfo info;
    private static final AtomicLong requiredSchemaNumber = new AtomicLong(0);

    public static long getRequiredSchemaNumber() {
        return requiredSchemaNumber.get();
    }

    public static long schemaVersion(boolean newRunnerRequired) {
        if (newRunnerRequired) {
            return newRunnerRequired();
        }
        return requiredSchemaNumber.get();
    }

    public static long newRunnerRequired() {
        var ret = requiredSchemaNumber.incrementAndGet();
        RunnerCallback callback;
        synchronized (RunnerNotification.class) {
            callback = RunnerNotification.callback;
        }
        if (callback != null) {
            callback.newSchemaNumber(ret);
        }
        return ret;
    }

    public static void setCallback(RunnerCallback callback) {
        RunnerInfo info;
        long schemaVersion;
        synchronized (RunnerNotification.class) {
            RunnerNotification.callback = callback;
            info = RunnerNotification.info;
            RunnerNotification.info = null;
            schemaVersion = RunnerNotification.requiredSchemaNumber.get();
        }
        if (schemaVersion > 0) {
            callback.newSchemaNumber(schemaVersion);
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
        boolean outdated = false;
        synchronized (RunnerNotification.class) {
            callback = RunnerNotification.callback;
            if (callback == null) {
                RunnerNotification.info = info;
            }
        }
        if (callback != null) {
            outdated = callback.runnerDetails(info);
        }
        return outdated;
    }

    public interface RunnerCallback {
        boolean runnerDetails(RunnerInfo info);

        void reloadStarted();

        void newSchemaNumber(long seq);
    }
}
