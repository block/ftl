package xyz.block.ftl.hotreload;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.function.Consumer;

public class RunnerNotification {

    private static volatile Consumer<RunnerInfo> callback;
    private static volatile RunnerInfo info;
    private static final List<Runnable> runnerDetailsCallbacks = Collections.synchronizedList(new ArrayList<>());

    public static synchronized void setCallback(Consumer<RunnerInfo> callback) {
        if (RunnerNotification.callback != null) {
            throw new IllegalStateException("Callback already set");
        }
        if (info != null) {
            callback.accept(info);
            info = null;
        } else {
            RunnerNotification.callback = callback;
        }
    }

    public static synchronized void clearCallback() {
        RunnerNotification.callback = null;
    }

    public static synchronized void setRunnerInfo(RunnerInfo info) {
        if (callback != null) {
            callback.accept(info);
            callback = null;
        } else {
            RunnerNotification.info = info;
        }
    }

    public static synchronized void setRequiresNewRunnerDetails() {
        for (Runnable callback : runnerDetailsCallbacks) {
            callback.run();
        }
    }

    public static synchronized void onRunnerDetails(Runnable callback) {
        runnerDetailsCallbacks.add(callback);
    }

}
