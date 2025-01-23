package xyz.block.ftl.hotreload;

import java.util.Map;

public record RunnerInfo(String address, String deployment, Map<String, String> databases) {

}
