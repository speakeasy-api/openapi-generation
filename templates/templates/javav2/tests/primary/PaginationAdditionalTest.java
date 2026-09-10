package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.openapi.CommonHelpers.recordTest;

import java.io.IOException;
import java.io.InputStream;
import java.net.URISyntaxException;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.openapi.models.operations.PaginationCursorBodyRequestBody;
import org.openapis.openapi.models.operations.PaginationCursorBodyResponse;
import org.openapis.openapi.models.operations.PaginationCursorNonNumericEmptyStringResponse;
import org.openapis.openapi.models.operations.PaginationCursorNonNumericNullableResponse;
import org.openapis.openapi.models.operations.PaginationCursorNonNumericResponse;
import org.openapis.openapi.models.operations.PaginationCursorNonNumericWithLimitRes;
import org.openapis.openapi.models.operations.PaginationCursorNonNumericWithLimitResponse;
import org.openapis.openapi.models.operations.PaginationCursorParamsResponse;
import org.openapis.openapi.models.operations.PaginationCursorResponseEnvelopeRes;
import org.openapis.openapi.models.operations.PaginationLimitOffsetDeepOutputsPageBodyResponse;
import org.openapis.openapi.models.operations.PaginationLimitOffsetDefaultOffsetBodyResponse;
import org.openapis.openapi.models.operations.PaginationLimitOffsetOffsetBodyResponse;
import org.openapis.openapi.models.operations.PaginationLimitOffsetOffsetParamsResponse;
import org.openapis.openapi.models.operations.PaginationLimitOffsetOptionalPageParamsResponse;
import org.openapis.openapi.models.operations.PaginationLimitOffsetPageBodyNullableResponse;
import org.openapis.openapi.models.operations.PaginationLimitOffsetPageBodyResponse;
import org.openapis.openapi.models.operations.PaginationOffsetNullableResponse;
import org.openapis.openapi.models.operations.PaginationLimitOffsetPageParamsResponse;
import org.openapis.openapi.models.operations.PaginationURLParamsResponse;
import org.openapis.openapi.models.operations.PaginationCursorNullableLimitResponse;
import org.openapis.openapi.models.operations.PaginationWithRetriesResponse;
import org.openapis.openapi.models.operations.PaginationWrappedOptionalBodyRequest;
import org.openapis.openapi.models.operations.PaginationWrappedOptionalBodyResponse;
import org.openapis.openapi.models.shared.LimitOffsetConfig;
import org.openapis.openapi.models.shared.LimitOffsetConfigWithDefaults;
import org.openapis.openapi.utils.HTTPClient;
import org.openapis.openapi.utils.SpeakeasyHTTPClient;
import org.openapis.openapi.utils.Utils;

public class PaginationAdditionalTest {

    @Test
    void paginationLimitOffsetPageParams() throws Exception {
        recordTest("pagination-limit-offset-page-params");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        long serverLimit = 20l;

        Iterator<PaginationLimitOffsetPageParamsResponse> iterator = s.pagination().paginationLimitOffsetPageParams()
                .page(1)
                .callAsIterable()
                .iterator();
        assertTrue(iterator.hasNext());
        PaginationLimitOffsetPageParamsResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(serverLimit, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationLimitOffsetPageParamsResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(0, nextRes.res().get().resultArray().size());

        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationLimitOffsetPageParamsUnwrapped() throws Exception {
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        long serverLimit = 20l;

        Stream<Long> results = s.pagination().paginationLimitOffsetPageParams().page(1).callAsStreamUnwrapped();
        assertEquals(serverLimit, results.count());
    }

    @Test
    void paginationLimitOffsetPageBody() throws Exception {
        recordTest("pagination-limit-offset-page-body");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        Long page = 1l;
        Long limit = 15l;

        LimitOffsetConfig request = LimitOffsetConfig.builder().limit(limit).page(page).build();
        Iterator<PaginationLimitOffsetPageBodyResponse> iterator = s.pagination().paginationLimitOffsetPageBody()
                .request(request)
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationLimitOffsetPageBodyResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(limit, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationLimitOffsetPageBodyResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertTrue(limit > nextRes.res().get().resultArray().size());

        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationLimitOffsetPageBodyNullable() throws Exception {
        recordTest("pagination-limit-offset-page-body-nullable");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        // first request sends a null body; later pages materialize one with
        // the advanced page (wire sequence enforced by the test service)
        Iterator<PaginationLimitOffsetPageBodyNullableResponse> iterator = s.pagination()
                .paginationLimitOffsetPageBodyNullable()
                .request(null)
                .callAsIterable().iterator();

        List<List<Long>> pages = new ArrayList<>();
        while (iterator.hasNext()) {
            PaginationLimitOffsetPageBodyNullableResponse res = iterator.next();
            assertEquals(200, res.statusCode());
            assertTrue(res.res().isPresent());
            pages.add(res.res().get().resultArray());
        }

        // with a null request the limit is unknown, so a final empty page is
        // fetched before stopping
        assertEquals(
                List.of(
                        List.of(0L, 1L, 2L, 3L, 4L, 5L, 6L),
                        List.of(7L, 8L, 9L, 10L, 11L, 12L, 13L),
                        List.of(14L, 15L, 16L, 17L, 18L, 19L),
                        List.of()),
                pages);
    }

    @Test
    void paginationWrappedOptionalBody() throws Exception {
        recordTest("pagination-wrapped-optional-body");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        // the body is omitted entirely; later pages materialize one carrying
        // the advanced offset
        Iterator<PaginationWrappedOptionalBodyResponse> iterator = s.pagination().paginationWrappedOptionalBody()
                .request(PaginationWrappedOptionalBodyRequest.builder().build())
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationWrappedOptionalBodyResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(20, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationWrappedOptionalBodyResponse emptyRes = iterator.next();
        assertNotNull(emptyRes);
        assertEquals(200, emptyRes.statusCode());
        assertTrue(emptyRes.res().get().resultArray().isEmpty());

        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationLimitOffsetDeepOutputsPageBody() throws Exception {
        recordTest("pagination-limit-offset-deep-outputs-page-body");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        Long page = 1l;
        Long limit = 15l;

        LimitOffsetConfig request = LimitOffsetConfig.builder().limit(limit).page(page).build();
        Iterator<PaginationLimitOffsetDeepOutputsPageBodyResponse> iterator = s.pagination()
                .paginationLimitOffsetDeepOutputsPageBody()
                .request(request)
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationLimitOffsetDeepOutputsPageBodyResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(limit, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationLimitOffsetDeepOutputsPageBodyResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertTrue(limit > nextRes.res().get().resultArray().size());

        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationLimitOffsetPageBodyUnwrapped() throws Exception {
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        long page = 1;
        long limit = 15;

        LimitOffsetConfig request = LimitOffsetConfig.builder().limit(limit).page(page).build();
        Stream<Long> results = s.pagination().paginationLimitOffsetPageBody()
                .request(request)
                .callAsStreamUnwrapped();
        assertEquals(20, results.count());
    }

    @Test
    void paginationLimitOffsetOffsetParams() throws Exception {
        recordTest("pagination-limit-offset-offset-params");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        Long limit = 15l;
        Long offset = 0l;

        Iterator<PaginationLimitOffsetOffsetParamsResponse> iterator = s.pagination().paginationLimitOffsetOffsetParams()
                .limit(limit)
                .offset(offset)
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationLimitOffsetOffsetParamsResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(limit, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationLimitOffsetOffsetParamsResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(20l - limit, nextRes.res().get().resultArray().size());

        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationLimitOffsetOffsetUnwrapped() throws Exception {
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        long limit = 15;
        long offset = 0;

        Stream<Long> results = s.pagination().paginationLimitOffsetOffsetParams()
                .limit(limit) //
                .offset(offset)//
                .callAsStreamUnwrapped();
        assertEquals(20, results.count());
    }

    @Test
    void paginationLimitOffsetOffsetBody() throws Exception {
        recordTest("pagination-limit-offset-offset-body");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        long limit = 15;
        long offset = 0;

        LimitOffsetConfig request = LimitOffsetConfig.builder() //
                .limit(limit) //
                .offset(offset) //
                .build();

        Iterator<PaginationLimitOffsetOffsetBodyResponse> iterator = s.pagination().paginationLimitOffsetOffsetBody()
                .request(request)
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationLimitOffsetOffsetBodyResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(limit, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationLimitOffsetOffsetBodyResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertTrue(limit > nextRes.res().get().resultArray().size());

        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationLimitOffsetOffsetBodyUnwrapped() throws Exception {
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        long limit = 15;
        long offset = 0;

        LimitOffsetConfig request = LimitOffsetConfig.builder() //
                .limit(limit) //
                .offset(offset) //
                .build();
        Stream<Long> results = s.pagination().paginationLimitOffsetOffsetBody()
                .request(request)
                .callAsStreamUnwrapped();
        assertEquals(20, results.count());
    }

    @Test
    void paginationLimitOffsetNullable() throws Exception {
        recordTest("pagination-limit-offset-nullable");

        // Calling .callAsIterable() with no args means the SDK builds the
        // request with JsonNullable.of(null) for the required+nullable
        // offset/limit. OffsetTracker takes long primitives, so without a
        // null-safe fallback in the tracker init it would NPE on auto-unbox
        // before we ever issue a request.
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        long serverLimit = 10L;

        Iterator<PaginationOffsetNullableResponse> iterator = s.pagination()
                .paginationOffsetNullable()
                .callAsIterable().iterator();

        assertTrue(iterator.hasNext());
        PaginationOffsetNullableResponse res = iterator.next();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(serverLimit, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationOffsetNullableResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(serverLimit, nextRes.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationOffsetNullableResponse trailing = iterator.next();
        assertNotNull(trailing);
        assertEquals(200, trailing.statusCode());
        assertEquals(0, trailing.res().get().resultArray().size());

        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationLimitOffsetNilPageParams() throws Exception {
        recordTest("pagination-limit-offset-nil-page-params");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        long serverLimit = 20L;

        Iterator<PaginationLimitOffsetOptionalPageParamsResponse> iterator = s.pagination()
                .paginationLimitOffsetOptionalPageParams()
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationLimitOffsetOptionalPageParamsResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(serverLimit, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationLimitOffsetOptionalPageParamsResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(0, nextRes.res().get().resultArray().size());

        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationLimitOffsetZeroPageParams() throws Exception {
        recordTest("pagination-limit-offset-zero-page-params");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        long serverLimit = 20L;

        Iterator<PaginationLimitOffsetOptionalPageParamsResponse> iterator = s.pagination()
                .paginationLimitOffsetOptionalPageParams()
                .page(0L)
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationLimitOffsetOptionalPageParamsResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(serverLimit, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationLimitOffsetOptionalPageParamsResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(serverLimit, nextRes.res().get().resultArray().size());
    }

    @Test
    void paginationLimitOffsetNilOffsetParams() throws Exception {
        recordTest("pagination-limit-offset-nil-offset-params");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        long defaultLimit = 20L;

        Iterator<PaginationLimitOffsetOffsetParamsResponse> iterator = s.pagination()
                .paginationLimitOffsetOffsetParams()
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationLimitOffsetOffsetParamsResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(defaultLimit, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationLimitOffsetOffsetParamsResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(20L - defaultLimit, nextRes.res().get().resultArray().size());

        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationLimitOffsetDefaultOffsetBody() throws Exception {
        recordTest("pagination-limit-offset-default-offset-body");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        // Sending an empty body should cause the generator to fill in the schema defaults
        // (limit=15, offset=10). The test service returns ceil(total/limit) pages, so the
        // first page comes back with 15 items starting at offset 10 (i.e., 10 items).
        PaginationLimitOffsetDefaultOffsetBodyResponse res = s.pagination()
                .paginationLimitOffsetDefaultOffsetBody()
                .request(LimitOffsetConfigWithDefaults.builder().build())
                .call();

        assertEquals(200, res.statusCode());
        // 20 total items, offset 10, limit 15 → 10 items returned on first page
        assertEquals(10, res.res().get().resultArray().size());
    }

    @Test
    void paginationCursorParams() throws Exception {
        recordTest("pagination-cursor-params");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        long cursor = -1;

        Iterator<PaginationCursorParamsResponse> iterator = s.pagination().paginationCursorParams()
                .cursor(cursor)
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationCursorParamsResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(15, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationCursorParamsResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(5, nextRes.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationCursorParamsResponse penultimateRes = iterator.next();
        assertNotNull(penultimateRes);
        assertEquals(200, penultimateRes.statusCode());
        assertEquals(0, penultimateRes.res().get().resultArray().size());

        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationCursorParamsUnwrapped() throws Exception {
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        long cursor = -1;

        Stream<Long> results = s.pagination().paginationCursorParams()
                .cursor(cursor)
                .callAsStreamUnwrapped();
        assertEquals(20, results.count());
    }

    @Test
    void paginationCursorBody() throws Exception {
        recordTest("pagination-cursor-body");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        Long cursor = -1l;

        PaginationCursorBodyRequestBody request = new PaginationCursorBodyRequestBody(cursor);

        Iterator<PaginationCursorBodyResponse> iterator = s.pagination().paginationCursorBody()
                .request(request)
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationCursorBodyResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(15, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationCursorBodyResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(5, nextRes.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationCursorBodyResponse penultimateRes = iterator.next();
        assertNotNull(penultimateRes);
        assertEquals(200, penultimateRes.statusCode());
        assertEquals(0, penultimateRes.res().get().resultArray().size());

        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationCursorBodyUnwrapped() throws Exception {
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        long cursor = -1;

        PaginationCursorBodyRequestBody request = new PaginationCursorBodyRequestBody(cursor);

        Stream<Long> results = s.pagination().paginationCursorBody()
                .request(request)
                .callAsStreamUnwrapped();
        assertEquals(20, results.count());
    }

    @Test
    void paginationCursorNonNumeric() throws Exception {
        recordTest("pagination-cursor-non-numeric");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        Iterator<PaginationCursorNonNumericResponse> iterator = s.pagination().paginationCursorNonNumeric().callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationCursorNonNumericResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(15, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationCursorNonNumericResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(5, nextRes.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationCursorNonNumericResponse penultimateRes = iterator.next();
        assertNotNull(penultimateRes);
        assertEquals(200, penultimateRes.statusCode());
        assertEquals(0, penultimateRes.res().get().resultArray().size());

        assertFalse(iterator.hasNext());
    }

    @Disabled
    @Test
    void paginationCursorNonNumericNullable() throws Exception {
        // recordTest("pagination-cursor-non-numeric-nullable");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        Iterator<PaginationCursorNonNumericNullableResponse> iterator = s.pagination()
                .paginationCursorNonNumericNullable()
                .cursor("2")
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationCursorNonNumericNullableResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(15, res.res().get().resultArray().size());
        assertEquals("17", res.res().get().cursor().get());

        assertTrue(iterator.hasNext());
        PaginationCursorNonNumericNullableResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(2, nextRes.res().get().resultArray().size());
        assertNull(nextRes.res().get().cursor());

        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationCursorNonNumericUnwrapped() throws Exception {
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        Stream<String> results = s.pagination().paginationCursorNonNumeric().callAsStreamUnwrapped();
        assertEquals(20, results.count());
    }

    @Test
    void paginationCursorBodyAsStream() throws Exception {
        recordTest("pagination-cursor-body");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        long cursor = -1l;

        long pagesCount = s.pagination().paginationCursorBody()
                .request(new PaginationCursorBodyRequestBody(cursor)).callAsStream()
                .flatMap(x -> x.res().stream()).count();
        assertEquals(3, pagesCount);

        long itemsCount = s.pagination().paginationCursorBody()
                .request(new PaginationCursorBodyRequestBody(cursor))
                .callAsStreamUnwrapped()
                .count();
        assertEquals(20, itemsCount);
    }

    @Test
    void paginationCursorResponseEnvelopeAsStream() throws Exception {
        recordTest("pagination-cursor-response-envelope");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        List<PaginationCursorResponseEnvelopeRes> pages = s.pagination().paginationCursorResponseEnvelope()
                .cursor("-1")
                .callAsStream()
                .flatMap(x -> x.res().stream())
                .collect(Collectors.toList());

        long pageCount = pages.size();
        long itemsCount = pages.stream().mapToLong(x -> x.resultArray().size()).sum();

        assertEquals(3, pageCount);
        assertEquals(20, itemsCount);
    }

    @Test
    void paginationCursorNonNumericWithLimit() throws Exception {
        recordTest("pagination-cursor-non-numeric-with-limit");
        // TODO: add this to test server side instead of mocking
        PaginationCursorNonNumericWithLimitRes o = PaginationCursorNonNumericWithLimitRes.builder()
                .resultArray(List.of("a", "b", "c", "d", "e"))
                .cursor("abc")
                .numPages(100)
                .build();
        String json = Utils.mapper().writeValueAsString(o);

        HTTPClient client = new HTTPClient() {
            @Override
            public HttpResponse<InputStream> send(HttpRequest request) {
                return CommonHelpers.createJsonResponse(request, 200, json);
            }
        };
        PaginationRecordingClient recordingClient = new PaginationRecordingClient(client);

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).client(recordingClient).build();

        PaginationCursorNonNumericWithLimitResponse res = s.pagination()
                .paginationCursorNonNumericWithLimit()
                .limit(5L)
                .call();

        assertEquals(200, res.statusCode());
        assertEquals(5, res.res().get().resultArray().size());

        // just want to know that the limit is passed as a query parameter
        assertTrue(recordingClient.requests().get(0).uri().toString().contains("limit=5"));
    }

    @Test
    void paginationWithRetries() throws Exception {
        recordTest("pagination-with-retries");
        SpeakeasyHTTPClient client = new SpeakeasyHTTPClient();
        PaginationRecordingClient recordingClient = new PaginationRecordingClient(client);

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).client(recordingClient).build();

        Iterator<PaginationWithRetriesResponse> iterator = s.pagination().paginationWithRetries()
                .requestId(UUID.randomUUID().toString())
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationWithRetriesResponse res = iterator.next();

        int count = 0;

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(15, res.res().get().resultArray().size());
        count += res.res().get().resultArray().size();

        assertTrue(iterator.hasNext());
        PaginationWithRetriesResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(5, nextRes.res().get().resultArray().size());
        count += nextRes.res().get().resultArray().size();

        assertTrue(iterator.hasNext());
        PaginationWithRetriesResponse penultimateRes = iterator.next();
        assertNotNull(penultimateRes);
        assertEquals(200, penultimateRes.statusCode());
        assertEquals(0, penultimateRes.res().get().resultArray().size());

        assertFalse(iterator.hasNext());

        assertEquals(20, count);

        var expected = List.of(
                "503:GET:/pagination/cursor_non_numeric",
                "503:GET:/pagination/cursor_non_numeric",
                "503:GET:/pagination/cursor_non_numeric",
                "200:GET:/pagination/cursor_non_numeric",
                "200:GET:/pagination/cursor_non_numeric",
                "200:GET:/pagination/cursor_non_numeric");

        assertEquals(expected, recordingClient.log());
    }

    @Test
    @Timeout(value = 30, unit = TimeUnit.SECONDS)
    void paginationWithRetriesUnwrapped() throws Exception {
        SpeakeasyHTTPClient client = new SpeakeasyHTTPClient();
        PaginationRecordingClient recordingClient = new PaginationRecordingClient(client);

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).client(recordingClient).build();

        Stream<String> results = s.pagination().paginationWithRetries()
                .requestId(UUID.randomUUID().toString())
                .callAsStreamUnwrapped();

        assertEquals(20, results.count());

        var expected = List.of(
                "503:GET:/pagination/cursor_non_numeric",
                "503:GET:/pagination/cursor_non_numeric",
                "503:GET:/pagination/cursor_non_numeric",
                "200:GET:/pagination/cursor_non_numeric",
                "200:GET:/pagination/cursor_non_numeric",
                "200:GET:/pagination/cursor_non_numeric");

        assertEquals(expected, recordingClient.log());
    }

    @Test
    void paginationCursorNullableLimit() throws Exception {
        recordTest("pagination-cursor-nullable-limit");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        int defaultLimit = 10;

        Iterator<PaginationCursorNullableLimitResponse> iterator = s.pagination().paginationCursorNullableLimit()
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationCursorNullableLimitResponse res = iterator.next();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(defaultLimit, res.paginationCursorNullableLimitNextCursor().get().results().size());

        assertTrue(iterator.hasNext());
        PaginationCursorNullableLimitResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(defaultLimit, nextRes.paginationCursorNullableLimitNextCursor().get().results().size());

        // Continue until exhausted
        while (iterator.hasNext()) {
            iterator.next();
        }

        assertFalse(iterator.hasNext());
    }

    private static final class PaginationRecordingClient implements HTTPClient {
        private final HTTPClient client;
        private final List<String> log = new ArrayList<>();
        private final List<HttpRequest> requests = new ArrayList<>();

        PaginationRecordingClient(HTTPClient client) {
            this.client = client;
        }

        @Override
        public HttpResponse<InputStream> send(HttpRequest request)
                throws IOException, InterruptedException, URISyntaxException {
            requests.add(request);
            HttpResponse<InputStream> response = client.send(request);
            this.log.add(response.statusCode() + ":" + request.method() + ":" + request.uri().getPath());
            return response;
        }

        List<String> log() {
            return log;
        }

        List<HttpRequest> requests() {
            return requests;
        }
    }
    
    @Test
    void paginationCursorNonNumericEmptyString() throws Exception {
        recordTest("pagination-cursor-non-numeric-empty-string");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        Iterator<PaginationCursorNonNumericEmptyStringResponse> iterator = s.pagination()
                .paginationCursorNonNumericEmptyString()
                .cursor("2")
                .endCursor("")
                .callAsIterable().iterator();
        assertTrue(iterator.hasNext());
        PaginationCursorNonNumericEmptyStringResponse res = iterator.next();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(15, res.res().get().resultArray().size());
        assertEquals("17", res.res().get().cursor().get());
        assertTrue(iterator.hasNext());
        PaginationCursorNonNumericEmptyStringResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(2, nextRes.res().get().resultArray().size());
        assertEquals("", nextRes.res().get().cursor().get());
        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationURLParams() throws Exception {
        recordTest("pagination-url");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        Iterator<PaginationURLParamsResponse> iterator = s.pagination().paginationURLParams()
                .attempts(3)
                .callAsIterable().iterator();

        // First page: 9 results (attempts=3 → 3*3=9)
        assertTrue(iterator.hasNext());
        PaginationURLParamsResponse res = iterator.next();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(9, res.res().get().resultArray().size());

        // Second page: 6 results (attempts=2 → 2*3=6)
        assertTrue(iterator.hasNext());
        PaginationURLParamsResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(6, nextRes.res().get().resultArray().size());

        // Third page: 3 results (attempts=1 → 1*3=3), no next URL
        assertTrue(iterator.hasNext());
        PaginationURLParamsResponse lastRes = iterator.next();
        assertNotNull(lastRes);
        assertEquals(200, lastRes.statusCode());
        assertEquals(3, lastRes.res().get().resultArray().size());

        // No more pages
        assertFalse(iterator.hasNext());
    }

    @Test
    void paginationURLParamsReferencePath() throws Exception {
        recordTest("pagination-url");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        Iterator<PaginationURLParamsResponse> iterator = s.pagination().paginationURLParams()
                .attempts(3)
                .isReferencePath("true")
                .callAsIterable().iterator();

        // Same pagination flow but with relative URL paths
        assertTrue(iterator.hasNext());
        PaginationURLParamsResponse res = iterator.next();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(9, res.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationURLParamsResponse nextRes = iterator.next();
        assertNotNull(nextRes);
        assertEquals(200, nextRes.statusCode());
        assertEquals(6, nextRes.res().get().resultArray().size());

        assertTrue(iterator.hasNext());
        PaginationURLParamsResponse lastRes = iterator.next();
        assertNotNull(lastRes);
        assertEquals(200, lastRes.statusCode());
        assertEquals(3, lastRes.res().get().resultArray().size());

        assertFalse(iterator.hasNext());
    }
}
