package xyz.block.ftl.runtime;

import jakarta.inject.Singleton;

import org.jboss.logging.Logger;

import io.grpc.stub.StreamObserver;
import io.quarkus.grpc.GrpcService;
import xyz.block.ftl.v1.*;

@Singleton
@GrpcService
public class VerbHandler extends VerbServiceGrpc.VerbServiceImplBase {

    private static final Logger log = Logger.getLogger(VerbHandler.class);

    final VerbRegistry registry;

    public VerbHandler(VerbRegistry registry) {
        this.registry = registry;
    }

    @Override
    public void call(CallRequest request, StreamObserver<CallResponse> responseObserver) {
        try {
            var response = registry.invoke(request);
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
