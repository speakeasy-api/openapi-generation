using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Xunit;
using HoistedSecurity;

public class PaginationShould
{

    [Fact]
    public async Task PaginationLimitOffsetPageParamsFlat()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-page-params-flat");
        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl);

        var serverLimit = 20;

        var res = await sdk.Pagination.PaginationLimitOffsetPageParamsAsync(page: 1);

        Assert.NotNull(res.Result);
        Assert.Equal(res.Result.ResultArray.Count(), serverLimit);

        var nextRes = await res.Next();
        Assert.NotNull(nextRes.Result);
        Assert.Empty(nextRes.Result.ResultArray);

        var nullRes = await nextRes.Next();
        Assert.Null(nullRes);
    }

    [Fact]
    public async Task PaginationLimitOffsetUnionOutputPageParamsFlat()
    {
        CommonHelpers.RecordTest("pagination-limit-offset-union-output-page-params-flat");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl);
        var available = 20;
        var res = await sdk.Pagination.PaginationLimitOffsetUnionOutputPageParamsAsync(page: 1);

        Assert.NotNull(res.Result.ResultObject);
        Assert.Equal(res.Result.ResultObject.ResultArray.Count(), available);

        var nextRes = await res.Next();
        Assert.NotNull(nextRes.Result.ResultObject);
        Assert.Empty(nextRes.Result.ResultObject.ResultArray);

        var nullRes = await nextRes.Next();
        Assert.Null(nullRes);
    }

}
