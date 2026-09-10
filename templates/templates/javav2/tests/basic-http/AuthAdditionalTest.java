package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.util.Optional;

import org.junit.jupiter.api.Test;

import org.openapis.openapi.models.operations.BasicAuthOptionalSecurity;
import org.openapis.openapi.models.operations.BasicAuthOptionalResponse;
import org.openapis.openapi.models.components.Security;

public class AuthAdditionalTest {
    @Test
    void testBasicAuthOperationOptional() throws Exception {
        CommonHelpers.recordTest("auth-basic-auth-operation-optional");


        SDK s = SDK.builder()
                .security(Security.builder()
                        .username("wrongUser")
                        .password("wrongPass")
                        .build())
                .build();

        BasicAuthOptionalResponse res = s.auth().basicAuthOptional()
                .security(BasicAuthOptionalSecurity.builder()
                        .username("testUser")
                        .password("testPass")
                        .build())
                .call();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.basicAuthResponse().isPresent());
        assertTrue(res.basicAuthResponse().get().authenticated());
        assertEquals("testUser", res.basicAuthResponse().get().user());


        SDK s2 = SDK.builder()
                .security(Security.builder()
                        .username("testUser")
                        .password("testPass")
                        .build())
                .build();

        assertThrows(Exception.class, () -> {
            s2.auth().basicAuthOptional()
                    .security(BasicAuthOptionalSecurity.builder().build())
                    .call();
        });
    }
}
