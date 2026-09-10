#nullable enable
using System;
using System.Collections;
using NUnit.Framework;
using UnityEngine.TestTools;
using Openapi;
using Openapi.Models.Operations;
using Openapi.Models.Shared;
using Openapi.Models.Errors;

public class ErrorsShould
{
    [UnityTest]
    public IEnumerator TestStatusGetError_DefaultErrorCodes()
    {
        CommonHelpers.RecordTest("errors-status-get-error-default-error-codes");
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            APIException? exception = null;

            try
            {
                using (var res = await sdk.Errors.StatusGetErrorAsync(400)) { }
            } catch (APIException e)
            {
                exception = e;
            }
            Assert.AreEqual(400, exception!.StatusCode);

            try
            {
                using (var res = await sdk.Errors.StatusGetErrorAsync(500)) { }
            } catch (APIException e)
            {
                exception = e;
            }
            Assert.AreEqual(500, exception!.StatusCode);
        });
    }

    [UnityTest]
    public IEnumerator TestStatusGetError_300_NonError()
    {
        CommonHelpers.RecordTest("errors-status-get-error300-non-error");
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.Errors.StatusGetErrorAsync(300)) {
                Assert.AreEqual(300, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator TestStatusGetErrorXSpeakeasyErrors()
    {
        System.Console.WriteLine("Begin TestStatusGetErrorXSpeakeasyErrors");
        CommonHelpers.RecordTest("errors-status-get-error-x-speakeasy-errors");
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            APIException? exception = null;

            try
            {
                using (var res = await sdk.Errors.StatusGetXSpeakeasyErrorsAsync(400)) { }
            } catch (APIException e)
            {
                exception = e;
            }
            Assert.AreEqual("API error occurred: Status 400\n{\"message\":\"an error occurred\",\"code\":\"400\",\"type\":\"internal\"}\n", exception!.ToString());
            Assert.AreEqual(400, exception!.StatusCode);

            try
            {
                using (var res = await sdk.Errors.StatusGetXSpeakeasyErrorsAsync(401)) { }
            } catch (APIException e)
            {
                exception = e;
            }

            Assert.AreEqual("API error occurred: Status 401\n{\"message\":\"an error occurred\",\"code\":\"401\",\"type\":\"internal\"}\n", exception!.ToString());
            Assert.AreEqual(401, exception!.StatusCode);

            try
            {
                using (var res = await sdk.Errors.StatusGetXSpeakeasyErrorsAsync(402)) { }
            } catch (APIException e)
            {
                exception = e;
            }

            Assert.AreEqual("API error occurred: Status 402\n{\"message\":\"an error occurred\",\"code\":\"402\",\"type\":\"internal\"}\n", exception!.ToString());
            Assert.AreEqual(402, exception!.StatusCode);

            Error? error = null;
            try
            {
                using (var res = await sdk.Errors.StatusGetXSpeakeasyErrorsAsync(500)) { }
            } catch (Error e)
            {
                error = e;
            }

            Assert.AreEqual("an error occurred", error!.Message);
            Assert.AreEqual("500", error!.Code);

            StatusGetXSpeakeasyErrorsResponseBody? error2 = null;
            try
            {
                using (var res = await sdk.Errors.StatusGetXSpeakeasyErrorsAsync(501)) { }
            } catch (StatusGetXSpeakeasyErrorsResponseBody e)
            {
                error2 = e;
            }
            Assert.AreEqual("501", error2!.Code);
        });
    }

    /* TODO: currently disabled as its not working
    [UnityTest]
    public IEnumerator TestConnectionErrorGet()
    {
        System.Console.WriteLine("Begin TestConnectionErrorGet");
        CommonHelpers.RecordTest("errors-connection-error");
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            Exception? ex = null;
            try
            {
                using (var res = await sdk.Errors.ConnectionErrorGetAsync()) {}
            } catch (Exception e) {
                ex = e;
                System.Console.WriteLine("Exception: " + ex.Message);
                System.Console.WriteLine("Exception: " + ex.GetType());
            }
            Assert.AreEqual("Insecure connection not allowed", ex!.Message);
        });
    }*/
}
