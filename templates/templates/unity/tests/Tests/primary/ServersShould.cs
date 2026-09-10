#nullable enable
using System;
using NUnit.Framework;
using UnityEngine.TestTools;
using Openapi;
using Openapi.Utils;
using System.Collections;
using System.Collections.Generic;

public class ServersShould
{
    [UnityTest]
    public IEnumerator SelectGlobalServerValid()
    {
        CommonHelpers.RecordTest("servers-select-global-server-valid");

        string url = $"http://127.0.0.1:{Helpers.HttpBinPort}";
        var sdk = new SDK(serverUrl: url);

        Assert.AreEqual(url, sdk.SDKConfiguration.GetTemplatedServerDetails());

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(serverUrl: SDKConfig.ServerList[0]);

            using (var res = await sdk.Servers.SelectGlobalServerAsync())
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator SelectGlobalServerBroken()
    {
        CommonHelpers.RecordTest("servers-select-global-server-broken");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(serverUrl: SDKConfig.ServerList[1]);

            Exception? exception = null;

            try
            {
                using (var res = await sdk.Servers.SelectGlobalServerAsync()) { }
            }
            catch (Exception e)
            {
                exception = e;
            }

            Assert.NotNull(exception);
        });
    }

    [UnityTest]
    public IEnumerator SelectServerWithIDDefault()
    {
        CommonHelpers.RecordTest("servers-select-server-with-id-default");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.Servers.SelectServerWithIDAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator SelectServerWithIDValid()
    {
        CommonHelpers.RecordTest("servers-select-server-with-id-valid");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(serverUrl: SDKConfig.ServerList[1]);  // broken server overridden by operation

            using (
                var res = await sdk.Servers.SelectServerWithIDAsync(
                    serverUrl: Servers.SelectServerWithIDServerMap[
                        Servers.SelectServerWithIDServers.Valid
                    ]
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator SelectServerWithIDBroken()
    {
        CommonHelpers.RecordTest("servers-select-server-with-id-broken");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            Exception? exception = null;

            try
            {
                using (
                    var res = await sdk.Servers.SelectServerWithIDAsync(
                        serverUrl: Servers.SelectServerWithIDServerMap[
                            Servers.SelectServerWithIDServers.Broken
                        ]
                    )
                ) { }
            }
            catch (Exception e)
            {
                exception = e;
            }

            Assert.NotNull(exception);
        });
    }

    [UnityTest]
    public IEnumerator ServerWithTemplatesGlobal()
    {
        CommonHelpers.RecordTest("servers-server-with-templates-global");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(serverIndex: 2, hostname: "localhost", port: Helpers.HttpBinPort);

            using (var res = await sdk.Servers.ServerWithTemplatesGlobalAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator ServerWithTemplatesGlobalDefaults()
    {
        CommonHelpers.RecordTest("servers-server-with-templates-global-defaults");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(serverIndex: 2);

            Assert.AreEqual(2, sdk.SDKConfiguration.serverIndex);

            using (var res = await sdk.Servers.ServerWithTemplatesGlobalAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
            }

        });
    }

    [UnityTest]
    public IEnumerator ServerWithTemplatesGlobalEnum()
    {
        CommonHelpers.RecordTest("servers-server-with-templates-global-enum");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(serverIndex: 2, something: Openapi.ServerSomething.SomethingElseAgain);

            using (var res = await sdk.Servers.ServerWithTemplatesGlobalAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator ServerWithTemplates()
    {
        CommonHelpers.RecordTest("servers-server-with-templates");
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.Servers.ServerWithTemplatesAsync(
                serverUrl: Utilities.TemplateUrl(
                    Servers.ServerWithTemplatesServerList[0],
                    new Dictionary<string, string>(){
                        {"protocol", "http"},
                        {"hostname", "localhost"},
                        {"port", Helpers.HttpBinPort},
                    }
                )
            ))
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator ServerWithTemplatesDefaults()
    {
        CommonHelpers.RecordTest("servers-server-with-templates-defaults");
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.Servers.ServerWithTemplatesAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator ServerByIDWithTemplates()
    {
        CommonHelpers.RecordTest("servers-server-by-id-with-templates");
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.Servers.ServersByIDWithTemplatesAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator GlobalServerWithTemplatedProtocol()
    {
        CommonHelpers.RecordTest("servers-global-server-with-templated-protocol");
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(serverIndex: 4, protocol: "http", hostname: "localhost", port: Helpers.HttpBinPort);

            using (var res = await sdk.Servers.SelectGlobalServerAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator GlobalServerWithInvalidTemplatedProtocol()
    {
        CommonHelpers.RecordTest("servers-global-server-with-invalid-templated-protocol");
        yield return CommonHelpers.Await(async () =>
        {

            var sdk = new SDK(serverIndex: 4, protocol: "invalid", hostname: "localhost", port: Helpers.HttpBinPort);

            Exception? exception = null;

            try
            {
                await sdk.Servers.SelectGlobalServerAsync();
            }
            catch (Exception e)
            {
                exception = e;
            }
            Assert.NotNull(exception);
        });

    }
    [UnityTest]
    public IEnumerator ServerWithProtocolTemplate()
    {
        CommonHelpers.RecordTest("servers-server-with-protocol-template");
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.Servers.ServerWithProtocolTemplateAsync(
                serverUrl: Utilities.TemplateUrl(
                    Servers.ServerWithProtocolTemplateServerMap[Servers.ServerWithProtocolTemplateServers.Main],
                    new Dictionary<string, string>(){
                        {"protocol", "http"},
                        {"hostname", "localhost"},
                        {"port", Helpers.HttpBinPort},
                    }
                )
            ))
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator ServerWithInvalidProtocolTemplate()
    {
        CommonHelpers.RecordTest("servers-server-with-invalid-protocol-template");
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            Exception? exception = null;

            try
            {
                await sdk.Servers.ServerWithProtocolTemplateAsync(
                    serverUrl: Utilities.TemplateUrl(
                        Servers.ServerWithProtocolTemplateServerMap[Servers.ServerWithProtocolTemplateServers.Main],
                        new Dictionary<string, string>(){
                            {"protocol", "invalid"},
                            {"hostname", "localhost"},
                            {"port", Helpers.HttpBinPort},
                        }
                    )
                );
            }
            catch (Exception e)
            {
                exception = e;
            }
            Assert.NotNull(exception);
        });
    }

}
