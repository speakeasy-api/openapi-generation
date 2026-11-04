#nullable enable
using System.Net.Http;
using No_Security.API;
using No_Security.API.Models.Errors;
using No_Security.API.Models.Shared;
using Xunit;

public class PartialTypesAdditionalShould
{
    [Fact]
    public void ExposeModelMembersFromPartialDeclarations()
    {
        var circle = new ImplicitConversionCircle { Kind = "circle", Radius = 2 };

        Assert.Equal(4.0, circle.Diameter);
    }

    [Fact]
    public void ExposeUnionMembersFromPartialDeclarations()
    {
        var shape = ImplicitConversionShape.FromRadius(3);

        Assert.True(shape.Type.IsCircle);
        Assert.Equal(3.0, shape.ImplicitConversionCircle!.Radius);
    }

    [Fact]
    public void ExposeOpenEnumMembersFromPartialDeclarations()
    {
        Assert.True(Color.Red.IsPrimary);
        Assert.False(Color.Return.IsPrimary);
    }

    [Fact]
    public void ExposeErrorMembersFromPartialDeclarations()
    {
        var payload = new RequestTimeoutErrorPayload { Code = 408, Message = "timeout" };
        var error = new RequestTimeoutError(payload, new HttpResponseMessage(), "");

        Assert.Equal("408: timeout", error.Payload.Summary);
        Assert.True(error.IsRetryable);
    }

    [Fact]
    public void ExposeAuxiliaryErrorMembersFromPartialDeclarations()
    {
        var response = new HttpResponseMessage(System.Net.HttpStatusCode.BadGateway);
        var defaultError = new SDKException("bad gateway", response, "");
        var validationError = new ResponseValidationException("invalid", response, "", new System.Exception("cause"));

        Assert.True(defaultError.IsServerError);
        Assert.True(defaultError.IsDefault);
        Assert.True(validationError.IsServerError);
        Assert.True(validationError.HasCause);
    }

    [Fact]
    public void ExposeSdkMembersFromPartialDeclarations()
    {
        ISDK sdk = new SDK();
        IUnions unions = sdk.Unions;

        Assert.Equal("sdk", sdk.PartialTypesMarker());
        Assert.Equal("unions", unions.PartialTypesMarker());
    }
}
