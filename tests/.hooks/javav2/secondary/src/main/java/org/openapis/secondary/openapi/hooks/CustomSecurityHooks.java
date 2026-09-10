package org.openapis.secondary.openapi.hooks;

import org.openapis.secondary.openapi.models.operations.CustomSchemeAppIdSecurity;
import org.openapis.secondary.openapi.utils.Hook.BeforeRequest;
import org.openapis.secondary.openapi.utils.Hook.BeforeRequestContext;
import org.openapis.secondary.openapi.utils.transport.HttpRequest;

public class CustomSecurityHooks implements BeforeRequest {

    @Override
    public HttpRequest beforeRequest(BeforeRequestContext context, HttpRequest request) throws Exception {
        if (context.operationId().equals("customSchemeAppId")) {
            if (!context.securitySource().isPresent()) {
                throw new IllegalArgumentException("security source is not present");
            }
            CustomSchemeAppIdSecurity sec = (CustomSchemeAppIdSecurity) context.securitySource().get().getSecurity();
            return request.toBuilder() //
                    .setHeader("X-Security-App-Id", sec.appId()) //
                    .setHeader("X-Security-Secret", sec.secret()) //
                    .build();
        } else {
            return request;
        }
    }
}
