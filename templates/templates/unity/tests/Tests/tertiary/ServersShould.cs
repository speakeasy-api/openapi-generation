#nullable enable
using System;
using System.Collections;
using System.Collections.Generic;
using NUnit.Framework;
using UnityEngine.TestTools;
using No_Security.API;

public class ServersShould
{
    [UnityTest]
    public IEnumerator SelectGlobalServerByIdDefault()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-id-default");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

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
    public IEnumerator SelectGlobalServerByIdValid()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-id-valid");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(serverIndex: 0);

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
    public IEnumerator SelectGlobalServerByIdInValid()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-id-invalid");


        yield return CommonHelpers.Await(async () =>
        {
            Exception? exception = null;
            try
            {
                new SDK(serverIndex: 2);
            }
            catch (Exception e)
            {
                exception = e;
            }

            Assert.NotNull(exception);
            Assert.AreEqual("Invalid server index 2", exception.Message);

        });
    }

    [UnityTest]
    public IEnumerator SelectGlobalServerByIdBroken()
    {
        CommonHelpers.RecordTest("servers-select-global-server-by-id-broken");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(serverIndex: 1);

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
