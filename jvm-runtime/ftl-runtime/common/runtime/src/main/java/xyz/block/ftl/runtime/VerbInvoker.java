package xyz.block.ftl.runtime;

import com.fasterxml.jackson.databind.ObjectMapper;

import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;

public interface VerbInvoker {

    CallResponse handle(CallRequest in, ObjectMapper mapper);
}
