using System.Net.Http;
using System.Threading;
using System.Threading.Tasks;
using Openapi;
using Openapi.Models.Errors;
using Openapi.Models.Operations;
using Openapi.Utils.Retries;
using Openapi.Hooks;
using Xunit;
using System.Net;
using System;

public class CancellationTokenShould
{

    [Fact]
    public async Task CancellationTokenNoCancellation()
    {
        // Tests that HTTP client handles non-cancelling tokens correctly
        // Both null and CancellationToken.None should complete successfully
        CommonHelpers.RecordTest("cancellation-token-no-cancellation");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // Test 1: Explicit null cancellation token
        var res1 = await sdk.Cancellation.CancelledRequestAsync(1, cancellationToken: null);
        Assert.Equal(HttpStatusCode.OK, res1.HttpMeta.Response.StatusCode);
        Assert.NotNull(res1);

        // Test 2: CancellationToken.None
        var res2 = await sdk.Cancellation.CancelledRequestAsync(1, cancellationToken: CancellationToken.None);
        Assert.Equal(HttpStatusCode.OK, res2.HttpMeta.Response.StatusCode);
        Assert.NotNull(res2);
    }

    [Fact]
    public async Task CancellationTokenCancelledBeforeRequest()
    {
        // Tests that pre-cancelled tokens are detected and throw OperationCanceledException
        // Token is cancelled before HTTP request execution (caught by hook context check)
        CommonHelpers.RecordTest("cancellation-token-cancelled-before-request");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var cts = new CancellationTokenSource();

        // Cancel the token before making the request
        cts.Cancel();

        // Should throw OperationCanceledException from hook context check
        await Assert.ThrowsAsync<OperationCanceledException>(
            async () => await sdk.Cancellation.CancelledRequestAsync(5, cancellationToken: cts.Token)
        );
    }

    [Fact]
    public async Task CancellationTokenCancelledDuringRequest()
    {
        // Tests that HTTP client properly cancels during request execution
        // Token cancels after 2s while waiting for 10s delay response
        CommonHelpers.RecordTest("cancellation-token-cancelled-during-request");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var cts = new CancellationTokenSource();

        // Set reasonable timeout to prevent test from running too long
        cts.CancelAfter(TimeSpan.FromSeconds(2));

        // Use a long delay (10 seconds) that should be cancelled during execution
        await Assert.ThrowsAsync<TaskCanceledException>(
            async () => await sdk.Cancellation.CancelledRequestAsync(100, cancellationToken: cts.Token)
        );
    }

    [Fact]
    public async Task CancellationTokenCancelledInsideHook()
    {
        // Tests that token is cancelled within the hook execution
        // Hook should start but not complete due to cancellation during hook delay
        CommonHelpers.RecordTest("cancellation-token-cancelled-inside-hook");

        TestCancellationHook.Reset();

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var cts = new CancellationTokenSource();

        // Cancel after 2 seconds - hook delays for 5 seconds, so cancellation happens during hook
        cts.CancelAfter(TimeSpan.FromSeconds(2));

        await Assert.ThrowsAsync<TaskCanceledException>(
            async () => await sdk.Cancellation.CancellationHookAsync(10, cancellationToken: cts.Token)
        );

        // Verify hook was entered but did not complete due to cancellation
        Assert.True(TestCancellationHook.WasStarted, "Hook should have started");
        Assert.False(TestCancellationHook.WasCompleted, "Hook should NOT have completed - proves cancellation happened within hook");
    }

}