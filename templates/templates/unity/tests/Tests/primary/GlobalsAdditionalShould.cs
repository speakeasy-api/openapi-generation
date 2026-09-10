using NUnit.Framework;
using UnityEngine.TestTools;
using Openapi;
using System.Collections;

public class GlobalsAdditionalShould
{
    [UnityTest]
    public IEnumerator GlobalsQueryParameterGetUsesGlobal()
    {
        CommonHelpers.RecordTest("globals-query-parameter-get-uses-global");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(globalQueryParam: "test");

            using (var res = await sdk.Globals.GlobalsQueryParameterGetAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("test", res.Res.Args.GlobalQueryParam);
            }
        });
    }

    [UnityTest]
    public IEnumerator GlobalsQueryParameterGetUsesLocal()
    {
        CommonHelpers.RecordTest("globals-query-parameter-get-uses-local");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(globalQueryParam: "test");

            using (var res = await sdk.Globals.GlobalsQueryParameterGetAsync("local"))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("local", res.Res.Args.GlobalQueryParam);
            }
        });
    }

    [UnityTest]
    public IEnumerator GlobalPathParameterGetUsesGlobal()
    {
        CommonHelpers.RecordTest("globals-path-parameter-get-uses-global");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(globalPathParam: 1);

            using (var res = await sdk.Globals.GlobalPathParameterGetAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/globals/pathParameter/1"",
                    res.Res.Url
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator GlobalPathParameterGetUsesLocal()
    {
        CommonHelpers.RecordTest("globals-path-parameter-get-uses-local");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(globalPathParam: 1);

            using (var res = await sdk.Globals.GlobalPathParameterGetAsync(2))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/globals/pathParameter/2"",
                    res.Res.Url
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator GlobalHeaderGetUsesGlobal()
    {
        CommonHelpers.RecordTest("globals-header-get-uses-global");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(globalHeaderParam: true);

            using (var res = await sdk.Globals.GlobalsHeaderGetAsync())
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "true",
                    res.Res.Headers["Globalheaderparam"]
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator GlobalHeaderGetUsesLocal()
    {
        CommonHelpers.RecordTest("globals-header-get-uses-local");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(globalHeaderParam: true);

            using (var res = await sdk.Globals.GlobalsHeaderGetAsync(false))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "false",
                    res.Res.Headers["Globalheaderparam"]
                );
            }
        });
    }
}
