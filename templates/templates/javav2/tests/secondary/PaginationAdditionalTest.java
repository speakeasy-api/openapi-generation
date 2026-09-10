package org.openapis.secondary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.openapis.secondary.openapi.CommonHelpers.recordTest;

import java.util.stream.Stream;

import org.junit.jupiter.api.Test;

public class PaginationAdditionalTest {

    // Cursor pagination whose `results` array is required but nullable: the
    // terminating page returns `results: null`. callAsStreamUnwrapped flattens
    // every fetched page's results into a single stream, so it must null-guard
    // the array before iterating it. Under the "raw" getter style the getter
    // returns a plain, possibly-null list, so an unguarded stream over the null
    // terminating page throws instead of yielding nothing.
    @Test
    void paginationCursorNullableResultsUnwrapped() throws Exception {
        recordTest("pagination-cursor-nullable-results");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        Stream<String> results = s.pagination()
                .paginationCursorNullableResults()
                .nullTerminal(true)
                .callAsStreamUnwrapped();

        assertEquals(20, results.count());
    }
}
