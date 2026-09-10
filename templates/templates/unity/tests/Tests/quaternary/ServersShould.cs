using NUnit.Framework;
using UnityEngine.TestTools;
using System;
using System.Collections;
using System.Collections.Generic;
using Company.Product.Feature.Subnamespace;

public class ServersShould
{
    [UnityTest]
    public IEnumerator SelectGlobalServerByNameDefault()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-default");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(apiKeyAuth: "token");

            using (
                var res = await sdk.Servers.SelectGlobalServerAsync()
            )
                  {
                      Assert.NotNull(res);
                      Assert.AreEqual(200, res.StatusCode);
                  }
        });
    }

    [UnityTest]
    public IEnumerator SelectGlobalServerByNameValid()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-valid");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(apiKeyAuth: "token", server: SDKConfig.Server.Default);

            using (
                var res = await sdk.Servers.SelectGlobalServerAsync()
            )
                  {
                      Assert.NotNull(res);
                      Assert.AreEqual(200, res.StatusCode);
                  }
        });
    }

    [Test]
    public void SelectGlobalServerByNameInvalid()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-invalid");

        // N.A. - Ensured by Server Enum
    }

    [UnityTest]
    public IEnumerator SelectGlobalServerByNameBroken()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-broken");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(apiKeyAuth: "token", server: SDKConfig.Server.Broken);

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
    public IEnumerator SelectGlobalServerByNameWithTemplatesDefaults()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-with-templates-defaults");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(apiKeyAuth: "token", server: SDKConfig.Server.Templated);

            using (
                var res = await sdk.Servers.SelectGlobalServerAsync()
            )
                  {
                      Assert.NotNull(res);
                      Assert.AreEqual(200, res.StatusCode);
                  }
        });
    }

    [UnityTest]
    public IEnumerator SelectGlobalServerByNameWithTemplatesValid()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-with-templates-valid");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(
                apiKeyAuth: "token",
                server: SDKConfig.Server.Templated,
                hostname: "127.0.0.1",
                port: CommonHelpers.HttpBinPort
            );

            Assert.AreEqual($"http://127.0.0.1:{CommonHelpers.HttpBinPort}", sdk.SDKConfiguration.GetTemplatedServerDetails());

            using (
                var res = await sdk.Servers.SelectGlobalServerAsync()
            )
                  {
                      Assert.NotNull(res);
                      Assert.AreEqual(200, res.StatusCode);
                  }
        });
    }

    [UnityTest]
    public IEnumerator SelectGlobalServerByNameWithTemplatesBroken()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-with-templates-broken");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(
                apiKeyAuth: "token",
                server: SDKConfig.Server.Templated,
                hostname: "broken",
                port: "12345"
            );

            Assert.AreEqual("http://broken:12345", sdk.SDKConfiguration.GetTemplatedServerDetails());

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
}
