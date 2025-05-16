package ftl.mysql;

import java.util.List;

import jakarta.inject.Inject;

import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;

import io.quarkus.test.common.QuarkusTestResource;
import io.quarkus.test.junit.QuarkusTest;
import xyz.block.ftl.java.test.internal.FTLTestResource;

@QuarkusTest
@QuarkusTestResource(FTLTestResource.class)
public class MySQLIntegrationTest {

    @Inject
    CreateRequestClient requestClient;

    @Inject
    GetRequestDataClient requestDataClient;

    @Test
    public void test() throws InterruptedException {
        requestClient.createRequest("foo");
        var data = requestDataClient.getRequestData();
        Assertions.assertEquals(List.of("foo"), data);

    }

}
