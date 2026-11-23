#nullable enable
using Openapi;
using Openapi.Models.Operations;
using Openapi.Models.Shared;
using System.Collections.Generic;
using System.Net.Http;
using System.Threading.Tasks;
using Xunit;

public class OptionalEnumListParamsAdditionalShould
{
    private static readonly string BaseUri = $"{Helpers.HttpBinUrl}/anything/optionalEnumListParams";

    [Fact]
    public async Task SendNothingWhenRequestIsOmitted()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.OptionalEnumListParamsGetAsync();

        AssertSent(res.HttpMeta.Request, BaseUri, null, null);
    }

    [Fact]
    public async Task SendNothingWhenListsAreOmitted()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.OptionalEnumListParamsGetAsync(
            new OptionalEnumListParamsGetRequest() { Filter = new WorkItemFilter() }
        );

        AssertSent(res.HttpMeta.Request, BaseUri, null, null);
    }

    [Fact]
    public async Task SendNothingWhenListsAreEmpty()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.OptionalEnumListParamsGetAsync(
            new OptionalEnumListParamsGetRequest()
            {
                Statuses = new List<WorkItemStatus>(),
                Priorities = new List<WorkItemPriority>(),
                StatusByKey = new Dictionary<string, WorkItemStatus>(),
                Filter = new WorkItemFilter() { Statuses = new List<WorkItemStatus>() },
                XStatuses = new List<WorkItemStatus>(),
                XPriorities = new List<WorkItemPriority>(),
            }
        );

        AssertSent(res.HttpMeta.Request, BaseUri, null, null);
    }

    [Fact]
    public async Task SendEnumWireValues()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.OptionalEnumListParamsGetAsync(
            new OptionalEnumListParamsGetRequest()
            {
                Statuses = new List<WorkItemStatus> { WorkItemStatus.InProgress, WorkItemStatus.Done },
                Priorities = new List<WorkItemPriority> { WorkItemPriority.LowPriority, WorkItemPriority.Of("urgent") },
                StatusByKey = new Dictionary<string, WorkItemStatus> { ["first"] = WorkItemStatus.Done },
                Filter = new WorkItemFilter() { Statuses = new List<WorkItemStatus> { WorkItemStatus.InProgress } },
                XStatuses = new List<WorkItemStatus> { WorkItemStatus.Done, WorkItemStatus.InProgress },
                XPriorities = new List<WorkItemPriority> { WorkItemPriority.HighPriority, WorkItemPriority.Of("urgent") },
            }
        );

        AssertSent(
            res.HttpMeta.Request,
            BaseUri
                + "?filter[statuses]=in-progress"
                + "&priorities=low-priority%2Curgent"
                + "&first=done"
                + "&statuses=in-progress&statuses=done",
            "done,in-progress",
            "high-priority,urgent"
        );
    }

    private static void AssertSent(HttpRequestMessage request, string uri, string? statuses, string? priorities)
    {
        Assert.Equal(uri, request.RequestUri!.OriginalString);
        Assert.Equal(statuses, Header(request, "X-Statuses"));
        Assert.Equal(priorities, Header(request, "X-Priorities"));
    }

    private static string? Header(HttpRequestMessage request, string name) =>
        request.Headers.TryGetValues(name, out var values) ? string.Join(",", values) : null;
}
