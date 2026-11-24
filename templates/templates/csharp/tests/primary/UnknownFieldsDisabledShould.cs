#nullable enable
using Openapi.Models.Shared;
using Xunit;

public class UnknownFieldsDisabledShould
{
    [Fact]
    public void DropUndeclaredFieldsWhenPreservationIsDisabled()
    {
        var envelope = Compat.DeserializeJson<FieldEnvelope>(
            "{\"label\":\"root\",\"future\":true,\"child\":{\"displayName\":\"child\",\"future\":{\"a\":1}}}"
        );

        var json = CommonHelpers.NormalizeJson(envelope);

        Assert.Equal("child", envelope.Child!.DisplayName);
        Assert.DoesNotContain("\"future\"", json);
        Assert.Null(typeof(FieldEnvelope).GetProperty("UnknownFields"));
        Assert.Null(typeof(FieldNode).GetProperty("UnknownFields"));
    }
}
