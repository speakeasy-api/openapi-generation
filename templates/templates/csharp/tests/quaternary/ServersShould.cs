using System;
using System.Net.Http;
using System.Threading.Tasks;
using System.Collections.Generic;
using Xunit;
using Company.Product.Feature.Subnamespace;

public class ServersShould
{
    [Fact]
    public async Task SelectGlobalServerByNameDefault()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-default");

        var sdk = new SDK(apiKeyAuth: "token", serverUrl: CommonHelpers.HttpBinUrl);

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
    }

    [Fact]
    public async Task SelectGlobalServerByNameDefaultUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-default-using-builder");

        var sdk = SDK.Builder().WithApiKeyAuth("token").WithServerUrl(CommonHelpers.HttpBinUrl).Build();

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
    }

    [Fact]
    public async Task SelectGlobalServerByNameValid()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-valid");

        var sdk = new SDK(apiKeyAuth: "token", serverUrl: CommonHelpers.HttpBinUrl);

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
    }

    [Fact]
    public async Task SelectGlobalServerByNameValidUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-valid-using-builder");

        var sdk = SDK.Builder().WithApiKeyAuth("token").WithServerUrl(CommonHelpers.HttpBinUrl).Build();

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
    }

    [Fact]
    public void SelectGlobalServerByNameInValid()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-invalid");

        // N.A. - Ensured by Server Enum
    }

    [Fact]
    public async Task SelectGlobalServerByNameBroken()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-broken");

        var sdk = new SDK(apiKeyAuth: "token", server: SDKConfig.Server.Broken);

        await Assert.ThrowsAsync<HttpRequestException>(
            async () => await sdk.Servers.SelectGlobalServerAsync()
        );
    }

    [Fact]
    public async Task SelectGlobalServerByNameBrokenUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-broken-using-builder");

        var sdk = SDK.Builder().WithApiKeyAuth("token").WithServer(SDKConfig.Server.Broken).Build();

        await Assert.ThrowsAsync<HttpRequestException>(
            async () => await sdk.Servers.SelectGlobalServerAsync()
        );
    }

    [Fact]
    public async Task SelectGlobalServerByNameWithTemplatesDefaults()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-with-templates-defaults");

        var sdk = new SDK(apiKeyAuth: "token", server: SDKConfig.Server.Templated, hostname: "localhost", port: CommonHelpers.HttpBinPort);

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
    }

    [Fact]
    public async Task SelectGlobalServerByNameWithTemplatesDefaultsUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-with-templates-defaults-using-builder");

        var sdk = SDK.Builder().WithApiKeyAuth("token").WithServer(SDKConfig.Server.Templated).WithHostname("localhost").WithPort(CommonHelpers.HttpBinPort).Build();

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
    }

    [Fact]
    public async Task SelectGlobalServerByNameWithTemplatesValid()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-with-templates-valid");

        var sdk = new SDK(
            apiKeyAuth: "token",
            server: SDKConfig.Server.Templated,
            hostname: "127.0.0.1",
            port: CommonHelpers.HttpBinPort
        );

        Assert.Equal($"http://127.0.0.1:{CommonHelpers.HttpBinPort}", sdk.SDKConfiguration.GetTemplatedServerUrl());

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
    }

    [Fact]
    public async Task SelectGlobalServerByNameWithTemplatesValidUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-with-templates-valid-using-builder");

        var sdk = SDK.Builder()
            .WithApiKeyAuth("token")
            .WithServer(SDKConfig.Server.Templated)
            .WithHostname("127.0.0.1")
            .WithPort(CommonHelpers.HttpBinPort)
            .Build();

        Assert.Equal($"http://127.0.0.1:{CommonHelpers.HttpBinPort}", sdk.SDKConfiguration.GetTemplatedServerUrl());

        var res = await sdk.Servers.SelectGlobalServerAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);
    }

    [Fact]
    public async Task SelectGlobalServerByNameWithTemplatesBroken()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-with-templates-broken");

        var sdk = new SDK(
            apiKeyAuth: "token",
            server: SDKConfig.Server.Templated,
            hostname: "broken",
            port: "12345"
        );

        Assert.Equal("http://broken:12345", sdk.SDKConfiguration.GetTemplatedServerUrl());

        await Assert.ThrowsAsync<HttpRequestException>(
            async () => await sdk.Servers.SelectGlobalServerAsync()
        );
    }

    [Fact]
    public async Task SelectGlobalServerByNameWithTemplatesBrokenUsingBuilder()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-with-templates-broken-using-builder");

        var sdk = SDK.Builder()
            .WithApiKeyAuth("token")
            .WithServer(SDKConfig.Server.Templated)
            .WithHostname("broken")
            .WithPort("12345")
            .Build();

        Assert.Equal("http://broken:12345", sdk.SDKConfiguration.GetTemplatedServerUrl());

        await Assert.ThrowsAsync<HttpRequestException>(
            async () => await sdk.Servers.SelectGlobalServerAsync()
        );
    }
}
