using System;
using System.Net.Http;
using System.Threading.Tasks;
using System.Collections.Generic;
using Xunit;
using No_Security.API;

public class ServersShould
{
    [Fact]
    public async Task SelectGlobalServerByIdDefault()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-id-default");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
    }

       [Fact]
    public async Task SelectGlobalServerByIdDefaultUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-id-default-using-builder");

        var sdk = SDK.Builder().WithServerUrl(Helpers.HttpBinUrl).Build();

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
    }

    [Fact]
    public async Task SelectGlobalServerByIdValid()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-id-valid");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
    }

    [Fact]
    public async Task SelectGlobalServerByIdValidUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-id-valid-using-builder");

        var sdk = SDK.Builder().WithServerUrl(Helpers.HttpBinUrl).Build();

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
    }

    [Fact]
    public void SelectGlobalServerByIdInValid()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-id-invalid");

        var ex = Assert.Throws<ArgumentOutOfRangeException>(() => new SDK(serverIndex: 2));
        Assert.Contains("Invalid server index 2: must be between 0 (inclusive) and 2 (exclusive).", ex.Message);
    }

        [Fact]
    public void SelectGlobalServerByIdInValidUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-id-invalid-using-builder");

        var ex = Assert.Throws<ArgumentOutOfRangeException>(() => SDK.Builder().WithServerIndex(2).Build());
        Assert.Contains("Invalid server index 2: must be between 0 (inclusive) and 2 (exclusive).", ex.Message);
    }

    [Fact]
    public async Task SelectGlobalServerByIdBroken()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-id-broken");

        var sdk = new SDK(serverIndex: 1);

        await Assert.ThrowsAsync<HttpRequestException>(
            async () => await sdk.Servers.SelectGlobalServerAsync()
        );
    }

        [Fact]
    public async Task SelectGlobalServerByIdBrokenUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-id-broken-using-builder");

        var sdk = SDK.Builder().WithServerIndex(1).Build();

        await Assert.ThrowsAsync<HttpRequestException>(
            async () => await sdk.Servers.SelectGlobalServerAsync()
        );
    }
}
