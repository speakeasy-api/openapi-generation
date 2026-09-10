using System;
using Xunit;
using Openapi;
using Openapi.Models.Shared;
using System.Collections.Generic;
using System.Threading.Tasks;
using Openapi.Models.Operations;
using FluentAssertions;
using System.Net;
using Openapi.Utils;

public class ParametersShould
{
    [Fact]
    public async Task MixedParameters()
    {
        CommonHelpers.RecordTest("parameters-mixed-primitives");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.MixedParametersPrimitivesAsync(
            "headerValue",
            "pathValue",
            "queryValue"
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/mixedParams/path/pathValue?queryStringParam=queryValue",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(res.HttpMeta.Request.RequestUri.OriginalString, res.Res.Url);
        Assert.Equal("headerValue", res.Res.Headers.Headerparam);
        Assert.Equal("queryValue", res.Res.Args.QueryStringParam);
    }

    [Fact]
    public async Task CamelCase()
    {
        CommonHelpers.RecordTest("parameters-camel-case");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.MixedParametersCamelCaseAsync(
            "headerValue",
            "pathValue",
            "queryValue"
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/mixedParams/path/pathValue/camelcase?query_string_param=queryValue",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(res.HttpMeta.Request.RequestUri.OriginalString, res.Res.Url);
        Assert.Equal("headerValue", res.Res.Headers.HeaderParam);
        Assert.Equal("queryValue", res.Res.Args.QueryStringParam);
    }

    [Fact]
    public async Task SimplePathParameterPrimitives()
    {
        CommonHelpers.RecordTest("parameters-simple-path-parameter-primitives");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.SimplePathParameterPrimitivesAsync(true, 1, 1.1D, "test :/?#[]@!$&'()*+,=;");

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/pathParams/str/test%20%3A%2F%3F%23%5B%5D%40%21%24%26%27%28%29%2A%2B%2C%3D%3B/bool/true/int/1/num/1.1",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/pathParams/str/test :/%3F#[]@!$&'()*+,=%3B/bool/true/int/1/num/1.1",
            res.Res.Url
        );
    }

    [Fact]
    public async Task SimplePathParameterObjects()
    {
        CommonHelpers.RecordTest("parameters-simple-path-parameter-objects");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.SimplePathParameterObjectsAsync(
            Helpers.CreateSimpleObject(),
            Helpers.CreateSimpleObject()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/pathParams/obj/any%2Cany%2Cbool%2Ctrue%2CboolOpt%2Ctrue%2Cdate%2C2020-01-01%2CdateTime%2C2020-01-01T00%3A00%3A00.0000001Z%2Cenum%2Cone%2Cfloat32%2C1.1%2Cint%2C1%2Cint32%2C1%2Cint32Enum%2C55%2CintEnum%2C2%2Cnum%2C1.1%2Cstr%2Ctest%2CstrOpt%2CtestOptional/objExploded/any%3Dany%2Cbool%3Dtrue%2CboolOpt%3Dtrue%2Cdate%3D2020-01-01%2CdateTime%3D2020-01-01T00%3A00%3A00.0000001Z%2Cenum%3Done%2Cfloat32%3D1.1%2Cint%3D1%2Cint32%3D1%2Cint32Enum%3D55%2CintEnum%3D2%2Cnum%3D1.1%2Cstr%3Dtest%2CstrOpt%3DtestOptional",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/pathParams/obj/any,any,bool,true,boolOpt,true,date,2020-01-01,dateTime,2020-01-01T00:00:00.0000001Z,enum,one,float32,1.1,int,1,int32,1,int32Enum,55,intEnum,2,num,1.1,str,test,strOpt,testOptional/objExploded/any=any,bool=true,boolOpt=true,date=2020-01-01,dateTime=2020-01-01T00:00:00.0000001Z,enum=one,float32=1.1,int=1,int32=1,int32Enum=55,intEnum=2,num=1.1,str=test,strOpt=testOptional",
            res.Res.Url
        );
    }

    [Fact]
    public async Task SimplePathParameterArrays()
    {
        CommonHelpers.RecordTest("parameters-simple-path-parameter-arrays");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.SimplePathParameterArraysAsync(
            new List<string>() { "test", "test2" }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/pathParams/arr/test%2Ctest2",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/pathParams/arr/test,test2",
            res.Res.Url
        );
    }

    [Fact]
    public async Task SimplePathParameterMaps()
    {
        CommonHelpers.RecordTest("parameters-simple-path-parameter-maps");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.SimplePathParameterMapsAsync(
            new Dictionary<string, string>() { { "test", "value" }, { "test2", "value2" } },
            new Dictionary<string, long>() { { "test", 1 }, { "test2", 2 } }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/pathParams/map/test%2Cvalue%2Ctest2%2Cvalue2/mapExploded/test%3D1%2Ctest2%3D2",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/pathParams/map/test,value,test2,value2/mapExploded/test=1,test2=2",
            res.Res.Url
        );
    }

    [Fact]
    public async Task PathParameterJson()
    {
        CommonHelpers.RecordTest("parameters-path-parameter-json");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.PathParameterJsonAsync(Helpers.CreateSimpleObject());

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/pathParams/json/%7B%22any%22%3A%22any%22%2C%22bool%22%3Atrue%2C%22boolOpt%22%3Atrue%2C%22date%22%3A%222020-01-01%22%2C%22dateTime%22%3A%222020-01-01T00%3A00%3A00.0000001Z%22%2C%22enum%22%3A%22one%22%2C%22float32%22%3A1.1%2C%22int%22%3A1%2C%22int32%22%3A1%2C%22int32Enum%22%3A55%2C%22intEnum%22%3A2%2C%22num%22%3A1.1%2C%22str%22%3A%22test%22%2C%22strOpt%22%3A%22testOptional%22%7D",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/pathParams/json/{{\"any\":\"any\",\"bool\":true,\"boolOpt\":true,\"date\":\"2020-01-01\",\"dateTime\":\"2020-01-01T00:00:00.0000001Z\",\"enum\":\"one\",\"float32\":1.1,\"int\":1,\"int32\":1,\"int32Enum\":55,\"intEnum\":2,\"num\":1.1,\"str\":\"test\",\"strOpt\":\"testOptional\"}}",
            res.Res.Url
        );
    }

    [Fact]
    public async Task FormQueryParamsPrimitive()
    {
        CommonHelpers.RecordTest("parameters-form-query-params-primitive");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.FormQueryParamsPrimitiveAsync(true, 1, 1.1D, "test :/?#[]@!$&'()*+,=;");

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/form/primitive?boolParam=true&intParam=1&numParam=1.1&strParam=test+%3A%2F%3F%23%5B%5D%40!%24%26%27()*%2B%2C%3D%3B",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/form/primitive?boolParam=true&intParam=1&numParam=1.1&strParam=test+%3A%2F%3F%23[]%40!%24%26'()*%2B%2C%3D%3B",
            res.Res.Url
        );
    }

    [Fact]
    public async Task FormQueryParamsObject()
    {
        CommonHelpers.RecordTest("parameters-form-query-params-object");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.FormQueryParamsObjectAsync(
            Helpers.CreateSimpleObject(),
            Helpers.CreateSimpleObject()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/form/obj?objParam=any%2Cany%2Cbool%2Ctrue%2CboolOpt%2Ctrue%2Cdate%2C2020-01-01%2CdateTime%2C2020-01-01T00%3A00%3A00.0000001Z%2Cenum%2Cone%2Cfloat32%2C1.1%2Cint%2C1%2Cint32%2C1%2Cint32Enum%2C55%2CintEnum%2C2%2Cnum%2C1.1%2Cstr%2Ctest%2CstrOpt%2CtestOptional&any=any&bool=true&boolOpt=true&date=2020-01-01&dateTime=2020-01-01T00%3A00%3A00.0000001Z&enum=one&float32=1.1&int=1&int32=1&int32Enum=55&intEnum=2&num=1.1&str=test&strOpt=testOptional",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(res.HttpMeta.Request.RequestUri.OriginalString, res.Res.Url);
    }

    [Fact]
    public async Task FormQueryParamsArray()
    {
        CommonHelpers.RecordTest("parameters-form-query-params-array");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.FormQueryParamsArrayAsync(
            new List<string>() { "test", "test2" },
            new List<long>() { 1, 2 }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/form/array?arrParam=test%2Ctest2&arrParamExploded=1&arrParamExploded=2",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(res.HttpMeta.Request.RequestUri.OriginalString, res.Res.Url);
        Assert.Equal("test,test2", res.Res.Args.ArrParam);
        Assert.Equal(new List<string>() { "1", "2" }, res.Res.Args.ArrParamExploded);
    }

    [Fact]
    public async Task PipeDelimitedQueryParamsArray()
    {
        CommonHelpers.RecordTest("parameters-pipe-query-params-array");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.PipeDelimitedQueryParamsArrayAsync(
            new List<string>() { "test", "test2" },
            new List<long> { 1, 2 },
            new Dictionary<string, string>() { { "key1", "val1" }, { "key2", "val2" } },
            Helpers.CreateSimpleObject()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/pipe/array?arrParam=test%7Ctest2&arrParamExploded=1&arrParamExploded=2&mapParam=key1%7Cval1%7Ckey2%7Cval2&objParam=any%7Cany%7Cbool%7Ctrue%7CboolOpt%7Ctrue%7Cdate%7C2020-01-01%7CdateTime%7C2020-01-01T00%3A00%3A00.0000001Z%7Cenum%7Cone%7Cfloat32%7C1.1%7Cint%7C1%7Cint32%7C1%7Cint32Enum%7C55%7CintEnum%7C2%7Cnum%7C1.1%7Cstr%7Ctest%7CstrOpt%7CtestOptional",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/pipe/array?arrParam=test|test2&arrParamExploded=1&arrParamExploded=2&mapParam=key1|val1|key2|val2&objParam=any|any|bool|true|boolOpt|true|date|2020-01-01|dateTime|2020-01-01T00%3A00%3A00.0000001Z|enum|one|float32|1.1|int|1|int32|1|int32Enum|55|intEnum|2|num|1.1|str|test|strOpt|testOptional",
            res.Res.Url
        );
        Assert.Equal("test|test2", res.Res.Args.ArrParam);
        Assert.Equal(new List<string>() { "1", "2" }, res.Res.Args.ArrParamExploded);
    }

    [Fact]
    public async Task FormQueryParamsMap()
    {
        CommonHelpers.RecordTest("parameters-form-query-params-map");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.FormQueryParamsMapAsync(
            new Dictionary<string, string>() { { "test", "value" }, { "test2", "value2" } },
            new Dictionary<string, long>() { { "test", 1 }, { "test2", 2 } }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/form/map?mapParam=test%2Cvalue%2Ctest2%2Cvalue2&test=1&test2=2",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/form/map?mapParam=test%2Cvalue%2Ctest2%2Cvalue2&test=1&test2=2",
            res.Res.Url
        );
        Assert.Equal(
            new Dictionary<string, string>()
            {
                { "mapParam", "test,value,test2,value2" },
                { "test", "1" },
                { "test2", "2" }
            },
            res.Res.Args
        );
    }

    [Fact]
    public async Task FormQueryParamsRefParamObject()
    {
        CommonHelpers.RecordTest("parameters-form-query-params-ref-param-object");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.FormQueryParamsRefParamObjectAsync(
            new RefQueryParamObj()
            {
                Bool = true,
                Int = 1,
                Num = 1.1D,
                Str = "test"
            },
            new RefQueryParamObjExploded()
            {
                Bool = true,
                Int = 1,
                Num = 1.1D,
                Str = "test"
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/form/refParamObject?refObjParam=bool%2Ctrue%2Cint%2C1%2Cnum%2C1.1%2Cstr%2Ctest&bool=true&int=1&num=1.1&str=test",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/form/refParamObject?refObjParam=bool%2Ctrue%2Cint%2C1%2Cnum%2C1.1%2Cstr%2Ctest&bool=true&int=1&num=1.1&str=test",
            res.Res.Url
        );
        Assert.Equal("true", res.Res.Args.Bool);
        Assert.Equal("1", res.Res.Args.Int);
        Assert.Equal("1.1", res.Res.Args.Num);
        Assert.Equal("test", res.Res.Args.Str);
        Assert.Equal("bool,true,int,1,num,1.1,str,test", res.Res.Args.RefObjParam);
    }

    [Fact]
    public async Task DeepObjectQueryParamsObject()
    {
        CommonHelpers.RecordTest("parameters-deep-object-query-params-object");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.DeepObjectQueryParamsObjectAsync(
            Helpers.CreateSimpleObject(),
            new ObjArrParam()
            {
                Arr = new List<string> { "test", "test2" }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/deepObject/obj?objArrParam[arr]=test&objArrParam[arr]=test2&objParam[any]=any&objParam[bool]=true&objParam[boolOpt]=true&objParam[date]=2020-01-01&objParam[dateTime]=2020-01-01T00%3A00%3A00.0000001Z&objParam[enum]=one&objParam[float32]=1.1&objParam[int]=1&objParam[int32]=1&objParam[int32Enum]=55&objParam[intEnum]=2&objParam[num]=1.1&objParam[str]=test&objParam[strOpt]=testOptional",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(res.HttpMeta.Request.RequestUri.OriginalString, res.Res.Url);
        Assert.Equal(new string[] { "test", "test2" }, res.Res.Args.ObjArrParamArr);
        Assert.Equal("any", res.Res.Args.ObjParamAny);
        Assert.Equal("true", res.Res.Args.ObjParamBool);
        Assert.Equal("true", res.Res.Args.ObjParamBoolOpt);
        Assert.Equal("2020-01-01", res.Res.Args.ObjParamDate);
        Assert.Equal("2020-01-01T00:00:00.0000001Z", res.Res.Args.ObjParamDateTime);
        Assert.Equal("one", res.Res.Args.ObjParamEnum);
        Assert.Equal("1.1", res.Res.Args.ObjParamFloat32);
        Assert.Equal("1", res.Res.Args.ObjParamInt32);
        Assert.Equal("1", res.Res.Args.ObjParamInt);
        Assert.Equal("1.1", res.Res.Args.ObjParamNum);
        Assert.Equal("test", res.Res.Args.ObjParamStr);
        Assert.Equal("testOptional", res.Res.Args.ObjParamStrOpt);
    }

    [Fact]
    public async Task DeepObjectQueryParamsMap()
    {
        CommonHelpers.RecordTest("parameters-deep-object-query-params-map");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.DeepObjectQueryParamsMapAsync(
            new Dictionary<string, string>() { { "test", "value" }, { "test2", "value2" } },
            new Dictionary<string, List<string>>()
            {
                {
                    "test",
                    new List<string>() { "value", "value2" }
                },
                {
                    "test2",
                    new List<string>() { "value3", "value4" }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/deepObject/map?mapArrParam[test]=value&mapArrParam[test]=value2&mapArrParam[test2]=value3&mapArrParam[test2]=value4&mapParam[test]=value&mapParam[test2]=value2",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(res.HttpMeta.Request.RequestUri.OriginalString, res.Res.Url);
        res.Res.Args.Should().BeEquivalentTo(
            new Dictionary<string, DeepObjectQueryParamsMapArgs>()
            {
                {
                    "mapArrParam[test2]",
                    new DeepObjectQueryParamsMapArgs(DeepObjectQueryParamsMapArgsType.ArrayOfStr) {
                        ArrayOfStr = new List<string>() { "value3", "value4" }
                    }
                },
                {
                    "mapArrParam[test]",
                    new DeepObjectQueryParamsMapArgs(DeepObjectQueryParamsMapArgsType.ArrayOfStr) {
                        ArrayOfStr = new List<string>() { "value", "value2" }
                    }
                },
                { "mapParam[test2]",
                    new DeepObjectQueryParamsMapArgs(DeepObjectQueryParamsMapArgsType.Str) {
                        Str = "value2"
                    }
                },
                { "mapParam[test]", 
                    new DeepObjectQueryParamsMapArgs(DeepObjectQueryParamsMapArgsType.Str) {
                        Str = "value"
                    }
                }
            }
        );
    }

    [Fact]
    public async Task JsonQueryParamsObject()
    {
        CommonHelpers.RecordTest("parameters-json-query-params-object");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.JsonQueryParamsObjectAsync(
            Helpers.CreateDeepObject(),
            Helpers.CreateSimpleObject()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/json/obj?deepObjParam=%7B%22any%22%3A%7B%22any%22%3A%22any%22%2C%22bool%22%3Atrue%2C%22boolOpt%22%3Atrue%2C%22date%22%3A%222020-01-01%22%2C%22dateTime%22%3A%222020-01-01T00%3A00%3A00.0000001Z%22%2C%22enum%22%3A%22one%22%2C%22float32%22%3A1.1%2C%22int%22%3A1%2C%22int32%22%3A1%2C%22int32Enum%22%3A55%2C%22intEnum%22%3A2%2C%22num%22%3A1.1%2C%22str%22%3A%22test%22%2C%22strOpt%22%3A%22testOptional%22%7D%2C%22arr%22%3A%5B%7B%22any%22%3A%22any%22%2C%22bool%22%3Atrue%2C%22boolOpt%22%3Atrue%2C%22date%22%3A%222020-01-01%22%2C%22dateTime%22%3A%222020-01-01T00%3A00%3A00.0000001Z%22%2C%22enum%22%3A%22one%22%2C%22float32%22%3A1.1%2C%22int%22%3A1%2C%22int32%22%3A1%2C%22int32Enum%22%3A55%2C%22intEnum%22%3A2%2C%22num%22%3A1.1%2C%22str%22%3A%22test%22%2C%22strOpt%22%3A%22testOptional%22%7D%2C%7B%22any%22%3A%22any%22%2C%22bool%22%3Atrue%2C%22boolOpt%22%3Atrue%2C%22date%22%3A%222020-01-01%22%2C%22dateTime%22%3A%222020-01-01T00%3A00%3A00.0000001Z%22%2C%22enum%22%3A%22one%22%2C%22float32%22%3A1.1%2C%22int%22%3A1%2C%22int32%22%3A1%2C%22int32Enum%22%3A55%2C%22intEnum%22%3A2%2C%22num%22%3A1.1%2C%22str%22%3A%22test%22%2C%22strOpt%22%3A%22testOptional%22%7D%5D%2C%22bool%22%3Atrue%2C%22int%22%3A1%2C%22map%22%3A%7B%22key%22%3A%7B%22any%22%3A%22any%22%2C%22bool%22%3Atrue%2C%22boolOpt%22%3Atrue%2C%22date%22%3A%222020-01-01%22%2C%22dateTime%22%3A%222020-01-01T00%3A00%3A00.0000001Z%22%2C%22enum%22%3A%22one%22%2C%22float32%22%3A1.1%2C%22int%22%3A1%2C%22int32%22%3A1%2C%22int32Enum%22%3A55%2C%22intEnum%22%3A2%2C%22num%22%3A1.1%2C%22str%22%3A%22test%22%2C%22strOpt%22%3A%22testOptional%22%7D%7D%2C%22num%22%3A1.1%2C%22obj%22%3A%7B%22any%22%3A%22any%22%2C%22bool%22%3Atrue%2C%22boolOpt%22%3Atrue%2C%22date%22%3A%222020-01-01%22%2C%22dateTime%22%3A%222020-01-01T00%3A00%3A00.0000001Z%22%2C%22enum%22%3A%22one%22%2C%22float32%22%3A1.1%2C%22int%22%3A1%2C%22int32%22%3A1%2C%22int32Enum%22%3A55%2C%22intEnum%22%3A2%2C%22num%22%3A1.1%2C%22str%22%3A%22test%22%2C%22strOpt%22%3A%22testOptional%22%7D%2C%22str%22%3A%22test%22%7D&simpleObjParam=%7B%22any%22%3A%22any%22%2C%22bool%22%3Atrue%2C%22boolOpt%22%3Atrue%2C%22date%22%3A%222020-01-01%22%2C%22dateTime%22%3A%222020-01-01T00%3A00%3A00.0000001Z%22%2C%22enum%22%3A%22one%22%2C%22float32%22%3A1.1%2C%22int%22%3A1%2C%22int32%22%3A1%2C%22int32Enum%22%3A55%2C%22intEnum%22%3A2%2C%22num%22%3A1.1%2C%22str%22%3A%22test%22%2C%22strOpt%22%3A%22testOptional%22%7D",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/json/obj?deepObjParam={{\"any\"%3A{{\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.0000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}}%2C\"arr\"%3A[{{\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.0000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}}%2C{{\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.0000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}}]%2C\"bool\"%3Atrue%2C\"int\"%3A1%2C\"map\"%3A{{\"key\"%3A{{\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.0000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}}}}%2C\"num\"%3A1.1%2C\"obj\"%3A{{\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.0000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}}%2C\"str\"%3A\"test\"}}&simpleObjParam={{\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.0000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}}",
            res.Res.Url
        );
    }

    [Fact]
    public async Task MixedQueryParams()
    {
        CommonHelpers.RecordTest("parameters-mixed-query-params");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.MixedQueryParamsAsync(
            Helpers.CreateSimpleObject(),
            Helpers.CreateSimpleObject(),
            Helpers.CreateSimpleObject()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/mixed?deepObjectParam[any]=any&deepObjectParam[bool]=true&deepObjectParam[boolOpt]=true&deepObjectParam[date]=2020-01-01&deepObjectParam[dateTime]=2020-01-01T00%3A00%3A00.0000001Z&deepObjectParam[enum]=one&deepObjectParam[float32]=1.1&deepObjectParam[int]=1&deepObjectParam[int32]=1&deepObjectParam[int32Enum]=55&deepObjectParam[intEnum]=2&deepObjectParam[num]=1.1&deepObjectParam[str]=test&deepObjectParam[strOpt]=testOptional&any=any&bool=true&boolOpt=true&date=2020-01-01&dateTime=2020-01-01T00%3A00%3A00.0000001Z&enum=one&float32=1.1&int=1&int32=1&int32Enum=55&intEnum=2&num=1.1&str=test&strOpt=testOptional&jsonParam=%7B%22any%22%3A%22any%22%2C%22bool%22%3Atrue%2C%22boolOpt%22%3Atrue%2C%22date%22%3A%222020-01-01%22%2C%22dateTime%22%3A%222020-01-01T00%3A00%3A00.0000001Z%22%2C%22enum%22%3A%22one%22%2C%22float32%22%3A1.1%2C%22int%22%3A1%2C%22int32%22%3A1%2C%22int32Enum%22%3A55%2C%22intEnum%22%3A2%2C%22num%22%3A1.1%2C%22str%22%3A%22test%22%2C%22strOpt%22%3A%22testOptional%22%7D",
            res.HttpMeta.Request.RequestUri.OriginalString
        );
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/queryParams/mixed?deepObjectParam[any]=any&deepObjectParam[bool]=true&deepObjectParam[boolOpt]=true&deepObjectParam[date]=2020-01-01&deepObjectParam[dateTime]=2020-01-01T00%3A00%3A00.0000001Z&deepObjectParam[enum]=one&deepObjectParam[float32]=1.1&deepObjectParam[int]=1&deepObjectParam[int32]=1&deepObjectParam[int32Enum]=55&deepObjectParam[intEnum]=2&deepObjectParam[num]=1.1&deepObjectParam[str]=test&deepObjectParam[strOpt]=testOptional&any=any&bool=true&boolOpt=true&date=2020-01-01&dateTime=2020-01-01T00%3A00%3A00.0000001Z&enum=one&float32=1.1&int=1&int32=1&int32Enum=55&intEnum=2&num=1.1&str=test&strOpt=testOptional&jsonParam={{\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.0000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}}",
            res.Res.Url
        );
    }

    [Fact]
    public async Task HeaderParamsPrimitive()
    {
        CommonHelpers.RecordTest("parameters-header-params-primitive");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.HeaderParamsPrimitiveAsync(true, 1, 1.1D, "test");

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("true", res.Res.Headers.XHeaderBoolean);
        Assert.Equal("1", res.Res.Headers.XHeaderInteger);
        Assert.Equal("1.1", res.Res.Headers.XHeaderNumber);
        Assert.Equal("test", res.Res.Headers.XHeaderString);
    }

    [Fact]
    public async Task HeaderParamsObject()
    {
        CommonHelpers.RecordTest("parameters-header-params-object");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.HeaderParamsObjectAsync(
            Helpers.CreateSimpleObject(),
            Helpers.CreateSimpleObject()
        );


        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            "any,any,bool,true,boolOpt,true,date,2020-01-01,dateTime,2020-01-01T00:00:00.0000001Z,enum,one,float32,1.1,int,1,int32,1,int32Enum,55,intEnum,2,num,1.1,str,test,strOpt,testOptional",
            res.Res.Headers.XHeaderObj
        );
        Assert.Equal(
            "any=any,bool=true,boolOpt=true,date=2020-01-01,dateTime=2020-01-01T00:00:00.0000001Z,enum=one,float32=1.1,int=1,int32=1,int32Enum=55,intEnum=2,num=1.1,str=test,strOpt=testOptional",
            res.Res.Headers.XHeaderObjExplode
        );
    }

    [Fact]
    public async Task HeaderParamsMap()
    {
        CommonHelpers.RecordTest("parameters-header-params-map");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.HeaderParamsMapAsync(
            new Dictionary<string, string>() { { "key1", "value1" }, { "key2", "value2" } },
            new Dictionary<string, string>() { { "test1", "val1" }, { "test2", "val2" } }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("key1,value1,key2,value2", res.Res.Headers.XHeaderMap);
        Assert.Equal("test1=val1,test2=val2", res.Res.Headers.XHeaderMapExplode);
    }

    [Fact]
    public async Task HeaderParamsArray()
    {
        CommonHelpers.RecordTest("parameters-header-params-array");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.HeaderParamsArrayAsync(
            new List<string> { "test1", "test2" }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("test1,test2", res.Res.Headers.XHeaderArray);
    }

    [Fact]
    public async Task AllowEmptyValueQueryParams()
    {
        CommonHelpers.RecordTest("parameters-allow-empty-value");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.AllowEmptyValueQueryParamsAsync(
            new List<string>(),
            new List<string>(),
            null,
            null,
            ""
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res?.Url);

        var expectedUrl = $"{Helpers.HttpBinUrl}/anything/allowEmptyValue?" +
                         "arrParam=&nullableStrParam=&numParam=&strParam=";
        Assert.Equal(expectedUrl, Helpers.SortQueryParameters(res.Res.Url));

        var res2 = await sdk.Parameters.AllowEmptyValueQueryParamsAsync(
            new List<string> { "a", "b" },
            new List<string> { "x", "y" },
            "nullable",
            123,
            "test"
        );

        Assert.Equal(HttpStatusCode.OK, res2.HttpMeta.Response.StatusCode);
        Assert.NotNull(res2.Res?.Url);

        var expectedUrl2 = $"{Helpers.HttpBinUrl}/anything/allowEmptyValue?" +
                          "arrParam=a&arrParam=b&arrParamOmitEmpty=x&arrParamOmitEmpty=y&" +
                          "nullableStrParam=nullable&numParam=123&strParam=test";
        Assert.Equal(expectedUrl2, Helpers.SortQueryParameters(res2.Res.Url));
    }

    [Fact]
    public async Task ParameterOpenEnum()
    {
        CommonHelpers.RecordTest("parameters-open-enum");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Parameters.ParameterOpenEnumAsync(
            paramH: OpenEnum.OneHundredAndOne,
            paramP: OpenEnum.OneHundredAndOne,
            paramQ: OpenEnum.FourHundredAndFour
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal($"{Helpers.HttpBinUrl}/anything/openEnum/101/suffix?param-q=404", res.Res.Url);
        Assert.Equal("101", res.Res.Headers["Param-H"].ToString());

        // Unrecognized open-enum values serialize the same way.
        var resUnknown = await sdk.Parameters.ParameterOpenEnumAsync(
            paramH: OpenEnum.Of(111),
            paramP: OpenEnum.Of(666),
            paramQ: OpenEnum.Of(303)
        );

        Assert.Equal(HttpStatusCode.OK, resUnknown.HttpMeta.Response.StatusCode);
        Assert.Equal($"{Helpers.HttpBinUrl}/anything/openEnum/666/suffix?param-q=303", resUnknown.Res.Url);
        Assert.Equal("111", resUnknown.Res.Headers["Param-H"].ToString());
    }

    // Reserved characters with no structural meaning inside a path segment: they
    // survive URL parsing, so the request reaches the server intact.
    private const string NonStructuralReserved = ":@!$&'()*+,;=";

    // The same reserved set prefixed with an unreserved chunk that contains a
    // space, so the wire assertions also prove an unsafe unreserved character
    // still percent-encodes (space -> %20) while reserved characters pass through.
    private const string ReservedWithSpace = "abc 123" + NonStructuralReserved;
    private const string ReservedWithSpaceEncoded = "abc%20123" + NonStructuralReserved;

    [Fact]
    public async Task PathEncoding()
    {
        CommonHelpers.RecordTest("parameters-path-encoding");

        var log = new List<CommonHelpers.RequestLogEntry>();
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, client: new CommonHelpers.RequestRecorderClient(log));

        var res = await sdk.Parameters.PathEncodingAsync(
            ReservedWithSpace,
            ReservedWithSpace
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Single(log);

        var segments = log[0]
            .RequestUri!.AbsoluteUri.Split("/anything/pathencoding/")[1]
            .Split('/');
        // param1 percent-encodes every reserved character; param2 carries them raw
        // while the space still becomes %20.
        Assert.Equal(ReservedWithSpace, Uri.UnescapeDataString(segments[0]));
        Assert.NotEqual(segments[0], segments[1]);
        Assert.Equal(ReservedWithSpaceEncoded, segments[1]);
    }

    [Fact]
    public async Task QueryEncoding()
    {
        CommonHelpers.RecordTest("parameters-query-encoding");

        var log = new List<CommonHelpers.RequestLogEntry>();
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, client: new CommonHelpers.RequestRecorderClient(log));

        var res = await sdk.Parameters.QueryEncodingAsync(ReservedWithSpace);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Single(log);
        Assert.Equal(
            Helpers.HttpBinUrl + "/anything/queryencoding?param1=" + ReservedWithSpaceEncoded,
            log[0].RequestUri!.AbsoluteUri
        );
    }

    // Two occurrences of the same key AND the same value, where only one of
    // them came from an allowReserved parameter. A (key, value) set cannot
    // tell these apart; positions can.
    [Fact]
    public void SerializeQueryParamsAppliesReservedPerOccurrence()
    {
        var queryParams = new Dictionary<string, List<string>>
        {
            ["color"] = new List<string> { "a/b", "a/b" },
        };
        var allowReserved = new Dictionary<string, HashSet<int>>
        {
            ["color"] = new HashSet<int> { 1 },
        };

        Assert.Equal(
            "color=a%2Fb&color=a/b",
            URLBuilder.SerializeQueryParams(queryParams, allowReserved)
        );
    }

    // An exploded object contributes a member named "color", and an ordinary
    // parameter of the same name contributes an identical value. Only the
    // exploded member is allowReserved.
    [Fact]
    public void SerializeQueryParamsSeparatesExplodedObjectFromOrdinaryParam()
    {
        var queryParams = new Dictionary<string, List<string>>
        {
            ["color"] = new List<string> { "x/y", "x/y" },
            ["shape"] = new List<string> { "p:q" },
        };
        var allowReserved = new Dictionary<string, HashSet<int>>
        {
            ["color"] = new HashSet<int> { 0 },
        };

        Assert.Equal(
            "color=x/y&color=x%2Fy&shape=p%3Aq",
            URLBuilder.SerializeQueryParams(queryParams, allowReserved)
        );
    }

    [Fact]
    public void SerializeQueryParamsWithoutTrackingEncodesEverything()
    {
        var queryParams = new Dictionary<string, List<string>>
        {
            ["color"] = new List<string> { "a/b" },
        };

        Assert.Equal("color=a%2Fb", URLBuilder.SerializeQueryParams(queryParams));
    }

    [Theory]
    [InlineData(":/?#[]@!$&'()*+,;=", ":/?#[]@!$&'()*+,;=")]
    [InlineData("a b", "a%20b")]
    [InlineData("100%", "100%25")]
    [InlineData("a%2Fb", "a%252Fb")]
    [InlineData("\u00e9\u2603", "%C3%A9%E2%98%83")]
    [InlineData("plain-_.~", "plain-_.~")]
    public void EscapeExceptReservedHandlesEachCharacterClass(string input, string expected)
    {
        Assert.Equal(expected, URLBuilder.EscapeExceptReserved(input));
    }
}
