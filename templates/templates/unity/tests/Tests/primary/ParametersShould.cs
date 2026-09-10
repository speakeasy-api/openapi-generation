using System.Collections.Generic;
using NUnit.Framework;
using UnityEngine.TestTools;
using Openapi;
using Openapi.Models.Operations;
using Openapi.Models.Shared;
using System.Collections;

public class ParametersShould
{
    [UnityTest]
    public IEnumerator MixedParameters()
    {
        CommonHelpers.RecordTest("parameters-mixed-primitives");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.MixedParametersPrimitivesAsync(
                    "headerValue",
                    "pathValue",
                    "queryValue"
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/mixedParams/path/pathValue?queryStringParam=queryValue"",
                    res.RawResponse.uri.ToString()
                );
                Assert.AreEqual("headerValue", res.Res.Headers.Headerparam);
                Assert.AreEqual("queryValue", res.Res.Args.QueryStringParam);
            }
        });
    }

    [UnityTest]
    public IEnumerator CamelCase()
    {
        CommonHelpers.RecordTest("parameters-camel-case");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.MixedParametersCamelCaseAsync(
                    "headerValue",
                    "pathValue",
                    "queryValue"
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/mixedParams/path/pathValue/camelcase?query_string_param=queryValue"",
                    res.RawResponse.url
                );
                Assert.AreEqual("headerValue", res.Res.Headers.HeaderParam);
                Assert.AreEqual("queryValue", res.Res.Args.QueryStringParam);
            }
        });
    }

    [UnityTest]
    public IEnumerator SimplePathParameterPrimitives()
    {
        CommonHelpers.RecordTest("parameters-simple-path-parameter-primitives");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.SimplePathParameterPrimitivesAsync(
                    true,
                    1,
                    1.1D,
                    "test"
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/pathParams/str/test/bool/true/int/1/num/1.1"",
                    res.RawResponse.uri.ToString()
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator SimplePathParameterObjects()
    {
        CommonHelpers.RecordTest("parameters-simple-path-parameter-objects");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.SimplePathParameterObjectsAsync(
                    Helpers.CreateSimpleObject(),
                    Helpers.CreateSimpleObject()
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/pathParams/obj/any,any,bool,true,date,2020-01-01,dateTime,2020-01-01T00:00:00.0000001Z,enum,four_and_more,float32,1.1,int,1,int32,1,int32Enum,55,intEnum,2,num,1.1,str,test,boolOpt,true,strOpt,testOptional/objExploded/any=any,bool=true,date=2020-01-01,dateTime=2020-01-01T00:00:00.0000001Z,enum=four_and_more,float32=1.1,int=1,int32=1,int32Enum=55,intEnum=2,num=1.1,str=test,boolOpt=true,strOpt=testOptional"",
                    res.RawResponse.uri.ToString()
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator SimplePathParameterArrays()
    {
        CommonHelpers.RecordTest("parameters-simple-path-parameter-arrays");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.SimplePathParameterArraysAsync(
                    new List<string>() { "test", "test2" }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/pathParams/arr/test,test2"",
                    res.RawResponse.uri.ToString()
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator SimplePathParameterMaps()
    {
        CommonHelpers.RecordTest("parameters-simple-path-parameter-maps");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.SimplePathParameterMapsAsync(
                    new Dictionary<string, string>() { { "test", "value" }, { "test2", "value2" } },
                    new Dictionary<string, long>() { { "test", 1 }, { "test2", 2 } }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/pathParams/map/test,value,test2,value2/mapExploded/test=1,test2=2"",
                    res.RawResponse.uri.ToString()
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator PathParameterJson()
    {
        CommonHelpers.RecordTest("parameters-path-parameter-json");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.PathParameterJsonAsync(Helpers.CreateSimpleObject())
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    $"{Helpers.HttpBinUrl}/anything/pathParams/json/{{\"any\":\"any\",\"bool\":true,\"date\":\"2020-01-01\",\"dateTime\":\"2020-01-01T00:00:00.0000001Z\",\"enum\":\"four_and_more\",\"float32\":1.1,\"int\":1,\"int32\":1,\"int32Enum\":55,\"intEnum\":2,\"num\":1.1,\"str\":\"test\",\"boolOpt\":true,\"strOpt\":\"testOptional\"}}",
                    res.Res.Url
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator FormQueryParamsPrimitive()
    {
        CommonHelpers.RecordTest("parameters-form-query-params-primitive");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.FormQueryParamsPrimitiveAsync(true, 1, 1.1D, "test")
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/queryParams/form/primitive?boolParam=true&intParam=1&numParam=1.1&strParam=test"",
                    res.RawResponse.uri.ToString()
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator FormQueryParamsObject()
    {
        CommonHelpers.RecordTest("parameters-form-query-params-object");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.FormQueryParamsObjectAsync(
                    Helpers.CreateSimpleObject(),
                    Helpers.CreateSimpleObject()
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/queryParams/form/obj?objParam=any,any,bool,true,boolOpt,true,date,2020-01-01,dateTime,2020-01-01T00:00:00.0000001Z,enum,four_and_more,float32,1.1,int,1,int32,1,int32Enum,55,intEnum,2,num,1.1,str,test,strOpt,testOptional&any=any&bool=true&boolOpt=true&date=2020-01-01&dateTime=2020-01-01T00:00:00.0000001Z&enum=four_and_more&float32=1.1&int=1&int32=1&int32Enum=55&intEnum=2&num=1.1&str=test&strOpt=testOptional"",
                    res.RawResponse.uri.ToString()
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator FormQueryParamsArray()
    {
        CommonHelpers.RecordTest("parameters-form-query-params-array");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.FormQueryParamsArrayAsync(
                    new List<string>() { "test", "test2" },
                    new List<long>() { 1, 2 }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/queryParams/form/array?arrParam=test,test2&arrParamExploded=1&arrParamExploded=2"",
                    res.RawResponse.uri.ToString()
                );
                Assert.AreEqual("test,test2", res.Res.Args.ArrParam);
                Assert.AreEqual(new List<string>() { "1", "2" }, res.Res.Args.ArrParamExploded);
            }
        });
    }

    [UnityTest]
    public IEnumerator PipeDelimitedQueryParamsArray()
    {
        CommonHelpers.RecordTest("parameters-pipe-query-params-array");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.PipeDelimitedQueryParamsArrayAsync(
                    new List<string>() { "test", "test2" },
                    new List<long> { 1, 2 },
                    new Dictionary<string, string>() { { "key1", "val1" }, { "key2", "val2" } },
                    Helpers.CreateSimpleObject()
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/queryParams/pipe/array?arrParam=test|test2&arrParamExploded=1&arrParamExploded=2&mapParam=key1|val1|key2|val2&objParam=any|any|bool|true|date|2020-01-01|dateTime|2020-01-01T00:00:00.0000001Z|enum|four_and_more|float32|1.1|int|1|int32|1|int32Enum|55|intEnum|2|num|1.1|str|test|boolOpt|true|strOpt|testOptional"",
                    res.Res.Url
                );
                Assert.AreEqual("test|test2", res.Res.Args.ArrParam);
                Assert.AreEqual(new List<string>() { "1", "2" }, res.Res.Args.ArrParamExploded);
            }
        });
    }

    [UnityTest]
    public IEnumerator FormQueryParamsMap()
    {
        CommonHelpers.RecordTest("parameters-form-query-params-map");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.FormQueryParamsMapAsync(
                    new Dictionary<string, string>() { { "test", "value" }, { "test2", "value2" } },
                    new Dictionary<string, long>() { { "test", 1 }, { "test2", 2 } }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/queryParams/form/map?mapParam=test,value,test2,value2&test=1&test2=2"",
                    res.RawResponse.uri.ToString()
                );
                Assert.AreEqual(
                    new Dictionary<string, string>()
                    {
                        { "mapParam", "test,value,test2,value2" },
                        { "test", "1" },
                        { "test2", "2" }
                    },
                    res.Res.Args
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator FormQueryParamsRefParamObject()
    {
        CommonHelpers.RecordTest("parameters-form-query-params-ref-param-object");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
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
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/queryParams/form/refParamObject?refObjParam=bool,true,int,1,num,1.1,str,test&bool=true&int=1&num=1.1&str=test"",
                    res.Res.Url
                );
                Assert.AreEqual("true", res.Res.Args.Bool);
                Assert.AreEqual("1", res.Res.Args.Int);
                Assert.AreEqual("1.1", res.Res.Args.Num);
                Assert.AreEqual("test", res.Res.Args.Str);
                Assert.AreEqual("bool,true,int,1,num,1.1,str,test", res.Res.Args.RefObjParam);
            }
        });
    }

    [UnityTest]
    public IEnumerator DeepObjectQueryParamsObject()
    {
        CommonHelpers.RecordTest("parameters-deep-object-query-params-object");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.DeepObjectQueryParamsObjectAsync(
                    Helpers.CreateSimpleObject(),
                    new ObjArrParam()
                    {
                        Arr = new List<string> { "test", "test2" }
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/queryParams/deepObject/obj?objArrParam[arr]=test&objArrParam[arr]=test2&objParam[any]=any&objParam[bool]=true&objParam[boolOpt]=true&objParam[date]=2020-01-01&objParam[dateTime]=2020-01-01T00:00:00.0000001Z&objParam[enum]=four_and_more&objParam[float32]=1.1&objParam[int]=1&objParam[int32]=1&objParam[int32Enum]=55&objParam[intEnum]=2&objParam[num]=1.1&objParam[str]=test&objParam[strOpt]=testOptional"",
                    res.RawResponse.uri.ToString()
                );
                Assert.AreEqual(new string[] { "test", "test2" }, res.Res.Args.ObjArrParamArr);
                Assert.AreEqual("any", res.Res.Args.ObjParamAny);
                Assert.AreEqual("true", res.Res.Args.ObjParamBool);
                Assert.AreEqual("true", res.Res.Args.ObjParamBoolOpt);
                Assert.AreEqual("2020-01-01", res.Res.Args.ObjParamDate);
                Assert.AreEqual("2020-01-01T00:00:00.0000001Z", res.Res.Args.ObjParamDateTime);
                Assert.AreEqual("four_and_more", res.Res.Args.ObjParamEnum);
                Assert.AreEqual("1.1", res.Res.Args.ObjParamFloat32);
                Assert.AreEqual("1", res.Res.Args.ObjParamInt32);
                Assert.AreEqual("1", res.Res.Args.ObjParamInt);
                Assert.AreEqual("1.1", res.Res.Args.ObjParamNum);
                Assert.AreEqual("test", res.Res.Args.ObjParamStr);
                Assert.AreEqual("testOptional", res.Res.Args.ObjParamStrOpt);
            }
        });
    }

    [UnityTest]
    public IEnumerator DeepObjectQueryParamsMap()
    {
        CommonHelpers.RecordTest("parameters-deep-object-query-params-map");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
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
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/queryParams/deepObject/map?mapArrParam[test]=value&mapArrParam[test]=value2&mapArrParam[test2]=value3&mapArrParam[test2]=value4&mapParam[test]=value&mapParam[test2]=value2"",
                    res.RawResponse.uri.ToString()
                );
                CommonHelpers.AssertDictsEqual(res.Res.Args, new Dictionary<string, DeepObjectQueryParamsMapArgs>()
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
        });
    }

    [UnityTest]
    public IEnumerator JsonQueryParamsObject()
    {
        CommonHelpers.RecordTest("parameters-json-query-params-object");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.JsonQueryParamsObjectAsync(
                    Helpers.CreateDeepObject(),
                    Helpers.CreateSimpleObject()
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    $"{Helpers.HttpBinUrl}/anything/queryParams/json/obj?deepObjParam={{\"any\":{{\"any\":\"any\",\"bool\":true,\"date\":\"2020-01-01\",\"dateTime\":\"2020-01-01T00:00:00.0000001Z\",\"enum\":\"four_and_more\",\"float32\":1.1,\"int\":1,\"int32\":1,\"int32Enum\":55,\"intEnum\":2,\"num\":1.1,\"str\":\"test\",\"boolOpt\":true,\"strOpt\":\"testOptional\"}},\"arr\":[{{\"any\":\"any\",\"bool\":true,\"date\":\"2020-01-01\",\"dateTime\":\"2020-01-01T00:00:00.0000001Z\",\"enum\":\"four_and_more\",\"float32\":1.1,\"int\":1,\"int32\":1,\"int32Enum\":55,\"intEnum\":2,\"num\":1.1,\"str\":\"test\",\"boolOpt\":true,\"strOpt\":\"testOptional\"}},{{\"any\":\"any\",\"bool\":true,\"date\":\"2020-01-01\",\"dateTime\":\"2020-01-01T00:00:00.0000001Z\",\"enum\":\"four_and_more\",\"float32\":1.1,\"int\":1,\"int32\":1,\"int32Enum\":55,\"intEnum\":2,\"num\":1.1,\"str\":\"test\",\"boolOpt\":true,\"strOpt\":\"testOptional\"}}],\"bool\":true,\"int\":1,\"map\":{{\"key\":{{\"any\":\"any\",\"bool\":true,\"date\":\"2020-01-01\",\"dateTime\":\"2020-01-01T00:00:00.0000001Z\",\"enum\":\"four_and_more\",\"float32\":1.1,\"int\":1,\"int32\":1,\"int32Enum\":55,\"intEnum\":2,\"num\":1.1,\"str\":\"test\",\"boolOpt\":true,\"strOpt\":\"testOptional\"}}}},\"num\":1.1,\"obj\":{{\"any\":\"any\",\"bool\":true,\"date\":\"2020-01-01\",\"dateTime\":\"2020-01-01T00:00:00.0000001Z\",\"enum\":\"four_and_more\",\"float32\":1.1,\"int\":1,\"int32\":1,\"int32Enum\":55,\"intEnum\":2,\"num\":1.1,\"str\":\"test\",\"boolOpt\":true,\"strOpt\":\"testOptional\"}},\"str\":\"test\"}}&simpleObjParam={{\"any\":\"any\",\"bool\":true,\"date\":\"2020-01-01\",\"dateTime\":\"2020-01-01T00:00:00.0000001Z\",\"enum\":\"four_and_more\",\"float32\":1.1,\"int\":1,\"int32\":1,\"int32Enum\":55,\"intEnum\":2,\"num\":1.1,\"str\":\"test\",\"boolOpt\":true,\"strOpt\":\"testOptional\"}}",
                    res.RawResponse.uri.ToString()
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator MixedQueryParams()
    {
        CommonHelpers.RecordTest("parameters-mixed-query-params");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.MixedQueryParamsAsync(
                    Helpers.CreateSimpleObject(),
                    Helpers.CreateSimpleObject(),
                    Helpers.CreateSimpleObject()
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    $"{Helpers.HttpBinUrl}/anything/queryParams/mixed?deepObjectParam[any]=any&deepObjectParam[bool]=true&deepObjectParam[date]=2020-01-01&deepObjectParam[dateTime]=2020-01-01T00:00:00.0000001Z&deepObjectParam[enum]=four_and_more&deepObjectParam[float32]=1.1&deepObjectParam[int]=1&deepObjectParam[int32]=1&deepObjectParam[int32Enum]=55&deepObjectParam[intEnum]=2&deepObjectParam[num]=1.1&deepObjectParam[str]=test&deepObjectParam[boolOpt]=true&deepObjectParam[strOpt]=testOptional&any=any&bool=true&date=2020-01-01&dateTime=2020-01-01T00:00:00.0000001Z&enum=four_and_more&float32=1.1&int=1&int32=1&int32Enum=55&intEnum=2&num=1.1&str=test&boolOpt=true&strOpt=testOptional&jsonParam={{\"any\":\"any\",\"bool\":true,\"date\":\"2020-01-01\",\"dateTime\":\"2020-01-01T00:00:00.0000001Z\",\"enum\":\"four_and_more\",\"float32\":1.1,\"int\":1,\"int32\":1,\"int32Enum\":55,\"intEnum\":2,\"num\":1.1,\"str\":\"test\",\"boolOpt\":true,\"strOpt\":\"testOptional\"}}",
                    res.RawResponse.uri.ToString()
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator HeaderParamsPrimitive()
    {
        CommonHelpers.RecordTest("parameters-header-params-primitive");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (var res = await sdk.Parameters.HeaderParamsPrimitiveAsync(true, 1, 1.1D, "test"))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("true", res.Res.Headers.XHeaderBoolean);
                Assert.AreEqual("1", res.Res.Headers.XHeaderInteger);
                Assert.AreEqual("1.1", res.Res.Headers.XHeaderNumber);
                Assert.AreEqual("test", res.Res.Headers.XHeaderString);
            }
        });
    }

    [UnityTest]
    public IEnumerator HeaderParamsObject()
    {
        CommonHelpers.RecordTest("parameters-header-params-object");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.HeaderParamsObjectAsync(
                    Helpers.CreateSimpleObject(),
                    Helpers.CreateSimpleObject()
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    "any,any,bool,true,date,2020-01-01,dateTime,2020-01-01T00:00:00.0000001Z,enum,four_and_more,float32,1.1,int,1,int32,1,int32Enum,55,intEnum,2,num,1.1,str,test,boolOpt,true,strOpt,testOptional",
                    res.Res.Headers.XHeaderObj
                );
                Assert.AreEqual(
                    "any=any,bool=true,date=2020-01-01,dateTime=2020-01-01T00:00:00.0000001Z,enum=four_and_more,float32=1.1,int=1,int32=1,int32Enum=55,intEnum=2,num=1.1,str=test,boolOpt=true,strOpt=testOptional",
                    res.Res.Headers.XHeaderObjExplode
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator HeaderParamsMap()
    {
        CommonHelpers.RecordTest("parameters-header-params-map");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.HeaderParamsMapAsync(
                    new Dictionary<string, string>() { { "key1", "value1" }, { "key2", "value2" } },
                    new Dictionary<string, string>() { { "test1", "val1" }, { "test2", "val2" } }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("key1,value1,key2,value2", res.Res.Headers.XHeaderMap);
                Assert.AreEqual("test1=val1,test2=val2", res.Res.Headers.XHeaderMapExplode);
            }
        });
    }

    [UnityTest]
    public IEnumerator HeaderParamsArray()
    {
        CommonHelpers.RecordTest("parameters-header-params-array");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Parameters.HeaderParamsArrayAsync(
                    new List<string> { "test1", "test2" }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("test1,test2", res.Res.Headers.XHeaderArray);
            }
        });
    }
}
