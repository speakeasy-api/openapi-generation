package org.openapis.openapi.hooks;

import java.net.http.HttpRequest;

import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.SchemeCustomSchemeAppID;
import org.openapis.openapi.utils.Helpers;
import org.openapis.openapi.utils.Hook.BeforeRequest;
import org.openapis.openapi.utils.Hook.BeforeRequestContext;

public class CustomSecurityHooks implements BeforeRequest {

    @Override
    public HttpRequest beforeRequest(BeforeRequestContext context, HttpRequest request) throws Exception {
        if (context.operationId().equals("customSchemeAppId")) {
            if (!context.securitySource().isPresent()) {
                throw new IllegalArgumentException("security source is not present");
            }

            Security security = (Security) context.securitySource().get().getSecurity();

            if (security.customSchemeAppId().isEmpty()) {
                throw new IllegalArgumentException("custom security is not defined");
            }

            SchemeCustomSchemeAppID customSec = security.customSchemeAppId().get();

            return Helpers.copy(request) //
                    .header("X-Security-App-Id", customSec.appId()) //
                    .header("X-Security-Secret", customSec.secret()) //
                    .build();
        }

        return request;
    }
}
