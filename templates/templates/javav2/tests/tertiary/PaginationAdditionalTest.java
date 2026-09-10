package org.openapis.tertiary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.tertiary.openapi.Helpers.HTTPBIN_URL;

import java.util.Iterator;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import java.util.concurrent.TimeUnit;
import org.openapis.tertiary.openapi.models.components.LimitOffsetConfig;
import org.openapis.tertiary.openapi.models.operations.PaginationCursorParamsResponse;
import org.openapis.tertiary.openapi.models.operations.PaginationLimitOffsetPageBodyResponse;
import org.openapis.tertiary.openapi.models.operations.PaginationLimitOffsetPageParamsResponse;

/**
 * Java 8 (okhttp) sync pagination coverage. The async arm is exercised by
 * {@link AsyncAdditionalTest}; these tests drive the blocking {@code callAsIterable}
 * path, which renders the Java 8 pageFetcher as
 * {@code unchecked(() -> operation.doRequest((pos == null ? req : req.withXxx(pos)))).get()}.
 * The first {@code iterator.next()} is the initial request where {@code pos == null}
 * (req passed through unchanged); subsequent pages take {@code pos != null}
 * (req.withXxx(pos)).
 */
public class PaginationAdditionalTest {

    @Test
    @Timeout(value = 20, unit = TimeUnit.SECONDS)
    void paginationLimitOffsetPageParams() throws Exception {
        // mirrors primary pagination-limit-offset-page-params
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        long serverLimit = 20L;

        Iterator<PaginationLimitOffsetPageParamsResponse> iterator = s.pagination()
                .paginationLimitOffsetPageParams()
                .page(1)
                .callAsIterable()
                .iterator();

        // pos == null: initial request, req passed through unchanged.
        assertTrue(iterator.hasNext());
        PaginationLimitOffsetPageParamsResponse res = iterator.next();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(serverLimit, res.res().get().resultArray().get().size());

        // pos != null: next page via req.withPage(pos).
        assertTrue(iterator.hasNext());
        PaginationLimitOffsetPageParamsResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(0, nextRes.res().get().resultArray().get().size());

        assertFalse(iterator.hasNext());
    }

    @Test
    @Timeout(value = 20, unit = TimeUnit.SECONDS)
    void paginationLimitOffsetPageBody() throws Exception {
        // mirrors primary pagination-limit-offset-page-body
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        long limit = 15L;

        LimitOffsetConfig request = LimitOffsetConfig.builder()
                .limit(limit)
                .page(1L)
                .build();

        Iterator<PaginationLimitOffsetPageBodyResponse> iterator = s.pagination()
                .paginationLimitOffsetPageBody()
                .request(request)
                .callAsIterable()
                .iterator();

        // pos == null: initial request with the concrete body, unchanged.
        assertTrue(iterator.hasNext());
        PaginationLimitOffsetPageBodyResponse res = iterator.next();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(limit, res.res().get().resultArray().get().size());

        // pos != null: next page via req.withPage(pos) on the concrete body.
        assertTrue(iterator.hasNext());
        PaginationLimitOffsetPageBodyResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(5, nextRes.res().get().resultArray().get().size());

        assertFalse(iterator.hasNext());
    }

    @Test
    @Timeout(value = 20, unit = TimeUnit.SECONDS)
    void paginationCursorParams() throws Exception {
        // mirrors primary pagination-cursor-params
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        Iterator<PaginationCursorParamsResponse> iterator = s.pagination()
                .paginationCursorParams()
                .cursor(-1)
                .callAsIterable()
                .iterator();

        // pos == null: initial request, req unchanged.
        assertTrue(iterator.hasNext());
        PaginationCursorParamsResponse res = iterator.next();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(15, res.res().get().resultArray().get().size());

        // pos != null: cursor advanced via req.withCursor(pos).
        assertTrue(iterator.hasNext());
        PaginationCursorParamsResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(5, nextRes.res().get().resultArray().get().size());

        assertTrue(iterator.hasNext());
        PaginationCursorParamsResponse lastRes = iterator.next();
        assertNotNull(lastRes);
        assertEquals(200, lastRes.statusCode());
        assertEquals(0, lastRes.res().get().resultArray().get().size());

        assertFalse(iterator.hasNext());
    }
}
