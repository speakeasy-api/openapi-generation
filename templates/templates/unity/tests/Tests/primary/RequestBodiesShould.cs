using NUnit.Framework;
using System;
using System.Collections;
using System.Collections.Generic;
using System.Linq;
using System.Numerics;
using System.Text;
using System.Text.RegularExpressions;
using UnityEngine.TestTools;
using Openapi;
using Openapi.Models.Operations;
using Openapi.Models.Shared;
using Openapi.Utils;

public class RequestBodiesShould
{
    [UnityTest]
    public IEnumerator PostApplicationJsonSimple()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-simple");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonSimpleAsync(
                    Helpers.CreateSimpleObject()
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Helpers.AssertSimpleObject(res.Res.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonArray()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayAsync(
                    new List<SimpleObject>()
                    {
                        Helpers.CreateSimpleObject(),
                        Helpers.CreateSimpleObject()
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(2, res.Res.Count());
                foreach (var obj in res.Res)
                {
                    Helpers.AssertSimpleObject(obj);
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonArrayOfArray()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array-of-array");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var obj = Helpers.CreateSimpleObject();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayOfArrayAsync(
                    new List<List<SimpleObject>>
                    {
                        new List<SimpleObject>() { obj, obj },
                        new List<SimpleObject>() { obj, obj }
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(2, res.Res.Count());
                Assert.AreEqual(2, res.Res.ToList()[0].Count());
                Assert.AreEqual(2, res.Res.ToList()[1].Count());

                for (var i = 0; i < 2; i++)
                {
                    for (var j = 0; j < 2; j++)
                    {
                        Helpers.AssertSimpleObject(res.Res.ToList()[i].ToList()[j]);
                    }
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonMap()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var obj = Helpers.CreateSimpleObject();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapAsync(
                    new Dictionary<string, SimpleObject>()
                    {
                        { "mapElem1", obj },
                        { "mapElem2", obj }
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(2, res.Res.Count());
                Helpers.AssertSimpleObject(res.Res["mapElem1"]);
                Helpers.AssertSimpleObject(res.Res["mapElem2"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonMapOfMap()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-of-map");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var obj = Helpers.CreateSimpleObject();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapOfMapAsync(
                    new Dictionary<string, Dictionary<string, SimpleObject>>()
                    {
                        {
                            "mapElem1",
                            new Dictionary<string, SimpleObject>()
                            {
                                { "subMapElem1", obj },
                                { "subMapElem2", obj }
                            }
                        },
                        {
                            "mapElem2",
                            new Dictionary<string, SimpleObject>()
                            {
                                { "subMapElem1", obj },
                                { "subMapElem2", obj }
                            }
                        },
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(2, res.Res.Count());
                Assert.AreEqual(2, res.Res["mapElem1"].Count());
                Assert.AreEqual(2, res.Res["mapElem2"].Count());
                Helpers.AssertSimpleObject(res.Res["mapElem1"]["subMapElem1"]);
                Helpers.AssertSimpleObject(res.Res["mapElem1"]["subMapElem2"]);
                Helpers.AssertSimpleObject(res.Res["mapElem2"]["subMapElem1"]);
                Helpers.AssertSimpleObject(res.Res["mapElem2"]["subMapElem2"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonMapOfAny()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-of-any");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapOfAnyAsync(
                    new Dictionary<string, dynamic>()
                    {
                        {
                            "mapElemArray",
                            new List<string>() { "array-test" }
                        },
                        {
                            "mapElemBool",
                            true
                        },
                        {
                            "mapElemNumber",
                            123
                        },
                        {
                            "mapElemObject",
                            new { property = "property-test" }
                        },
                        {
                            "mapElemString",
                            "string-test"
                        }
                    }
                )
            )
            {
                Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
                Assert.Equal(5, res.Res.Count());
                Assert.Collection(
                    ((IEnumerable<object>)res.Res["mapElemArray"]).Cast<string>(),
                    item => Assert.Equal("array-test", item)
                );
                Assert.Equal(true, res.Res["mapElemBool"]);
                Assert.Equal(123L, res.Res["mapElemNumber"]);
                Assert.Equal("property-test", ((Dictionary<string, object>)res.Res["mapElemObject"])["property"]);
                Assert.Equal("string-test", res.Res["mapElemString"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonMapOfArray()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-of-array");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var obj = Helpers.CreateSimpleObject();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapOfArrayAsync(
                    new Dictionary<string, List<SimpleObject>>()
                    {
                        {
                            "mapElem1",
                            new List<SimpleObject>() { obj, obj }
                        },
                        {
                            "mapElem2",
                            new List<SimpleObject>() { obj, obj }
                        }
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(2, res.Res.Count());
                Assert.AreEqual(2, res.Res["mapElem1"].Count());
                Assert.AreEqual(2, res.Res["mapElem2"].Count());
                Helpers.AssertSimpleObject(res.Res["mapElem1"].First());
                Helpers.AssertSimpleObject(res.Res["mapElem1"].Last());
                Helpers.AssertSimpleObject(res.Res["mapElem2"].First());
                Helpers.AssertSimpleObject(res.Res["mapElem2"].Last());
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonArrayOfMap()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array-of-map");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var maps = new List<Dictionary<string, SimpleObject>>();
            for (int i = 0; i < 2; i++)
            {
                maps.Add(
                    new Dictionary<string, SimpleObject>()
                    {
                        { "mapElem1", Helpers.CreateSimpleObject() },
                        { "mapElem2", Helpers.CreateSimpleObject() }
                    }
                );
            }

            using (
                var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayOfMapAsync(
                    maps
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(2, res.Res.Count());
                Assert.AreEqual(2, res.Res.ToList()[0].Count());
                Assert.AreEqual(2, res.Res.ToList()[1].Count());
                Helpers.AssertSimpleObject(res.Res.ToList()[0]["mapElem1"]);
                Helpers.AssertSimpleObject(res.Res.ToList()[0]["mapElem2"]);
                Helpers.AssertSimpleObject(res.Res.ToList()[1]["mapElem1"]);
                Helpers.AssertSimpleObject(res.Res.ToList()[1]["mapElem2"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonMapOfPrimitive()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-of-primitive");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapOfPrimitiveAsync(
                    new Dictionary<string, string>()
                    {
                        { "mapElem1", "hello" },
                        { "mapElem2", "world" }
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(2, res.Res.Count());
                Assert.AreEqual("hello", res.Res["mapElem1"]);
                Assert.AreEqual("world", res.Res["mapElem2"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonArrayOfPrimitive()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array-of-primitive");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res =
                    await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayOfPrimitiveAsync(
                        new List<string>() { "hello", "world" }
                    )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(2, res.Res.Count());
                Assert.AreEqual("hello", res.Res.ToList()[0]);
                Assert.AreEqual("world", res.Res.ToList()[1]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonMapOfMapOfPrimitive()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-of-map-of-primitive");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res =
                    await sdk.RequestBodies.RequestBodyPostApplicationJsonMapOfMapOfPrimitiveAsync(
                        new Dictionary<string, Dictionary<string, string>>()
                        {
                            {
                                "mapElem1",
                                new Dictionary<string, string>()
                                {
                                    { "subMapElem1", "foo" },
                                    { "subMapElem2", "bar" }
                                }
                            },
                            {
                                "mapElem2",
                                new Dictionary<string, string>()
                                {
                                    { "subMapElem1", "buzz" },
                                    { "subMapElem2", "bazz" }
                                }
                            }
                        }
                    )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(2, res.Res.Count());
                Assert.AreEqual(2, res.Res["mapElem1"].Count());
                Assert.AreEqual(2, res.Res["mapElem2"].Count());
                Assert.AreEqual("foo", res.Res["mapElem1"]["subMapElem1"]);
                Assert.AreEqual("bar", res.Res["mapElem1"]["subMapElem2"]);
                Assert.AreEqual("buzz", res.Res["mapElem2"]["subMapElem1"]);
                Assert.AreEqual("bazz", res.Res["mapElem2"]["subMapElem2"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonArrayOfArrayOfPrimitive()
    {
        CommonHelpers.RecordTest(
            "request-bodies-post-application-json-array-of-array-of-primitive"
        );

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res =
                    await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayOfArrayOfPrimitiveAsync(
                        new List<List<string>>()
                        {
                            new List<string>() { "foo", "bar" },
                            new List<string>() { "buzz", "bazz" }
                        }
                    )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(2, res.Res.Count());
                Assert.AreEqual(2, res.Res.First().Count());
                Assert.AreEqual(2, res.Res.Last().Count());
                Assert.AreEqual("foo", res.Res.First().First());
                Assert.AreEqual("bar", res.Res.First().Last());
                Assert.AreEqual("buzz", res.Res.Last().First());
                Assert.AreEqual("bazz", res.Res.Last().Last());
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonArrayObject()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array-object");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var obj = Helpers.CreateSimpleObject();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayObjAsync(
                    new List<SimpleObject>() { obj, obj }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(2, res.ArrObjValue.Json.Count());
                Helpers.AssertSimpleObject(res.ArrObjValue.Json.ToList()[0]);
                Helpers.AssertSimpleObject(res.ArrObjValue.Json.ToList()[1]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonMapObject()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-object");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var obj = Helpers.CreateSimpleObject();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapObjAsync(
                    new Dictionary<string, SimpleObject>()
                    {
                        { "mapElem1", obj },
                        { "mapElem2", obj }
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(2, res.MapObjValue.Json.Count());
                Helpers.AssertSimpleObject(res.MapObjValue.Json["mapElem1"]);
                Helpers.AssertSimpleObject(res.MapObjValue.Json["mapElem2"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonDeep()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-deep");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonDeepAsync(
                    Helpers.CreateDeepObject()
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Helpers.AssertDeepObject(res.Res.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostApplicationJsonMultipleJsonFiltered()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-multiple-json-filtered");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res =
                    await sdk.RequestBodies.RequestBodyPostApplicationJsonMultipleJsonFilteredAsync(
                        Helpers.CreateSimpleObject()
                    )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Helpers.AssertSimpleObject(res.Res.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostMultipleContentTypesComponentFiltered()
    {
        CommonHelpers.RecordTest("request-bodies-post-multiple-content-types-component-filtered-application-json");
        CommonHelpers.RecordTest("request-bodies-post-multiple-content-types-component-filtered-multipart-form-data");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res =
                    await sdk.RequestBodies.RequestBodyPostMultipleContentTypesComponentFilteredAsync(
                        Helpers.CreateSimpleObject()
                    )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Helpers.AssertSimpleObject(res.Res.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostMultipleContentTypesInlineFiltered()
    {
        CommonHelpers.RecordTest("request-bodies-post-multiple-content-types-inline-filtered");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res =
                    await sdk.RequestBodies.RequestBodyPostMultipleContentTypesInlineFilteredAsync(
                        new RequestBodyPostMultipleContentTypesInlineFilteredRequestBody()
                        {
                            Bool = true,
                            Num = 1.1F,
                            Str = "test"
                        }
                    )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(3, res.Res.Json.Count());
                Assert.AreEqual(true, res.Res.Json["bool"]);
                Assert.That(1.1, Is.EqualTo((double)res.Res.Json["num"]).Within(.01).Percent);
                Assert.AreEqual("test", res.Res.Json["str"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostMultipleContentTypeSplitJson()
    {
        CommonHelpers.RecordTest("request-bodies-post-multiple-content-types-split-json");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitJsonAsync(
                    new RequestBodyPostMultipleContentTypesSplitJsonRequestBody()
                    {
                        Bool = true,
                        Num = 1.1F,
                        Str = "test"
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(true, res.Res.Json["bool"]);
                Assert.That(1.1, Is.EqualTo((double)res.Res.Json["num"]).Within(.01).Percent);
                Assert.AreEqual("test", res.Res.Json["str"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostMutlipleContentTypesSplitMultipart()
    {
        CommonHelpers.RecordTest("request-bodies-post-multiple-content-types-split-multipart");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res =
                    await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitMultipartAsync(
                        new RequestBodyPostMultipleContentTypesSplitMultipartRequestBody()
                        {
                            Bool2 = true,
                            Num2 = 1.1D,
                            Str2 = "test"
                        }
                    )
            )
            {
                Assert.AreEqual(200, res.StatusCode);

                Assert.AreEqual("true", res.Res.Form["bool2"]);
                Assert.AreEqual("1.1", res.Res.Form["num2"]);
                Assert.AreEqual("test", res.Res.Form["str2"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostMultipleContentTypesSplitForm()
    {
        CommonHelpers.RecordTest("request-bodies-post-multiple-content-types-split-form");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitFormAsync(
                    new RequestBodyPostMultipleContentTypesSplitFormRequestBody()
                    {
                        Bool3 = true,
                        Num3 = 1.1D,
                        Str3 = "test"
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("true", res.Res.Form["bool3"]);
                Assert.AreEqual("1.1", res.Res.Form["num3"]);
                Assert.AreEqual("test", res.Res.Form["str3"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostMultipleContentTypesSplitJsonWithParam()
    {
        CommonHelpers.RecordTest(
            "request-bodies-post-multiple-content-types-split-json-with-param"
        );

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var requestBody = new RequestBodyPostMultipleContentTypesSplitParamJsonRequestBody()
            {
                Bool = true,
                Num = 1.1D,
                Str = "test body"
            };

            using (
                var res =
                    await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitParamJsonAsync(
                        requestBody,
                        "test param"
                    )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.True((bool)res.Res.Json["bool"]);
                Assert.AreEqual(1.1, (double)res.Res.Json["num"]);
                Assert.AreEqual("test body", res.Res.Json["str"].ToString());
                Assert.AreEqual("test param", res.Res.Args["paramStr"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostMultipleContentTypesSplitMultiplartWithParam()
    {
        CommonHelpers.RecordTest(
            "request-bodies-post-multiple-content-types-split-multipart-with-param"
        );

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var formData = new RequestBodyPostMultipleContentTypesSplitParamMultipartRequestBody()
            {
                Bool2 = true,
                Num2 = 1.1D,
                Str2 = "test body"
            };

            using (
                var res =
                    await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitParamMultipartAsync(
                        formData,
                        "test param"
                    )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual("true", res.Res.Form["bool2"]);
                Assert.AreEqual("1.1", res.Res.Form["num2"]);
                Assert.AreEqual("test body", res.Res.Form["str2"]);
                Assert.AreEqual("test param", res.Res.Args["paramStr"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostMultipleContentTypesSplitFormWithParam()
    {
        CommonHelpers.RecordTest(
            "request-bodies-post-multiple-content-types-split-form-with-param"
        );

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var requestBody = new RequestBodyPostMultipleContentTypesSplitParamFormRequestBody()
            {
                Bool3 = true,
                Num3 = 1.1D,
                Str3 = "test body"
            };

            using (
                var res =
                    await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitParamFormAsync(
                        requestBody,
                        "test param"
                    )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual("true", res.Res.Form["bool3"]);
                Assert.AreEqual("1.1", res.Res.Form["num3"]);
                Assert.AreEqual("test body", res.Res.Form["str3"]);
                Assert.AreEqual("test param", res.Res.Args["paramStr"]);
            }
        });
    }

    [UnityTest]
    public IEnumerator PutMultipartSimple()
    {
        CommonHelpers.RecordTest("request-bodies-put-multipart-simple");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPutMultipartSimpleAsync(
                    Helpers.CreateSimpleObject()
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("any", res.Res.Form.Any);
                Assert.AreEqual("true", res.Res.Form.Bool);
                Assert.AreEqual("true", res.Res.Form.BoolOpt);
                Assert.AreEqual("2020-01-01", res.Res.Form.Date);
                Assert.AreEqual("2020-01-01T00:00:00.0000001Z", res.Res.Form.DateTime);
                Assert.AreEqual("four_and_more", res.Res.Form.Enum);
                Assert.AreEqual("1.1", res.Res.Form.Float32);
                Assert.AreEqual("1", res.Res.Form.Int);
                Assert.AreEqual("1", res.Res.Form.Int32);
                Assert.AreEqual("1.1", res.Res.Form.Num);
                Assert.AreEqual("test", res.Res.Form.Str);
                Assert.AreEqual("testOptional", res.Res.Form.StrOpt);
            }
        });
    }

    [UnityTest]
    public IEnumerator PutMultipartDeep()
    {
        CommonHelpers.RecordTest("request-bodies-put-multipart-deep");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var obj = Helpers.CreateDeepObject();

            using (var res = await sdk.RequestBodies.RequestBodyPutMultipartDeepAsync(obj))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(Utilities.ToString(obj.Arr), res.Res.Form.Arr);
                Assert.AreEqual("true", res.Res.Form.Bool);
                Assert.AreEqual("1", res.Res.Form.Int);
                Assert.AreEqual(Utilities.ToString(obj.Map), res.Res.Form.Map);
                Assert.AreEqual("1.1", res.Res.Form.Num);
                Assert.AreEqual(Utilities.ToString(obj.Obj), res.Res.Form.Obj);
                Assert.AreEqual("test", res.Res.Form.Str);
            }
        });
    }

    [UnityTest]
    public IEnumerator PutMultipartFile()
    {
        CommonHelpers.RecordTest("request-bodies-put-multipart-file");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var data = Helpers.GetData();

            using (
                var res = await sdk.RequestBodies.RequestBodyPutMultipartFileAsync(
                    new RequestBodyPutMultipartFileRequestBody()
                    {
                        File = new File() { Content = data, FileName = "testUpload.json" }
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(
                    Encoding.UTF8.GetString(data, 0, data.Length),
                    res.Res.Files["file"]
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator PutMultipartFileRef()
    {
        CommonHelpers.RecordTest("request-bodies-put-multipart-file-ref");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var data = Helpers.GetData();

            using (
                var res = await sdk.RequestBodies.RequestBodyPutMultipartFileRefAsync(
                    new RequestBodyPutMultipartFileRefRequestBody()
                    {
                        File = new BinaryString() { Content = data, FileName = "testUpload.json" }
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(
                    Encoding.UTF8.GetString(data, 0, data.Length),
                    res.Res.Files["file"]
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator PostFormSimple()
    {
        CommonHelpers.RecordTest("request-bodies-post-form-simple");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostFormSimpleAsync(
                    Helpers.CreateSimpleObject()
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual("any", res.Res.Form.Any);
                Assert.AreEqual("true", res.Res.Form.Bool);
                Assert.AreEqual("true", res.Res.Form.BoolOpt);
                Assert.AreEqual("2020-01-01", res.Res.Form.Date);
                Assert.AreEqual("2020-01-01T00:00:00.0000001Z", res.Res.Form.DateTime);
                Assert.AreEqual("four_and_more", res.Res.Form.Enum);
                Assert.AreEqual("1.1", res.Res.Form.Float32);
                Assert.AreEqual("1", res.Res.Form.Int);
                Assert.AreEqual("1", res.Res.Form.Int32);
                Assert.AreEqual("1.1", res.Res.Form.Num);
                Assert.AreEqual("test", res.Res.Form.Str);
                Assert.AreEqual("testOptional", res.Res.Form.StrOpt);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostFormDeep()
    {
        CommonHelpers.RecordTest("request-bodies-post-form-deep");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var obj = Helpers.CreateDeepObject();

            using (var res = await sdk.RequestBodies.RequestBodyPostFormDeepAsync(obj))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(Utilities.ToString(obj.Arr), res.Res.Form.Arr);
                Assert.AreEqual("true", res.Res.Form.Bool);
                Assert.AreEqual("1", res.Res.Form.Int);
                Assert.AreEqual(Utilities.ToString(obj.Map), res.Res.Form.Map);
                Assert.AreEqual("1.1", res.Res.Form.Num);
                Assert.AreEqual(Utilities.ToString(obj.Obj), res.Res.Form.Obj);
                Assert.AreEqual("test", res.Res.Form.Str);
            }
        });
    }

    [UnityTest]
    public IEnumerator PostFormMapPrimitive()
    {
        CommonHelpers.RecordTest("request-bodies-post-form-map-primitive");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var map = new Dictionary<string, string>()
            {
                { "key1", "value1" },
                { "key2", "value2" },
                { "key3", "value3" }
            };

            using (var res = await sdk.RequestBodies.RequestBodyPostFormMapPrimitiveAsync(map))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(map, res.Res.Form);
            }
        });
    }

    [UnityTest]
    public IEnumerator PutString()
    {
        CommonHelpers.RecordTest("request-bodies-put-string");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var str = "Hello world";

            using (var res = await sdk.RequestBodies.RequestBodyPutStringAsync(str))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(str, res.Res.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator PutBytes()
    {
        CommonHelpers.RecordTest("request-bodies-put-bytes");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var data = Helpers.GetData();

            using (var res = await sdk.RequestBodies.RequestBodyPutBytesAsync(data))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(Encoding.UTF8.GetString(data, 0, data.Length), res.Res.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator PutStringWithParams()
    {
        CommonHelpers.RecordTest("request-bodies-put-string-with-params");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPutStringWithParamsAsync(
                    "Hello world",
                    "test param"
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("Hello world", res.Res.Data);
                Assert.AreEqual("test param", res.Res.Args.QueryStringParam);
            }
        });
    }

    [UnityTest]
    public IEnumerator PutBytesWithParams()
    {
        CommonHelpers.RecordTest("request-bodies-put-bytes-with-params");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var data = Helpers.GetData();

            using (
                var res = await sdk.RequestBodies.RequestBodyPutBytesWithParamsAsync(
                    data,
                    "test param"
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(Encoding.UTF8.GetString(data, 0, data.Length), res.Res.Data);
                Assert.AreEqual("test param", res.Res.Args.QueryStringParam);
            }
        });
    }

    [UnityTest]
    public IEnumerator EmptyObject()
    {
        CommonHelpers.RecordTest("request-bodies-post-empty-object");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostEmptyObjectAsync(
                    new RequestBodyPostEmptyObjectRequestBody()
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator CamelCase()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-simple-camel-case");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res =
                    await sdk.RequestBodies.RequestBodyPostApplicationJsonSimpleCamelCaseAsync(
                        Helpers.CreateSimpleObjectCamelCase()
                    )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Helpers.AssertSimpleObjectCamelCase(res.Res.Json);

                Assert.AreEqual(
                    28,
                    Regex.Matches(res.RawResponse.downloadHandler.text, "_val").Count
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyReadOnlyInput()
    {
        CommonHelpers.RecordTest("request-bodies-read-only-input");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyReadOnlyInputAsync(
                    new ReadOnlyObjectInput()
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.True(res.ReadOnlyObject.Bool);
                Assert.AreEqual(1.0, res.ReadOnlyObject.Num);
                Assert.AreEqual("hello", res.ReadOnlyObject.String);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyWriteOnlyOutput()
    {
        CommonHelpers.RecordTest("request-bodies-write-only-output");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyWriteOnlyOutputAsync(
                    new WriteOnlyObject()
                    {
                        Bool = true,
                        Num = 1.0F,
                        String = "hello"
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyWriteOnly()
    {
        CommonHelpers.RecordTest("request-bodies-write-only");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyWriteOnlyAsync(
                    new WriteOnlyObject()
                    {
                        Bool = true,
                        Num = 1.0F,
                        String = "hello"
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.True(res.ReadOnlyObject.Bool);
                Assert.AreEqual(1.0, res.ReadOnlyObject.Num);
                Assert.AreEqual("hello", res.ReadOnlyObject.String);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyReadAndWrite()
    {
        CommonHelpers.RecordTest("request-bodies-read-and-write");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyReadAndWriteAsync(
                    new ReadWriteObject()
                    {
                        Num1 = 1,
                        Num2 = 2,
                        Num3 = 4,
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(4, res.ReadWriteObject.Num3);
                Assert.AreEqual(7, res.ReadWriteObject.Sum);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyReadOnly()
    {
        CommonHelpers.RecordTest("request-bodies-complex-number-types");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var req = new RequestBodyPostComplexNumberTypesRequest()
            {
                ComplexNumberTypes = new ComplexNumberTypes()
                {
                    Bigint = BigInteger.Parse("9007199254740991"),
                    BigintStr = BigInteger.Parse("9223372036854775807"),
                    Decimal = 3.141592653589793M,
                    DecimalStr = 3.141592653589793238462643383279M
                },
                PathBigInt = BigInteger.Parse("9007199254740991"),
                PathBigIntStr = BigInteger.Parse("9223372036854775807"),
                PathDecimal = 3.141592653589793M,
                PathDecimalStr = 3.141592653589793238462643383279M,
                QueryBigInt = BigInteger.Parse("9007199254740991"),
                QueryBigIntStr = BigInteger.Parse("9223372036854775807"),
                QueryDecimal = 3.141592653589793M,
                QueryDecimalStr = 3.141592653589793238462643383279M
            };

            using (var res = await sdk.RequestBodies.RequestBodyPostComplexNumberTypesAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(
                    req.ComplexNumberTypes.Bigint,
                    res.Object.Json.Bigint
                );
                Assert.AreEqual(
                    req.ComplexNumberTypes.BigintStr,
                    res.Object.Json.BigintStr
                );
                Assert.AreEqual(
                    req.ComplexNumberTypes.Decimal,
                    res.Object.Json.Decimal
                );
                Assert.AreEqual(
                    req.ComplexNumberTypes.DecimalStr,
                    res.Object.Json.DecimalStr
                );
                Assert.AreEqual(
                    "$"{Helpers.HttpBinUrl}/anything/requestBodies/post/9007199254740991/9223372036854775807/3.141592653589793/3.1415926535897932384626433833/complex-number-types?queryBigInt=9007199254740991&queryBigIntStr=9223372036854775807&queryDecimal=3.141592653589793&queryDecimalStr=3.1415926535897932384626433833"",
                    res.Object.Url
                );
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesString()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-string");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesStringAsync("test")
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("test", res.Object.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesInteger()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-integer");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesIntegerAsync(1)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(1, res.Object.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesInt32()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-int32");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesInt32Async(1)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(1, res.Object.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesBigInt()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-bigint");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesBigIntAsync(
                    new BigInteger(1)
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(typeof(BigInteger), res.Object.Json.GetType());
                Assert.AreEqual(new BigInteger(1), res.Object.Json);
                Assert.AreEqual("1", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesBigIntStr()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-bigint-str");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = BigInteger.Parse("1");
            var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesBigIntStrAsync(req);

            Assert.AreEqual(200, res.StatusCode);
            Assert.AreEqual(typeof(BigInteger), res.Object.Json.GetType());
            Assert.AreEqual(req, res.Object.Json);
            Assert.AreEqual("\"1\"", res.Object.Data);
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesNumber()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-number");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesNumberAsync(1.1)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(1.1, res.Object.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesFloat()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-float32");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesFloat32Async(1.1)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(1.1, res.Object.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesDecimal()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-decimal");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesDecimalAsync(1.1M)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(typeof(decimal), res.Object.Json.GetType());
                Assert.AreEqual(1.1M, res.Object.Json);
                Assert.AreEqual("1.1", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesDecimalStr()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-decimal-str");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesDecimalStrAsync(1.1M)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(typeof(decimal), res.Object.Json.GetType());
                Assert.AreEqual(1.1M, res.Object.Json);
                Assert.AreEqual("\"1.1\"", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesBoolean()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-boolean");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesBooleanAsync(true)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(true, res.Object.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesDate()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-date");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var date = DateOnly.FromDateTime(new DateTime(2020, 1, 1));

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesDateAsync(date)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(date, res.Object.Json);
                Assert.AreEqual("\"2020-01-01\"", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesDateTime()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-date-time");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var dateTime = new DateTime(2020, 1, 1, 0, 0, 0, DateTimeKind.Utc);

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesDateTimeAsync(dateTime)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(dateTime, res.Object.Json);
                Assert.AreEqual("\"2020-01-01T00:00:00.0000000Z\"", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesMapDateTime()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-map-date-time");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = new Dictionary<string, DateTime>()
          {
                { "test", new DateTime(2020, 1, 1, 0, 0, 0, DateTimeKind.Utc) }
          };

            using (
                  var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesMapDateTimeAsync(req)
              )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(req, res.Object.Json);
                Assert.AreEqual("{\"test\":\"2020-01-01T00:00:00.0000000Z\"}", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesMapBigIntStr()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-map-bigint-str");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = new Dictionary<string, BigInteger>()
            {
                { "test", new BigInteger(1) }
            };

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesMapBigIntStrAsync(req)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(req, res.Object.Json);
                Assert.AreEqual("{\"test\":\"1\"}", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesMapDecimal()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-map-decimal");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = new Dictionary<string, Decimal>()
            {
                { "test", 3.141592653589793M }
            };

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesMapDecimalAsync(req)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(req, res.Object.Json);
                Assert.AreEqual("{\"test\":3.141592653589793}", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesArrayDate()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-array-date");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = new List<DateOnly> { DateOnly.FromDateTime(new DateTime(2020, 1, 1)) };

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesArrayDateAsync(req)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(req, res.Object.Json);
                Assert.AreEqual("[\"2020-01-01\"]", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesArrayBigInt()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-array-bigint");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = new List<BigInteger> { new BigInteger(1) };

            using (
                  var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesArrayBigIntAsync(req)
              )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(req, res.Object.Json);
                Assert.AreEqual("[1]", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesArrayDecimalStr()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-array-decimal-str");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = new List<Decimal> { 3.141592653589793438462643383279M };

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesArrayDecimalStrAsync(req)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(req, res.Object.Json);
                Assert.AreEqual("[\"3.1415926535897934384626433833\"]", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesComplexNumberArrays()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-complex-number-arrays");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = new ComplexNumberArrays()
            {
                BigintArray = new List<BigInteger> { BigInteger.Parse("9007199254740991") },
                BigintStrArray = new List<BigInteger> { BigInteger.Parse("9223372036854775807") },
                DecimalArray = new List<Decimal> { 3.141592653589793M },
                DecimalStrArray = new List<Decimal> { 3.141592653589793238462643383279M }
            };

            var json = Helpers.GetSerializedBodyJson(req);
            Assert.AreEqual("{\"bigintArray\":[9007199254740991],\"bigintStrArray\":[\"9223372036854775807\"],\"decimalArray\":[3.141592653589793],\"decimalStrArray\":[\"3.1415926535897932384626433833\"]}", json);

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesComplexNumberArraysAsync(req)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(req.BigintArray, res.Res.Json.BigintArray);
                Assert.AreEqual(req.BigintStrArray, res.Res.Json.BigintStrArray);
                Assert.AreEqual(req.DecimalArray, res.Res.Json.DecimalArray);
                Assert.AreEqual(req.DecimalStrArray, res.Res.Json.DecimalStrArray);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostJsonDataTypesComplexNumberMaps()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-complex-number-maps");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();
            var req = new ComplexNumberMaps()
            {
                BigintMap = new Dictionary<string, BigInteger> { { "bigint", BigInteger.Parse("9007199254740991") } },
                BigintStrMap = new Dictionary<string, BigInteger> { { "bigintStr", BigInteger.Parse("9223372036854775807") } },
                DecimalMap = new Dictionary<string, Decimal> { { "decimal", 3.141592653589793M } },
                DecimalStrMap = new Dictionary<string, Decimal> { { "decimalStr", 3.141592653589793238462643383279M } }
            };

            var json = Helpers.GetSerializedBodyJson(req);
            Assert.AreEqual("{\"bigintMap\":{\"bigint\":9007199254740991},\"bigintStrMap\":{\"bigintStr\":\"9223372036854775807\"},\"decimalMap\":{\"decimal\":3.141592653589793},\"decimalStrMap\":{\"decimalStr\":\"3.1415926535897932384626433833\"}}", json);

            using (
                var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesComplexNumberMapsAsync(req)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(req.BigintMap, res.Res.Json.BigintMap);
                Assert.AreEqual(req.BigintStrMap, res.Res.Json.BigintStrMap);
                Assert.AreEqual(req.DecimalMap, res.Res.Json.DecimalMap);
                Assert.AreEqual(req.DecimalStrMap, res.Res.Json.DecimalStrMap);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostNullableRequiredStringBody()
    {
        CommonHelpers.RecordTest("request-bodies-post-nullable-required-string-body");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostNullableRequiredStringBodyAsync(null)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("null", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostNullableNotRequiredStringBody()
    {
        CommonHelpers.RecordTest("request-bodies-post-nullable-not-required-string-body");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostNullableNotRequiredStringBodyAsync(null)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("null", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostNotNullableNotRequiredStringBody()
    {
        CommonHelpers.RecordTest("request-bodies-post-not-nullable-not-required-string-body");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.RequestBodies.RequestBodyPostNotNullableNotRequiredStringBodyAsync(null)
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("", res.Object.Data);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostNullableRequiredProperty()
    {
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var tests = new CommonHelpers.TestTableEntry[] {
                new CommonHelpers.TestTableEntry()
                {
                    name = "Empty initializer",
                    arg = new NullableRequiredPropertyPostRequestBody{},
                    want = "{\"NullableRequiredArray\":null,\"NullableRequiredBigIntStr\":null,\"NullableRequiredDateTime\":null,\"NullableRequiredDecimalStr\":null,\"NullableRequiredEnum\":null,\"NullableRequiredInt\":null}",
                testId = "request-bodies-post-nullable-optional-fields"
                },
                new CommonHelpers.TestTableEntry()
                {
                    name = "All fields set to null",
                    arg = new NullableRequiredPropertyPostRequestBody
                    {
                        NullableOptionalInt = null,
                        NullableRequiredArray = null,
                        NullableRequiredEnum = null,
                        NullableRequiredInt = null,
                        NullableRequiredDateTime = null,
                        NullableRequiredBigIntStr = null,
                        NullableRequiredDecimalStr = null
                    },
                    want = "{\"NullableRequiredArray\":null,\"NullableRequiredBigIntStr\":null,\"NullableRequiredDateTime\":null,\"NullableRequiredDecimalStr\":null,\"NullableRequiredEnum\":null,\"NullableRequiredInt\":null}",
                    testId = "request-bodies-post-nullable-required-property-all-null"
                },
                new CommonHelpers.TestTableEntry()
                {
                    name = "Optional field initialized",
                    arg = new NullableRequiredPropertyPostRequestBody{NullableOptionalInt = 0},
                    want = "{\"NullableRequiredArray\":null,\"NullableRequiredBigIntStr\":null,\"NullableRequiredDateTime\":null,\"NullableRequiredDecimalStr\":null,\"NullableRequiredEnum\":null,\"NullableRequiredInt\":null,\"NullableOptionalInt\":0}"
                },
                new CommonHelpers.TestTableEntry()
                {
                    name = "All fields set to non-null value",
                    arg = new NullableRequiredPropertyPostRequestBody
                    {
                        NullableOptionalInt = 0,
                        NullableRequiredArray = new List<double>{1.1, 2.2, 3.3},
                        NullableRequiredEnum = NullableRequiredEnum.Second,
                        NullableRequiredInt = 1,
                        NullableRequiredDateTime = System.DateTime.Parse("2020-01-01T00:00:00Z"),
                        NullableRequiredBigIntStr = BigInteger.Parse("9223372036854775807"),
                        NullableRequiredDecimalStr = 3.141592653589793238462643383279M
                    },
                    want = "{\"NullableRequiredArray\":[1.1,2.2,3.3],\"NullableRequiredBigIntStr\":\"9223372036854775807\",\"NullableRequiredDateTime\":\"2020-01-01T00:00:00.0000000Z\",\"NullableRequiredDecimalStr\":\"3.1415926535897932384626433833\",\"NullableRequiredEnum\":\"second\",\"NullableRequiredInt\":1,\"NullableOptionalInt\":0}",
                    testId = "request-bodies-post-nullable-required-property-all-set"
                }
            };

            foreach (var test in tests)
            {
                if (!string.IsNullOrEmpty(test.testId))
                {
                    CommonHelpers.RecordTest(test.testId);
                }

                var req = (NullableRequiredPropertyPostRequestBody)test.arg;
                Assert.AreEqual(test.want, Helpers.GetSerializedBodyJson(req));

                using (
                    var res = await sdk.RequestBodies.NullableRequiredPropertyPostAsync(req)
                )
                {
                    Assert.AreEqual(200, res.StatusCode);
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostNullableRequiredSharedObject()
    {
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var tests = new CommonHelpers.TestTableEntry[] {
                new CommonHelpers.TestTableEntry()
                {
                    name = "Required field set to null",
                    arg = new NullableRequiredSharedObjectPostRequestBody{},
                    want = "{\"NullableRequiredObj\":null}",
                    testId = "request-bodies-post-nullable-required-shared-object-required-null"
                },
                new CommonHelpers.TestTableEntry()
                {
                    name = "Both fields set to null",
                    arg = new NullableRequiredSharedObjectPostRequestBody
                    {
                        NullableOptionalObj = null,
                        NullableRequiredObj = null
                    },
                    want = "{\"NullableRequiredObj\":null}",
                    testId = "request-bodies-post-nullable-required-shared-object-all-null"
                },
                new CommonHelpers.TestTableEntry()
                {
                    name = "Optional field set to non-null value",
                    arg = new NullableRequiredSharedObjectPostRequestBody
                    {
                        NullableOptionalObj = new NullableOptionalObject{Required = 1}
                    },
                    want = "{\"NullableRequiredObj\":null,\"NullableOptionalObj\":{\"required\":1}}",
                    testId = "request-bodies-post-nullable-required-shared-object-optional-non-null"
                },
                new CommonHelpers.TestTableEntry()
                {
                    name = "Both fields set to non-null value",
                    arg = new NullableRequiredSharedObjectPostRequestBody
                    {
                        NullableOptionalObj = new NullableOptionalObject{Required = 1, Optional = "test"},
                        NullableRequiredObj = new NullableObject{Required = 2}
                    },
                    want = "{\"NullableRequiredObj\":{\"required\":2},\"NullableOptionalObj\":{\"optional\":\"test\",\"required\":1}}",
                    testId = "request-bodies-post-nullable-required-shared-object-all-set"
                }
            };

            foreach (var test in tests)
            {
                if (!string.IsNullOrEmpty(test.testId))
                {
                    CommonHelpers.RecordTest(test.testId);
                }

                var req = (NullableRequiredSharedObjectPostRequestBody)test.arg;
                Assert.AreEqual(test.want, Helpers.GetSerializedBodyJson(req));

                using (
                    var res = await sdk.RequestBodies.NullableRequiredSharedObjectPostAsync(req)
                )
                {
                    Assert.AreEqual(200, res.StatusCode);
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPostNullableRequiredEmptyObject()
    {
        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            var tests = new CommonHelpers.TestTableEntry[] {
                new CommonHelpers.TestTableEntry()
                {
                    name = "Empty initializer",
                    arg = new NullableRequiredEmptyObjectPostRequestBody{},
                    want = "{\"NullableRequiredObj\":null}"  // TODO SPE-2837
                    // want = "{\"NullableRequiredObj\":null,\"RequiredObj\":{}}"
                },
                new CommonHelpers.TestTableEntry()
                {
                    name = "Required field initialized only",
                    arg = new NullableRequiredEmptyObjectPostRequestBody
                    {
                        RequiredObj = new RequiredObj{}
                    },
                    want = "{\"NullableRequiredObj\":null,\"RequiredObj\":{}}",
                    testId = "request-bodies-post-nullable-required-empty-object-nullable-set"
                },
                new CommonHelpers.TestTableEntry()
                {
                    name = "Optional field initialized only",
                    arg = new NullableRequiredEmptyObjectPostRequestBody
                    {
                        NullableOptionalObj = new NullableOptionalObj{},
                    },
                    want = "{\"NullableOptionalObj\":{},\"NullableRequiredObj\":null}", // TODO SPE-2837
                    // want = "{\"NullableRequiredObj\":null,\"NullableOptionalObj\":{},\"RequiredObj\":{}}",
                    testId = "request-bodies-post-nullable-required-empty-object-optional-set"
                },
                new CommonHelpers.TestTableEntry()
                {
                    name = "All fields initialized",
                    arg = new NullableRequiredEmptyObjectPostRequestBody
                    {
                        RequiredObj = new RequiredObj{},
                        NullableOptionalObj = new NullableOptionalObj{},
                        NullableRequiredObj = new NullableRequiredObj{},
                    },
                    want = "{\"NullableRequiredObj\":{},\"RequiredObj\":{},\"NullableOptionalObj\":{}}",
                    testId = "request-bodies-post-nullable-required-empty-object-all-set"
                }
            };

            foreach (var test in tests)
            {
                if (!string.IsNullOrEmpty(test.testId))
                {
                    CommonHelpers.RecordTest(test.testId);
                }

                var req = (NullableRequiredEmptyObjectPostRequestBody)test.arg;
                Assert.AreEqual(test.want, Helpers.GetSerializedBodyJson(req));

                using (
                    var res = await sdk.RequestBodies.NullableRequiredEmptyObjectPostAsync(req)
                )
                {
                    Assert.AreEqual(200, res.StatusCode);
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyNoBodyNoContentType()
    {
        CommonHelpers.RecordTest("request-bodies-no-body-no-content-type");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK();

            using (
                var res = await sdk.Methods.MethodGetAsync(Helpers.ApiTestServiceUrl)
            )
            {
                Assert.Null(res.RawResponse.uploadHandler);

                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual("OK", res.Object.Status);
            }
        });
    }

    [UnityTest]
    public IEnumerator RequestBodyPutMultipartFilesArray()
    {
        CommonHelpers.RecordTest("request-bodies-post-multipart-files-array");

        yield return CommonHelpers.Await(async () =>
        {
            var sdk = new SDK(serverUrl: Helpers.ApiTestServiceUrl);

            var data = Helpers.GetData();

            using (
                var res = await sdk.RequestBodies.RequestBodyPutMultipartFilesArrayAsync(
                    new RequestBodyPutMultipartFilesArrayRequestBody()
                    {
                        Files = new List<Files>()
                        {
                            new Files() { Content = data, FileName = "testUpload.json" },
                            new Files() { Content = data, FileName = "some-other-name.json" },
                        },
                    }
                )
            )
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.NotNull(res.Res);
                Assert.AreEqual(2, res.Res.Files.Count);

                Assert.AreEqual(
                    "ewAKACAAIAAiAHMAbwBtAGUAIgA6ACAAIgBqAHMAbwBuACIALAANAAoAIAAgACIAdABvACIAOgAgACIAYgBlACIALAANAAoAIAAgACIAdQBwAGwAbwBhAGQAZQBkACIAOgAgACIAaQBuACIALAANAAoAIAAgACIAYQAiADoAIAAiAGYAaQBsAGUAIgANAAoAfQANAAoA",
                    res.Res.Files[0].Content
                );
                Assert.AreEqual("application/octet-stream", res.Res.Files[0].ContentType);
                Assert.AreEqual("files", res.Res.Files[0].FieldName);
                Assert.AreEqual("testUpload.json", res.Res.Files[0].Filename);
                Assert.AreEqual(150, res.Res.Files[0].Size);

                Assert.AreEqual(
                    "ewAKACAAIAAiAHMAbwBtAGUAIgA6ACAAIgBqAHMAbwBuACIALAANAAoAIAAgACIAdABvACIAOgAgACIAYgBlACIALAANAAoAIAAgACIAdQBwAGwAbwBhAGQAZQBkACIAOgAgACIAaQBuACIALAANAAoAIAAgACIAYQAiADoAIAAiAGYAaQBsAGUAIgANAAoAfQANAAoA",
                    res.Res.Files[1].Content
                );
                Assert.AreEqual("application/octet-stream", res.Res.Files[1].ContentType);
                Assert.AreEqual("files", res.Res.Files[1].FieldName);
                Assert.AreEqual("some-other-name.json", res.Res.Files[1].Filename);
                Assert.AreEqual(150, res.Res.Files[1].Size);

                Assert.NotNull(res.Res.FormFields);
                Assert.IsEmpty(res.Res.FormFields);
            }
        });
    }
}
