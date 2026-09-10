#nullable enable
using Xunit;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Web;
using System.Threading;
using System.Threading.Tasks;
using Openapi;
using Openapi.Models.Operations;
using Openapi.Models.Shared;
using Openapi.Utils;


public class PaginationShould
{

    internal class PaginationLogEntry
    {
        public Uri RequestUri { get; set; } = default!;
        public string RequestBody { get; set; } = default!;
        public HttpStatusCode StatusCode { get; set; }
    }

    internal class PaginationRecorderClient : DefaultHttpClient
    {
        private List<PaginationLogEntry> log;

        public PaginationRecorderClient(List<PaginationLogEntry> log)
        {
            this.log = log ?? throw new ArgumentNullException(nameof(log));
        }

        public override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage httpRequest, CancellationToken? cancellationToken = null)
        {
            var requestBody = "";
            if (httpRequest.Content != null)
            {
                requestBody = await httpRequest.Content.ReadAsStringAsync();
            }

            var httpResponse = await base.SendAsync(httpRequest, cancellationToken);
            var statusCode = httpResponse.StatusCode;

            Assert.NotNull(httpRequest.RequestUri);
            log.Add(new PaginationLogEntry
            {
                RequestUri = httpRequest.RequestUri,
                RequestBody = requestBody,
                StatusCode = statusCode,
            });

            return httpResponse;
        }
    }

    [Fact]
    public async Task PaginationLimitOffsetPageParams()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-page-params");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var serverLimit = 20;

        var res = await sdk.Pagination.PaginationLimitOffsetPageParamsAsync(page: 1);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(res.Res.ResultArray.Count(), serverLimit);

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.Empty(nextRes.Res.ResultArray);

        var nullRes = await nextRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationLimitOffsetUnionOutputPageParams()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-union-output-page-params");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var available = 20;
        var res = await sdk.Pagination.PaginationLimitOffsetUnionOutputPageParamsAsync(page: 1);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.NotNull(res.Res.ResultObject);
        Assert.Equal(res.Res.ResultObject.ResultArray.Count(), available);

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.NotNull(nextRes.Res.ResultObject);
        Assert.Empty(nextRes.Res.ResultObject.ResultArray);

        var nullRes = await nextRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationLimitOffsetNilPageParams()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-nil-page-params");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var available = 20;
        var res = await sdk.Pagination.PaginationLimitOffsetOptionalPageParamsAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(available, res.Res.ResultArray.Count());

        res = await res.Next!();
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Empty(res.Res.ResultArray);

        res = await res.Next!();
        Assert.Null(res);
    }

    [Fact]
    public async Task PaginationLimitOffsetZeroPageParams()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-zero-page-params");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var available = 20;
        var res = await sdk.Pagination.PaginationLimitOffsetOptionalPageParamsAsync(page: 0);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(available, res.Res.ResultArray.Count());

        var queryParams = HttpUtility.ParseQueryString(res.HttpMeta.Request.RequestUri!.Query);
        Assert.Equal("0", queryParams.Get("page"));

        res = await res.Next!();
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(available, res.Res.ResultArray.Count());

        queryParams = HttpUtility.ParseQueryString(res.HttpMeta.Request.RequestUri!.Query);
        Assert.Equal("1", queryParams.Get("page"));
    }

    [Fact]
    public async Task PaginationLimitOffsetPageBody()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-page-body");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var limit = 15;

        var res = await sdk.Pagination.PaginationLimitOffsetPageBodyAsync(request: new LimitOffsetConfig{Page = 1, Limit = limit});

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(res.Res.ResultArray.Count(), limit);

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.True(nextRes.Res.ResultArray.Count() < limit, "result count is expected to be less than the limit");

        var nullRes = await nextRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationLimitOffsetPageBodyNullable()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-page-body-nullable");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // first request sends a null body; later pages materialize one with
        // the advanced page (wire sequence enforced by the test service)
        var res = await sdk.Pagination.PaginationLimitOffsetPageBodyNullableAsync(request: null);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(new List<long> { 0, 1, 2, 3, 4, 5, 6 }, res.Res.ResultArray);

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.Equal(new List<long> { 7, 8, 9, 10, 11, 12, 13 }, nextRes.Res.ResultArray);

        var lastRes = await nextRes.Next!();
        Assert.NotNull(lastRes);
        Assert.Equal(HttpStatusCode.OK, lastRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(lastRes.Res);
        Assert.Equal(new List<long> { 14, 15, 16, 17, 18, 19 }, lastRes.Res.ResultArray);

        var nullRes = await lastRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationLimitOffsetDeepOutputsPageBody()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-deep-outputs-page-body");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var limit = 15;

        var res = await sdk.Pagination.PaginationLimitOffsetDeepOutputsPageBodyAsync(request: new LimitOffsetConfig{Page = 1, Limit = limit});

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(res.Res.ResultArray.Count(), limit);

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.True(nextRes.Res.ResultArray.Count() < limit, "result count is expected to be less than the limit");

        var nullRes = await nextRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationOffsetNullable()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-nullable");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // Passing null for the required+nullable offset/limit forces the SDK to
        // serialise null values; the generated paginator must narrow them to a
        // non-null value type before computing the next offset and comparing the
        // result count against the limit, otherwise the arithmetic fails to compile.
        var serverLimit = 10;
        var res = await sdk.Pagination.PaginationOffsetNullableAsync(offset: null, limit: null);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(serverLimit, res.Res.ResultArray.Count());

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.NotNull(nextRes.Res);
        Assert.Equal(serverLimit, nextRes.Res.ResultArray.Count());

        var emptyRes = await nextRes.Next!();
        Assert.NotNull(emptyRes);
        Assert.NotNull(emptyRes.Res);
        Assert.Empty(emptyRes.Res.ResultArray);

        var nullRes = await emptyRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationLimitOffsetPageBodyOptionalNullable()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-page-body-optional-nullable");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var limit = 15;

        // page and limit are optional+nullable body fields, so they are emitted as
        // OptionalNullable<long?> (no arithmetic/comparison operators); the paginator
        // must unwrap them before computing the next page.
        var res = await sdk.Pagination.PaginationLimitOffsetPageBodyOptionalNullableAsync(
            request: new NullableLimitOffsetConfig { Page = 1, Limit = limit }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(limit, res.Res.ResultArray.Count());

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.NotNull(nextRes.Res);
        Assert.True(nextRes.Res.ResultArray.Count() < limit, "result count is expected to be less than the limit");

        var nullRes = await nextRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationLimitOffsetOffsetParams()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-offset-params");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var limit = 15;

        var res = await sdk.Pagination.PaginationLimitOffsetOffsetParamsAsync(limit: limit, offset: 0);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(res.Res.ResultArray.Count(), limit);

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.True(nextRes.Res.ResultArray.Count() < limit, "result count is expected to be less than the limit");

        var nullRes = await nextRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationLimitOffsetNilOffsetParams()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-nil-offset-params");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var available = 20;
        var defaultLimit = 20;

        var res = await sdk.Pagination.PaginationLimitOffsetOffsetParamsAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(res.Res.ResultArray.Count(), defaultLimit);

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.Equal(nextRes.Res.ResultArray.Count(), available - defaultLimit);

        var nullRes = await nextRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationLimitOffsetOffsetBody()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-offset-body");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var limit = 15;

        var res = await sdk.Pagination.PaginationLimitOffsetOffsetBodyAsync(request: new LimitOffsetConfig{Limit = limit, Offset = 0});

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(limit, res.Res.ResultArray.Count());

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.True(nextRes.Res.ResultArray.Count() < limit, "result count is expected to be less than the limit");

        var nullRes = await nextRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationLimitOffsetDefaultOffsetBody()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-default-offset-body");

        var available = 20;
        var defaultOffset = 10;
        var log = new List<PaginationLogEntry>{};
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, client: new PaginationRecorderClient(log));

        var res = await sdk.Pagination.PaginationLimitOffsetDefaultOffsetBodyAsync(
            new LimitOffsetConfigWithDefaults{}
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(res.Res.ResultArray.Count(), available-defaultOffset);

        Assert.Single(log);
        var requestBody = log[0].RequestBody;
        Assert.NotNull(requestBody);
        Assert.Contains("\"limit\":15", requestBody);
        Assert.Contains("\"offset\":10", requestBody);
    }

    [Fact]
    public async Task PaginationLimitOffsetDefaultOffsetParams()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-default-offset-params");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var res = await sdk.Pagination.PaginationLimitOffsetDefaultOffsetParamsAsync();
        var queryParams = HttpUtility.ParseQueryString(res.HttpMeta.Request.RequestUri!.Query);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("15", queryParams.Get("limit"));
        Assert.Equal("10", queryParams.Get("offset"));
    }

    [Fact]
    public async Task PaginationCursorParams()
    {
        CommonHelpers.RecordTest("pagination-cursor-params");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var limit = 15;

        var res = await sdk.Pagination.PaginationCursorParamsAsync(cursor: -1);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(res.Res.ResultArray.Count(), limit);

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.True(nextRes.Res.ResultArray.Count() < limit, "result count is expected to be less than the limit");

        var penultimateRes = await nextRes.Next!();
        Assert.NotNull(penultimateRes);
        Assert.Equal(HttpStatusCode.OK, penultimateRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(penultimateRes.Res);
        Assert.Empty(penultimateRes.Res.ResultArray);


        var nullRes = await penultimateRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationCursorBody()
    {
        CommonHelpers.RecordTest("pagination-cursor-body");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var limit = 15;

        var res = await sdk.Pagination.PaginationCursorBodyAsync(request: new PaginationCursorBodyRequestBody{Cursor = -1});

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(res.Res.ResultArray.Count(), limit);

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.True(nextRes.Res.ResultArray.Count() < limit, "result count is expected to be less than the limit");

        var penultimateRes = await nextRes.Next!();
        Assert.NotNull(penultimateRes);
        Assert.Equal(HttpStatusCode.OK, penultimateRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(penultimateRes.Res);
        Assert.Empty(penultimateRes.Res.ResultArray);


        var nullRes = await penultimateRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationCursorNonNumeric()
    {
        CommonHelpers.RecordTest("pagination-cursor-non-numeric");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Pagination.PaginationCursorNonNumericAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(15, res.Res.ResultArray.Count());

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.Equal(5, nextRes.Res.ResultArray.Count());

        var penultimateRes = await nextRes.Next!();
        Assert.NotNull(penultimateRes);
        Assert.Equal(HttpStatusCode.OK, penultimateRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(penultimateRes.Res);
        Assert.Empty(penultimateRes.Res.ResultArray);

        var nullRes = await penultimateRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationCursorNonNumericNullable()
    {
        CommonHelpers.RecordTest("pagination-cursor-non-numeric-nullable");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Pagination.PaginationCursorNonNumericNullableAsync("2");

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(15, res.Res.ResultArray.Count());
        Assert.Equal("17", res.Res.Cursor);

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.Equal(2, nextRes.Res.ResultArray.Count());
        Assert.True(nextRes.Res.Cursor.IsNull);

        var nullRes = await nextRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationCursorNonNumericEmptyString()
    {
        CommonHelpers.RecordTest("pagination-cursor-non-numeric-empty-string");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Pagination.PaginationCursorNonNumericEmptyStringAsync("", "2");

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(15, res.Res.ResultArray.Count());
        Assert.Equal("17", res.Res.Cursor);

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.Equal(2, nextRes.Res.ResultArray.Count());
        Assert.Equal("", nextRes.Res.Cursor);

        var nullRes = await nextRes.Next!();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationWithRetriesAsync()
    {
        CommonHelpers.RecordTest("pagination-with-retries");

        var available = 20;
        var count = 0;
        var log = new List<PaginationLogEntry>{};
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, client: new PaginationRecorderClient(log));

        var res = await sdk.Pagination.PaginationWithRetriesAsync(
          requestId: System.Guid.NewGuid().ToString(),
          faultSettings: "{\"error_code\": 503, \"error_count\": 3}"
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        count += res.Res.ResultArray.Count();

        res = await res.Next!();
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        count += res.Res.ResultArray.Count();

        res = await res.Next!();
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Empty(res.Res.ResultArray);

        res = await res.Next!();
        Assert.Null(res);

        Assert.Equal(available, count);

        Assert.Equal(6, log.Count);
        for (var i = 0; i < 3; i++)
        {
            Assert.Equal("", log[i].RequestUri.Query);
            Assert.Equal(HttpStatusCode.ServiceUnavailable, log[i].StatusCode);
        }
        Assert.Equal(HttpStatusCode.OK, log[3].StatusCode);
        Assert.Equal("", log[3].RequestUri.Query);
        Assert.Equal(HttpStatusCode.OK, log[4].StatusCode);
        Assert.Equal("?cursor=14", log[4].RequestUri.Query);
        Assert.Equal(HttpStatusCode.OK, log[5].StatusCode);
        Assert.Equal("?cursor=19", log[5].RequestUri.Query);
    }

    [Fact]
    public async Task PaginationBodyWrappedRequest()
    {
        CommonHelpers.RecordTest("pagination-body-wrapped-request");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var available = 20;
        var limit = 15;

        var res = await sdk.Pagination.PaginationBodyWrappedRequestAsync(
          request: new PaginationBodyWrappedRequestRequest
          {
              LimitOffsetConfig = new LimitOffsetConfig
              {
                  Page = 1,
                  Limit = limit
              }
          }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(limit, res.Res.ResultArray.Count());

        res = await res.Next!();
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(available-limit, res.Res.ResultArray.Count());

        res = await res.Next!();
        Assert.Null(res);
    }

    [Fact]
    public async Task PaginationParamsWrappedRequest()
    {
        CommonHelpers.RecordTest("pagination-params-wrapped-request");

        var log = new List<PaginationLogEntry>{};
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, client: new PaginationRecorderClient(log));

        var available = 20;
        var limit = 15;
        var offset = 1;

        var res = await sdk.Pagination.PaginationParamsWrappedRequestAsync(
            request: new PaginationParamsWrappedRequestRequest
            {
                Limit = limit,
                Offset = offset,
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(limit, res.Res.ResultArray.Count());

        res = await res.Next!();
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(available-offset-limit, res.Res.ResultArray.Count());

        res = await res.Next!();
        Assert.Null(res);
    }

    [Fact]
    public async Task PaginationBodyFlattenedWithSecurity()
    {
        CommonHelpers.RecordTest("pagination-body-flattened-with-security");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var available = 20;
        var limit = 15;

        var res = await sdk.Pagination.PaginationBodyFlattenedWithSecurityAsync(
            new PaginationBodyFlattenedWithSecuritySecurity
            {
                PaginationAuth = "test"
            },
            limit,
            0
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(limit, res.Res.ResultArray.Count());

        res = await res.Next!();
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(available-limit, res.Res.ResultArray.Count());

        res = await res.Next!();
        Assert.Null(res);
    }

    [Fact]
    public async Task PaginationBodyFlattenedOptionalSecurity()
    {
        CommonHelpers.RecordTest("pagination-body-flattened-optional-security");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var available = 20;
        var limit = 15;

        var res = await sdk.Pagination.PaginationBodyFlattenedOptionalSecurityAsync(
            limit,
            0,
            new PaginationBodyFlattenedOptionalSecuritySecurity
            {
                PaginationAuth = "test"
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(limit, res.Res.ResultArray.Count());

        res = await res.Next!();
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(available-limit, res.Res.ResultArray.Count());

        res = await res.Next!();
        Assert.Null(res);
    }

    [Fact]
    public async Task PaginationAmbiguousInput()
    {
        CommonHelpers.RecordTest("pagination-ambiguous-input");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var available = 20;
        var limit = 15;
        var cursor = available - limit - 1;

        var res = await sdk.Pagination.PaginationAmbiguousInputAsync(
            requestBody: new PaginationAmbiguousInputRequestBody
            {
                Cursor = cursor,
            },
            cursor: 100
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(limit, res.Res.ResultArray.Count());

        res = await res.Next!();
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Empty(res.Res.ResultArray);

        res = await res.Next!();
        Assert.Null(res);
    }

    [Fact]
    public async Task PaginationWrappedOptionalBody()
    {
        CommonHelpers.RecordTest("pagination-wrapped-optional-body");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var available = 20;

        var res = await sdk.Pagination.PaginationWrappedOptionalBodyAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(available, res.Res.ResultArray.Count());

        res = await res.Next!();
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Empty(res.Res.ResultArray);

        res = await res.Next!();
        Assert.Null(res);
    }

    [Fact]
    public async Task PaginationEncapsulatedParameter()
    {
        CommonHelpers.RecordTest("pagination-encapsulated-parameter");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var available = 20;
        var limit = 15;
        var cursor = available - limit - 1;

        var res = await sdk.Pagination.PaginationEncapsulatedParameterAsync(
            request: new PaginationEncapsulatedParameterRequest
            {
                Cursor = cursor,
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(limit, res.Res.ResultArray.Count());

        res = await res.Next!();
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Empty(res.Res.ResultArray);

        res = await res.Next!();
        Assert.Null(res);
    }

    [Fact]
    public async Task PaginationCursorNullableLimit()
    {
        CommonHelpers.RecordTest("pagination-cursor-nullable-limit");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var defaultLimit = 10;

        var res = await sdk.Pagination.PaginationCursorNullableLimitAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.PaginationCursorNullableLimitNextCursor);
        Assert.Equal(defaultLimit, res.PaginationCursorNullableLimitNextCursor.Results.Count());

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.PaginationCursorNullableLimitNextCursor);
        Assert.Equal(defaultLimit, nextRes.PaginationCursorNullableLimitNextCursor.Results.Count());

        // Paginate until exhausted
        var lastRes = nextRes;
        while (true)
        {
            var n = await lastRes.Next!();
            if (n == null) break;
            lastRes = n;
        }

        Assert.NotNull(lastRes.PaginationCursorNullableLimitNextCursor);
    }

    [Fact]
    public async Task PaginationURLParams()
    {
        CommonHelpers.RecordTest("pagination-url");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Pagination.PaginationURLParamsAsync(attempts: 3);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(9, res.Res.ResultArray.Count());

        var nextRes = await res.Next!();
        Assert.NotNull(nextRes);
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes.Res);
        Assert.Equal(6, nextRes.Res.ResultArray.Count());

        var penultimateRes = await nextRes.Next!();
        Assert.NotNull(penultimateRes);
        Assert.Equal(HttpStatusCode.OK, penultimateRes.HttpMeta.Response.StatusCode);
        Assert.NotNull(penultimateRes.Res);
        Assert.Equal(3, penultimateRes.Res.ResultArray.Count());

        var nullRes = await penultimateRes.Next!();
        Assert.Null(nullRes);

        // Test with is-reference-path which uses a relative URL for the next link
        var res2 = await sdk.Pagination.PaginationURLParamsAsync(attempts: 3, isReferencePath: "true");

        Assert.Equal(HttpStatusCode.OK, res2.HttpMeta.Response.StatusCode);
        Assert.NotNull(res2.Res);
        Assert.Equal(9, res2.Res.ResultArray.Count());

        var nextRes2 = await res2.Next!();
        Assert.NotNull(nextRes2);
        Assert.Equal(HttpStatusCode.OK, nextRes2.HttpMeta.Response.StatusCode);
        Assert.NotNull(nextRes2.Res);
        Assert.Equal(6, nextRes2.Res.ResultArray.Count());

        var penultimateRes2 = await nextRes2.Next!();
        Assert.NotNull(penultimateRes2);
        Assert.Equal(HttpStatusCode.OK, penultimateRes2.HttpMeta.Response.StatusCode);
        Assert.NotNull(penultimateRes2.Res);
        Assert.Equal(3, penultimateRes2.Res.ResultArray.Count());

        var nullRes2 = await penultimateRes2.Next!();
        Assert.Null(nullRes2);
    }
}
