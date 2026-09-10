using System;
using System.Collections;
using NUnit.Framework;
using UnityEngine.TestTools;
using HoistedSecurity;

public class ServersShould
{
    [UnityTest]
    public IEnumerator SelectGlobalServerByNameDefault()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-name-default");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(server: null);

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
            var sdk = new SDK(server: SDKConfig.Server.DefaultServer);

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
    public IEnumerator SelectGlobalServerByNameBroken()
    {
            CommonHelpers.RecordTest("servers-select-global-server-by-name-broken");

            yield return CommonHelpers.Await(async () =>
            {
                var sdk = new SDK(server: SDKConfig.Server.BrokenServer);

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
