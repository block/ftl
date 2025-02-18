package xyz.block.ftl.java.test.internal;

import io.grpc.stub.StreamObserver;
import xyz.block.ftl.v1.ControllerServiceGrpc;
import xyz.block.ftl.v1.GetDeploymentContextRequest;
import xyz.block.ftl.v1.GetDeploymentContextResponse;
import xyz.block.ftl.v1.PingRequest;
import xyz.block.ftl.v1.PingResponse;

public class TestDeploymentServer extends ControllerServiceGrpc.ControllerServiceImplBase {
    @Override
    public void ping(PingRequest request, StreamObserver<PingResponse> responseObserver) {
        responseObserver.onNext(PingResponse.newBuilder().build());
        responseObserver.onCompleted();
    }

    @Override
    public void getDeploymentContext(GetDeploymentContextRequest request,
            StreamObserver<GetDeploymentContextResponse> responseObserver) {
        //TODO: add a way to test secrets and other module context values
        responseObserver.onNext(GetDeploymentContextResponse.newBuilder().build());
    }

}
