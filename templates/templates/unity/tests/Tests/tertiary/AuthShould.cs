#nullable enable
using NUnit.Framework;
using UnityEngine.TestTools;
using System;
using System.Collections;
using System.Collections.Generic;
using No_Security.API;
using No_Security.API.Models.Errors;

public class AuthShould
{
    [UnityTest]
    public IEnumerator NoAuth()
    {
        CommonHelpers.RecordTest("auth-no-auth");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Auth.NoAuthAsync()
            )
            {
                Assert.NotNull(res);
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator ApiKeyAuthGlobal()
    {
        CommonHelpers.RecordTest("auth-api-key-auth-global");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            APIException? exception = null;
            try
            {
                using (var res = await sdk.Auth.ApiKeyAuthGlobalAsync()) { }
            }
            catch (APIException e)
            {
                exception = e;
            }
            Assert.NotNull(exception);
            Assert.AreEqual(401, exception.StatusCode);
        });
    }
}
