using Xunit;
using Speakeasy.BasicHttp;
using Speakeasy.BasicHttp.Models.Components;
using Speakeasy.BasicHttp.Models.Requests;
using System.Net;
using System.Threading.Tasks;

public class AuthAdditionalTests
{
    [Fact]
    public async Task BasicAuthOperationOptional()
    {
        CommonHelpers.RecordTest("auth-basic-auth-operation-optional");


        var sdk = new SDK(
            security: new Security()
            {
                Username = "wrongUser",
                Password = "wrongPass"
            }
        );

        var res = await sdk.Auth.BasicAuthOptionalAsync(
            security: new BasicAuthOptionalSecurity()
            {
                Username = "testUser",
                Password = "testPass"
            }
        );

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.BasicAuthResponse);
        Assert.True(res.BasicAuthResponse.Authenticated);
        Assert.Equal("testUser", res.BasicAuthResponse.User);


        var sdk2 = new SDK(
            security: new Security()
            {
                Username = "testUser",
                Password = "testPass"
            }
        );

        await Assert.ThrowsAsync<Speakeasy.BasicHttp.Models.Errors.APIException>(async () =>
        {
            await sdk2.Auth.BasicAuthOptionalAsync(
                security: new BasicAuthOptionalSecurity()
            );
        });
    }
}
