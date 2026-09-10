package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.openapi.utils.Utils.recordTest;

import java.util.List;
import org.junit.jupiter.api.Test;
import org.openapis.openapi.utils.Utils;

public class CustomCodeAdditionalTest {
    @Test
    public void testCustomCodeRegionSdkMethodWithImports() throws Exception {
        recordTest("custom-code-region-sdk-method-with-imports");
        SDK sdk = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        boolean result = sdk.customHealthCheck();
        assertTrue(result);
    }

    @Test
    public void testCustomCodeRegionSubSdkMethodWithImports() throws Exception {
        recordTest("custom-code-region-sub-sdk-method-with-imports");
        SDK sdk = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        boolean result = sdk.health().customHealthCheck();
        assertTrue(result);
    }


    @Test
    public void testCustomCodeRegionInModel() throws Exception {
        recordTest("custom-code-region-model-method-with-imports");
        var testHttpClient = Utils.createTestHTTPClient("unwieldyGet");
        SDK sdk = SDK.builder()
                .serverURL(Utils.environmentVariable("TEST_SERVER_URL", ""))
                .client(testHttpClient)
            .build();
        List<String> result = sdk.unwieldyGetFlattened();
        assertEquals(5, result.size());
    }

}
