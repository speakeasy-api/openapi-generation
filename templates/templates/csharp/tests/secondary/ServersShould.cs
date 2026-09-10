using System;
using System.Net.Http;
using System.Threading.Tasks;
using System.Collections.Generic;
using Xunit;
using HoistedSecurity;
using HoistedSecurity.Utils;

public class ServersShould
{
    [Fact]
    public async Task SelectGlobalServerByNameDefault()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-default");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl);

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
    }

    [Fact]
    public async Task SelectGlobalServerByNameDefaultUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-default-using-builder");

        var sdk = SDK.Builder().WithServerUrl(CommonHelpers.HttpBinUrl).Build();

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
    }

    [Fact]
    public async Task SelectGlobalServerByNameValid()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-valid");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl);

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
    }

    [Fact]
    public async Task SelectGlobalServerByNameValidUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-valid-using-builder");

        var sdk = SDK.Builder().WithServerUrl(CommonHelpers.HttpBinUrl).Build();

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
    }

    [Fact]
    public async Task SelectGlobalServerByNameBroken()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-broken");

        var sdk = new SDK(server: SDKConfig.Server.BrokenServer);

        await Assert.ThrowsAsync<HttpRequestException>(
            async () => await sdk.Servers.SelectGlobalServerAsync()
        );
    }

    [Fact]
    public async Task SelectGlobalServerByNameBrokenUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-broken-using-builder");

        var sdk = SDK.Builder().WithServer(SDKConfig.Server.BrokenServer).Build();

        await Assert.ThrowsAsync<HttpRequestException>(
            async () => await sdk.Servers.SelectGlobalServerAsync()
        );
    }
}
