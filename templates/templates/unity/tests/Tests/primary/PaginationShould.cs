using Openapi;
using Openapi.Models.Shared;
using System.Linq;
using Openapi.Models.Operations;
using NUnit.Framework;
using UnityEngine.TestTools;
using System.Collections;

public class PaginationShould
{
    [UnityTest]
    public IEnumerator PaginationLimitOffsetPageParams()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-page-params");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var serverLimit = 20;

            using (var res = await sdk.Pagination.PaginationLimitOffsetPageParamsAsync(page: 1))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(res.Res.ResultArray.Count(), serverLimit);

                using (var nextRes = await res.Next())
                {
                    Assert.AreEqual(200, nextRes.StatusCode);
                    Assert.NotNull(nextRes.Res);
                    Assert.AreEqual(nextRes.Res.ResultArray.Count(), 0);

                    using (var nullRes = await nextRes.Next())
                    {
                        Assert.Null(nullRes);
                    }
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator PaginationLimitOffsetPageBody()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-page-body");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var limit = 15;

            using (
                var res = await sdk.Pagination.PaginationLimitOffsetPageBodyAsync(
                    request: new LimitOffsetConfig { Page = 1, Limit = limit }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(res.Res.ResultArray.Count(), limit);

                using (var nextRes = await res.Next())
                {
                    Assert.AreEqual(200, nextRes.StatusCode);
                    Assert.NotNull(nextRes.Res);
                    Assert.True(
                        nextRes.Res.ResultArray.Count() < limit,
                        "result count is expected to be less than the limit"
                    );

                    using (var nullRes = await nextRes.Next())
                    {
                        Assert.Null(nullRes);
                    }
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator PaginationLimitOffsetDeepOutputsPageBody()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-deep-outputs-page-body");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var limit = 15;

            using (
                var res = await sdk.Pagination.PaginationLimitOffsetDeepOutputsPageBodyAsync(
                    request: new LimitOffsetConfig { Page = 1, Limit = limit }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(res.Res.ResultArray.Count(), limit);

                using (var nextRes = await res.Next())
                {
                    Assert.AreEqual(200, nextRes.StatusCode);
                    Assert.NotNull(nextRes.Res);
                    Assert.True(
                        nextRes.Res.ResultArray.Count() < limit,
                        "result count is expected to be less than the limit"
                    );

                    using (var nullRes = await nextRes.Next())
                    {
                        Assert.Null(nullRes);
                    }
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator PaginationLimitOffsetOffsetParams()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-offset-params");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var limit = 15;

            using (
                var res = await sdk.Pagination.PaginationLimitOffsetOffsetParamsAsync(
                    limit: limit,
                    offset: 0
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(res.Res.ResultArray.Count(), limit);

                using (var nextRes = await res.Next())
                {
                    Assert.AreEqual(200, nextRes.StatusCode);
                    Assert.NotNull(nextRes.Res);
                    Assert.True(
                        nextRes.Res.ResultArray.Count() < limit,
                        "result count is expected to be less than the limit"
                    );

                    using (var nullRes = await nextRes.Next())
                    {
                        Assert.Null(nullRes);
                    }
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator PaginationLimitOffsetOffsetBody()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-offset-body");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var limit = 15;

            using (
                var res = await sdk.Pagination.PaginationLimitOffsetOffsetBodyAsync(
                    request: new LimitOffsetConfig { Limit = limit, Offset = 0 }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(res.Res.ResultArray.Count(), limit);

                using (var nextRes = await res.Next())
                {
                    Assert.AreEqual(200, nextRes.StatusCode);
                    Assert.NotNull(nextRes.Res);
                    Assert.True(
                        nextRes.Res.ResultArray.Count() < limit,
                        "result count is expected to be less than the limit"
                    );

                    using (var nullRes = await nextRes.Next())
                    {
                        Assert.Null(nullRes);
                    }
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator PaginationCursorParams()
    {
        CommonHelpers.RecordTest("pagination-cursor-params");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var limit = 15;

            using (var res = await sdk.Pagination.PaginationCursorParamsAsync(cursor: -1))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(res.Res.ResultArray.Count(), limit);

                using (var nextRes = await res.Next())
                {
                    Assert.AreEqual(200, nextRes.StatusCode);
                    Assert.NotNull(nextRes.Res);
                    Assert.True(
                        nextRes.Res.ResultArray.Count() < limit,
                        "result count is expected to be less than the limit"
                    );

                    using (var penultimateRes = await nextRes.Next())
                    {
                        Assert.AreEqual(200, penultimateRes.StatusCode);
                        Assert.NotNull(penultimateRes.Res);
                        Assert.AreEqual(penultimateRes.Res.ResultArray.Count(), 0);

                        using (var nullRes = await penultimateRes.Next())
                        {
                            Assert.Null(nullRes);
                        }
                    }
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator PaginationCursorBody()
    {
        CommonHelpers.RecordTest("pagination-cursor-body");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var limit = 15;

            using (
                var res = await sdk.Pagination.PaginationCursorBodyAsync(
                    request: new PaginationCursorBodyRequestBody { Cursor = -1 }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(res.Res.ResultArray.Count(), limit);

                using (var nextRes = await res.Next())
                {
                    Assert.AreEqual(200, nextRes.StatusCode);
                    Assert.NotNull(nextRes.Res);
                    Assert.True(
                        nextRes.Res.ResultArray.Count() < limit,
                        "result count is expected to be less than the limit"
                    );

                    using (var penultimateRes = await nextRes.Next())
                    {
                        Assert.AreEqual(200, penultimateRes.StatusCode);
                        Assert.NotNull(penultimateRes.Res);
                        Assert.AreEqual(penultimateRes.Res.ResultArray.Count(), 0);

                        using (var nullRes = await penultimateRes.Next())
                        {
                            Assert.Null(nullRes);
                        }
                    }
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator PaginationCursorNonNumeric()
    {
        CommonHelpers.RecordTest("pagination-cursor-non-numeric");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.Pagination.PaginationCursorNonNumericAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(15, res.Res.ResultArray.Count());

                using (var nextRes = await res.Next())
                {
                    Assert.AreEqual(200, nextRes.StatusCode);
                    Assert.NotNull(nextRes.Res);
                    Assert.AreEqual(5, nextRes.Res.ResultArray.Count());

                    using (var penultimateRes = await nextRes.Next())
                    {
                        Assert.AreEqual(200, penultimateRes.StatusCode);
                        Assert.NotNull(penultimateRes.Res);
                        Assert.AreEqual(penultimateRes.Res.ResultArray.Count(), 0);

                        using (var nullRes = await penultimateRes.Next())
                        {
                            Assert.Null(nullRes);
                        }
                    }
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator PaginationCursorNonNumericNullable()
    {
        CommonHelpers.RecordTest("pagination-cursor-non-numeric-nullable");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.Pagination.PaginationCursorNonNumericAsync("2"))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(15, res.Res.ResultArray.Count());
                Assert.AreEqual("17", res.Res.Cursor);

                using (var nextRes = await res.Next())
                {
                    Assert.AreEqual(200, nextRes.StatusCode);
                    Assert.NotNull(nextRes.Res);
                    Assert.AreEqual(2, nextRes.Res.ResultArray.Count());
                    Assert.Null(nextRes.Res.Cursor);

                    using (var nullRes = await nextRes.Next())
                    {
                        Assert.Null(nullRes);
                    }
                }
            }
        });
    }
}
