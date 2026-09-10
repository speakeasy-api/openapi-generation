package org.openapis.openapi.hooks;

import org.apache.hc.core5.net.URIBuilder;
import org.openapis.openapi.utils.AsyncHook;
import org.openapis.openapi.utils.Helpers;
import org.openapis.openapi.utils.Hook.AfterErrorContext;
import org.openapis.openapi.utils.Hook.AfterSuccessContext;
import org.openapis.openapi.utils.Hook.BeforeRequestContext;
import org.openapis.openapi.utils.Blob;

import java.net.URISyntaxException;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;

final class AsyncTestHooks implements AsyncHook.BeforeRequest, AsyncHook.AfterSuccess, AsyncHook.AfterError {

    @Override
    public CompletableFuture<HttpRequest> beforeRequest(BeforeRequestContext context, HttpRequest request) {
        HttpRequest.Builder b = Helpers.copy(request);
        //noinspection UastIncorrectHttpHeaderInspection
        b.header("Idempotency-Key", "some-key");
        if ("testHooks".equals(context.operationId())) {
            URIBuilder ub = new URIBuilder(request.uri());
            ub.setParameter("someParam", "overriddenParam");
            try {
                b.uri(ub.build());
            } catch (URISyntaxException e) {
                return CompletableFuture.failedFuture(e);
            }
        } else if ("authorizationHeaderModification".equals(context.operationId())) {
            // overwrite the existing header by using setHeader
            request
                    .headers()
                    .firstValue("Authorization")
                    .ifPresent(value -> b.setHeader("Authorization", value + " modified"));
        }
        return CompletableFuture.completedFuture(b.build());
    }

    @Override
    public CompletableFuture<HttpResponse<Blob>> afterSuccess(AfterSuccessContext context, HttpResponse<Blob> response) {
        if ("testHooksAfterResponse".equals(context.operationId())) {
            return CompletableFuture.failedFuture(new RuntimeException("validation failed"));
        } else {
            return CompletableFuture.completedFuture(response);
        }

    }

    @Override
    public CompletableFuture<HttpResponse<Blob>> afterError(AfterErrorContext context, HttpResponse<Blob> response, Throwable error) {
        if ("testHooksError".equals(context.operationId())) {
            boolean is400StatusCode = Optional.ofNullable(response)
                    .map(resp -> resp.statusCode() == 400).orElse(false);
            if (!is400StatusCode) {
                return CompletableFuture.failedFuture(new IllegalStateException("expected status code 400"));
            }
            return CompletableFuture.failedFuture(new IllegalStateException("special test error case"));
        }
        if (error != null) {
            return CompletableFuture.failedFuture(error);
        }
        return CompletableFuture.completedFuture(response);
    }

}
