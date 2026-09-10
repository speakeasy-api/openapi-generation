using UnityEngine.TestTools;
using NUnit.Framework;
using Openapi;
using Openapi.Models.Operations;
using System;
using System.Collections;

public class FlatteningShould
{
    [UnityTest]
    public IEnumerator ComponentBodyAndParamNoConflict()
    {
        CommonHelpers.RecordTest("flattening-component-body-and-param-no-conflict");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var obj = Helpers.CreateSimpleObject();

            using (
                var res = await sdk.Flattening.ComponentBodyAndParamNoConflictAsync(
                    "param test",
                    Helpers.CreateSimpleObject()
                )
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("param test", res.Res.Args["paramStr"]);
                Helpers.AssertSimpleObject(res.Res.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator ComponentBodyAndParamConflict()
    {
        CommonHelpers.RecordTest("flattening-component-body-and-param-conflict");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Flattening.ComponentBodyAndParamConflictAsync(
                    Helpers.CreateSimpleObject(),
                    "param test"
                )
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("param test", res.Res.Args["str"]);
                Helpers.AssertSimpleObject(res.Res.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator InlineBodyAndParamConflict()
    {
        CommonHelpers.RecordTest("flattening-inline-body-and-param-conflict");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Flattening.InlineBodyAndParamConflictAsync(
                    new InlineBodyAndParamConflictRequestBody() { Str = "body test" },
                    "param test"
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("param test", res.Res.Args["str"]);
                Assert.AreEqual("body test", res.Res.Json.Str);
            }
        });
    }

    [UnityTest]
    public IEnumerator InlineBodyAndParamNoConflict()
    {
        CommonHelpers.RecordTest("flattening-inline-body-and-param-no-conflict");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Flattening.InlineBodyAndParamNoConflictAsync(
                    new InlineBodyAndParamNoConflictRequestBody() { BodyStr = "body test" },
                    "param test"
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("param test", res.Res.Args["paramStr"]);
                Assert.AreEqual("body test", res.Res.Json.BodyStr);
            }
        });
    }

    [UnityTest]
    public IEnumerator ConflictingParams()
    {
        CommonHelpers.RecordTest("flattening-conflicting-params");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.Flattening.ConflictingParamsAsync("pathParam", "queryParam"))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.That(res.Res.Url, Does.Contain("/pathParam?"));
                Assert.AreEqual("queryParam", res.Res.Args["str"]);
            }
        });
    }
}
