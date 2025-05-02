package xyz.block.ftl.runtime;

import jakarta.enterprise.inject.Instance;
import jakarta.inject.Singleton;

import org.jboss.logging.Logger;

import com.fasterxml.jackson.databind.ObjectMapper;

import io.grpc.stub.StreamObserver;
import io.quarkus.grpc.GrpcService;
import xyz.block.ftl.v1.*;

@Singleton
@GrpcService
public class VerbHandler extends VerbServiceGrpc.VerbServiceImplBase {

    private static final Logger log = Logger.getLogger(VerbHandler.class);

    final VerbRegistry registry;
    final Instance<ObjectMapper> mapper;

    public VerbHandler(VerbRegistry registry, Instance<ObjectMapper> mapper) {
        this.registry = registry;
        this.mapper = mapper;
    }

    @Override
    public void call(CallRequest request, StreamObserver<CallResponse> responseObserver) {
        try {
            var response = registry.invoke(request, mapper.get());
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.errorf(e, "Verb invocation failed: %s.%s", request.getVerb().getModule(), request.getVerb().getName());
            responseObserver.onError(e);
        }
    }

    @Override
    public void ping(PingRequest request, StreamObserver<PingResponse> responseObserver) {
        responseObserver.onNext(PingResponse.newBuilder().build());
        responseObserver.onCompleted();
    }
}
