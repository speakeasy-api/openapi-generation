using System;
using System.Collections.Generic;
using System.Net;
using System.Threading.Tasks;
using System.Web;
using Newtonsoft.Json;
using NodaTime;
using Company.Product.Feature.Subnamespace;
using Company.Product.Feature.Subnamespace.Models.Operations;
using Company.Product.Feature.Subnamespace.Models.Shared;
using Company.Product.Feature.Subnamespace.Utils;
using Xunit;

// Legacy-mode union tests (unionStrategy: left-to-right)
public class UnionsShould
{
    private static void AssertJsonEquivalent(object? expected, object? actual)
    {
        Assert.Equal(CommonHelpers.NormalizeJson(expected), CommonHelpers.NormalizeJson(actual));
    }

    // Primary: unions-collections-one-of-post
    [Fact]
    public async Task CollectionOneOfPost()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var req = CollectionOneOfObject.CreateArrayOfAny(new List<object> { "one", "two" });

        var res = await s.Unions.CollectionOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        AssertJsonEquivalent(req.ArrayOfAny, res.Res!.Json.ArrayOfAny);

        var req2 = CollectionOneOfObject.CreateMapOfAny(
            new Dictionary<string, object>() { { "1", "one" }, { "2", "two" } }
        );

        var res2 = await s.Unions.CollectionOneOfPostAsync(req2);
        Assert.Equal(HttpStatusCode.OK, res2.RawResponse.StatusCode);
        AssertJsonEquivalent(req.MapOfAny, res.Res!.Json.MapOfAny);
    }

    // Primary: unions-weakly-typed-one-of-post-basic
    [Fact]
    public async Task WeaklyTypedOneOfPost_Basic()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var obj = Helpers.CreateSimpleObject();

        var req = WeaklyTypedOneOfObject.CreateSimpleObject(obj);

        var res = await s.Unions.WeaklyTypedOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(WeaklyTypedOneOfObjectType.SimpleObject, res.Res!.Json.Type.ToString());
        AssertJsonEquivalent(obj, res.Res!.Json.SimpleObject);
    }

    // Primary: unions-weakly-typed-one-of-post-deep
    [Fact]
    public async Task WeaklyTypedOneOfPost_Deep()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var obj = Helpers.CreateDeepObject();

        var req = WeaklyTypedOneOfObject.CreateDeepObject(obj);

        var res = await s.Unions.WeaklyTypedOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(WeaklyTypedOneOfObjectType.DeepObject, res.Res!.Json.Type.ToString());
        AssertJsonEquivalent(obj, res.Res!.Json.DeepObject);
    }

    // Primary: unions-typed-object-one-of-post-null
    // Divergence note: the legacy converter still throws InvalidOperationException on a null
    // JSON value for a non-nullable union; the smart-union (populated-fields) path throws
    // JsonSerializationException instead. This suite keeps the ORIGINAL legacy expectation.
    [Fact]
    public void TypedObjectOneOfPost_Null()
    {
        // The type system prevents creating a TypedObjectOneOf with Type = "null";
        var method = typeof(TypedObjectOneOf).GetMethod("CreateNull");
        Assert.Null(method);

        // Deserialization should fail since TypedObjectOneOf is non-nullable
        var ex = Assert.Throws<InvalidOperationException>(
            () => ResponseBodyDeserializer.DeserializeNotNull<TypedObjectOneOfPostRes>(
                "{\"json\": null}",
                NullValueHandling.Ignore
            )
        );

        Assert.Equal("Received unexpected null JSON value", ex.Message);
    }

    // Primary: unions-flattened-typed-object-post-obj1
    [Fact]
    public async Task FlattenedTypedObject_Obj1()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var obj = FlattenedTypedObject1.CreateTypedObject1(
            new TypedObject1 { Value = "one", Type = TypedObject1Type.Obj1 }
        );

        var res = await s.Unions.FlattenedTypedObjectPostAsync(obj);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        AssertJsonEquivalent(obj, res.Res!.Json);
    }

    // Primary: unions-nullable-typed-object-post-obj1
    [Fact]
    public async Task NullableTypedObjectPost_Obj1()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var obj = new TypedObject1 { Value = "one", Type = TypedObject1Type.Obj1 };

        var res = await s.Unions.NullableTypedObjectPostAsync(obj);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        AssertJsonEquivalent(obj, res.Res!.Json);
    }

    // Primary: unions-nullable-typed-object-post-null
    [Fact]
    public async Task NullableTypedObjectPost_Null()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var res = await s.Unions.NullableTypedObjectPostAsync(null);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Null(res.Res!.Json);
    }

    // Primary: unions-nullable-oneof-type-in-object-post
    [Fact]
    public async Task NullableOneOfTypeInObject()
    {
        var tests = new CommonHelpers.TestTableEntry[]
        {
            new CommonHelpers.TestTableEntry
            {
                name = "Non-nullable field set only",
                arg = new NullableOneOfTypeInObject { OneOfOne = true },
                want = "{\"NullableOneOfOne\":null,\"NullableOneOfTwo\":null,\"OneOfOne\":true}"
            },
            new CommonHelpers.TestTableEntry
            {
                name = "Nullable fields set to null",
                arg = new NullableOneOfTypeInObject
                {
                    NullableOneOfOne = null,
                    NullableOneOfTwo = null,
                    OneOfOne = true
                },
                want = "{\"NullableOneOfOne\":null,\"NullableOneOfTwo\":null,\"OneOfOne\":true}"
            },
            new CommonHelpers.TestTableEntry
            {
                name = "All fields set to non-null values",
                arg = new NullableOneOfTypeInObject
                {
                    NullableOneOfOne = true,
                    NullableOneOfTwo = NullableOneOfTypeInObjectNullableOneOfTwo.CreateInteger(2),
                    OneOfOne = true
                },
                want = "{\"NullableOneOfOne\":true,\"NullableOneOfTwo\":2,\"OneOfOne\":true}"
            }
        };

        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");
        foreach (var test in tests)
        {
            var req = test.arg;
            var serializedBody = RequestBodySerializer.Serialize(
                req,
                "Request",
                "json",
                false,
                false
            );

            var res = await s.Unions.NullableOneOfTypeInObjectPostAsync(
                (NullableOneOfTypeInObject)req
            );
            Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
            AssertJsonEquivalent(req, res.Res!.Json);
        }
    }

    // Primary: unions-primitive-type-one-of-post-string
    [Fact]
    public async Task PrimitiveTypeOneOfPost_String()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var req = PrimitiveTypeOneOfPostRequestBody.CreateStr("test");

        var res = await s.Unions.PrimitiveTypeOneOfPostAsync(req);

        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(PrimitiveTypeOneOfPostJsonType.Str, res.Res!.Json.Type.ToString());
        Assert.Equal("test", res.Res!.Json.Str);
    }

    // Primary: unions-primitive-type-one-of-post-integer
    [Fact]
    public async Task PrimitiveTypeOneOfPost_Integer()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var req = PrimitiveTypeOneOfPostRequestBody.CreateInteger(111);

        var res = await s.Unions.PrimitiveTypeOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(PrimitiveTypeOneOfPostJsonType.Integer, res.Res!.Json.Type.ToString());
        Assert.Equal(111, res.Res!.Json.Integer);
    }

    // Primary: unions-primitive-type-one-of-post-number
    [Fact]
    public async Task PrimitiveTypeOneOfPost_Number()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var req = PrimitiveTypeOneOfPostRequestBody.CreateNumber(22.2);

        var res = await s.Unions.PrimitiveTypeOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(PrimitiveTypeOneOfPostJsonType.Number, res.Res!.Json.Type.ToString());
        Assert.Equal(22.2, res.Res!.Json.Number);
    }

    // Primary: unions-primitive-type-one-of-post-boolean
    [Fact]
    public async Task PrimitiveTypeOneOfPost_Boolean()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var req = PrimitiveTypeOneOfPostRequestBody.CreateBoolean(true);

        var res = await s.Unions.PrimitiveTypeOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(PrimitiveTypeOneOfPostJsonType.Boolean, res.Res!.Json.Type.ToString());
        Assert.Equal(true, res.Res!.Json.Boolean);
    }

    // Primary: unions-mixed-type-one-of-post-string
    [Fact]
    public async Task MixedTypeOneOfPost_String()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var req = MixedTypeOneOfPostRequestBody.CreateStr("test");

        var res = await s.Unions.MixedTypeOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(MixedTypeOneOfPostJsonType.Str, res.Res!.Json.Type.ToString());
        Assert.Equal("test", res.Res!.Json.Str);
    }

    // Primary: unions-mixed-type-one-of-post-integer
    [Fact]
    public async Task MixedTypeOneOfPost_Integer()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var req = MixedTypeOneOfPostRequestBody.CreateInteger(111);

        var res = await s.Unions.MixedTypeOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(MixedTypeOneOfPostJsonType.Integer, res.Res!.Json.Type.ToString());
        Assert.Equal(111, res.Res!.Json.Integer);
    }

    // Primary: unions-mixed-type-one-of-post-object
    [Fact]
    public async Task MixedTypeOneOfPost_Object()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var obj = Helpers.CreateSimpleObject();

        var req = MixedTypeOneOfPostRequestBody.CreateSimpleObject(obj);

        var res = await s.Unions.MixedTypeOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(MixedTypeOneOfPostJsonType.SimpleObject, res.Res!.Json.Type.ToString());
        AssertJsonEquivalent(obj, res.Res!.Json.SimpleObject);
    }

    // Primary: unions-date-null
    [Fact]
    public async Task DateNullUnion()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var date = LocalDate.FromDateTime(System.DateTime.Parse("2020-01-01"));

        var res = await s.Unions.UnionDateNullAsync(date);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(date, res.Res!.Json);
    }

    // Primary: unions-datetime-null
    [Fact]
    public async Task DateTimeNullUnion()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var dateTime = System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime();

        var res = await s.Unions.UnionDateTimeNullAsync(dateTime);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(dateTime, res.Res!.Json);
    }

    // Primary: unions-datetime-bigint
    [Fact]
    public async Task DateTimeBigintUnion()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var req = UnionDateTimeBigIntRequestBody.CreateDateTime(
            System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime()
        );
        var res = await s.Unions.UnionDateTimeBigIntAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(UnionDateTimeBigIntRequestBodyType.DateTime, res.Res!.Json.Type.ToString());
        Assert.Equal(
            System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime(),
            res.Res!.Json.DateTime
        );

        var nextRes = await s.Unions.UnionDateTimeBigIntAsync(
            UnionDateTimeBigIntRequestBody.CreateBigint(9007199254740991)
        );
        Assert.Equal(HttpStatusCode.OK, nextRes.RawResponse.StatusCode);
        Assert.Equal(UnionDateTimeBigIntRequestBodyType.Bigint, nextRes.Res!.Json.Type);
        Assert.Equal(9007199254740991, nextRes.Res!.Json.Bigint);
    }

    // Primary: unions-bigint-str-decimal
    [Fact]
    public async Task UnionBigIntStrDecimal()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var req = UnionBigIntStrDecimalRequestBody.CreateDecimal(3.141592653589793M);
        var json = await Helpers.GetSerializedBodyJson(req);
        Assert.Equal("3.141592653589793", json);

        var res = await s.Unions.UnionBigIntStrDecimalAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(UnionBigIntStrDecimalRequestBodyType.Decimal, res.Res!.Json.Type);
        Assert.Equal(3.141592653589793M, res.Res!.Json.Decimal);

        req = UnionBigIntStrDecimalRequestBody.CreateBigint(9223372036854775807);
        json = await Helpers.GetSerializedBodyJson(req);
        Assert.Equal("\"9223372036854775807\"", json);

        res = await s.Unions.UnionBigIntStrDecimalAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(UnionBigIntStrDecimalRequestBodyType.Bigint, res.Res!.Json.Type);
        Assert.Equal(9223372036854775807, res.Res!.Json.Bigint);
    }

    // Primary: unions-union-map
    [Fact]
    public async Task UnionMap()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var res = await s.Unions.UnionMapAsync(new UnionMapRequestBody
        {
            Input = new Dictionary<string, OneOfPrimitives>{
                    {"str", OneOfPrimitives.CreateStr("test")},
                    {"bool", OneOfPrimitives.CreateBoolean(true)}
                }
        });

        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal("test", res.Res!.Json.Input["str"].Str);
        Assert.Equal(true, res.Res!.Json.Input["bool"].Boolean);
    }

    // Primary: unions-extra-json-properties
    [Fact]
    public async Task UnionExtraJsonProperties()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var res = await s.Unions.OneOfOverlappingObjectsAsync(new OneOfOverlappingObjectsRequestBody
        {
            Field1 = "test1", // common to Obj1 and Obj2
            Field3 = 1  // not part of Obj1 nor Obj2
        });

        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(OneOfOverlappingObjectsType.Obj1, res.Res!.Json.Type);
        Assert.Equal("test1", res.Res!.Json.Obj1!.Field1);

        res = await s.Unions.OneOfOverlappingObjectsAsync(new OneOfOverlappingObjectsRequestBody
        {
            Field1 = "test2",  // common to Obj1 and Obj2
            Field2 = true,  // exclusive to Obj2
            Field3 = 1  // not part of Obj1 nor Obj2
        });

        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(OneOfOverlappingObjectsType.Obj2, res.Res!.Json.Type);
        Assert.Equal("test2", res.Res!.Json.Obj2!.Field1);
        Assert.True(res.Res!.Json.Obj2!.Field2);
    }

    // Primary: unions-nested-enums-form
    [Fact]
    public async Task UnionNestedEnumsForm()
    {
        var sdk = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var obj1 = new NestedEnumArray()
        {
            Tags = "one,two",
            Enums = new List<Company.Product.Feature.Subnamespace.Models.Shared.Enum>()
            {
                Company.Product.Feature.Subnamespace.Models.Shared.Enum.One,
                Company.Product.Feature.Subnamespace.Models.Shared.Enum.Two,
            }
        };

        var req1 = UnionNestedEnumsFormRequestBody.CreateNestedEnumArray(obj1);
        var serializedForm1 = RequestBodySerializer.Serialize(req1, "Request", "form", false, false);
        var form1 = HttpUtility.UrlDecode(await serializedForm1!.ReadAsStringAsync());
        Assert.Equal("enums[]=one&enums[]=two&tags=one,two", form1);

        var res1 = await sdk.Unions.UnionNestedEnumsFormAsync(req1);
        Assert.Equal(HttpStatusCode.OK, res1.RawResponse.StatusCode);

        var obj2 = new NestedEnumMap()
        {
            Tags = "two,three",
            Enums = new Dictionary<string, Company.Product.Feature.Subnamespace.Models.Shared.Enum>()
            {
                { "key2", Company.Product.Feature.Subnamespace.Models.Shared.Enum.Two },
                { "key3", Company.Product.Feature.Subnamespace.Models.Shared.Enum.Three },
            },
        };

        var req2 = UnionNestedEnumsFormRequestBody.CreateNestedEnumMap(obj2);
        var serializedForm2 = RequestBodySerializer.Serialize(req2, "Request", "form", false, false);
        var form2 = HttpUtility.UrlDecode(await serializedForm2!.ReadAsStringAsync());
        Assert.Equal("enums={\"key2\":\"two\",\"key3\":\"three\"}&tags=two,three", form2);

        var res2 = await sdk.Unions.UnionNestedEnumsFormAsync(req2);
        Assert.Equal(HttpStatusCode.OK, res2.RawResponse.StatusCode);
    }

    // Primary: unions-nested-enums-multipart
    [Fact]
    public async Task UnionNestedEnumsMultipart()
    {
        var sdk = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var enumArray = new List<Company.Product.Feature.Subnamespace.Models.Shared.Enum>()
        {
            Company.Product.Feature.Subnamespace.Models.Shared.Enum.One,
            Company.Product.Feature.Subnamespace.Models.Shared.Enum.Two,
        };

        var req1 = new UnionNestedEnumsMultipartRequestBody() { Enums = Company.Product.Feature.Subnamespace.Models.Operations.Enums.CreateArrayOfEnum(enumArray) };
        var serializedForm1 = RequestBodySerializer.Serialize(req1, "Request", "multipart", false, false);
        var form1 = await serializedForm1!.ReadAsStringAsync();
        Assert.Contains("Content-Disposition: form-data; name=enums", form1);
        Assert.Contains("[\"one\",\"two\"]", form1);

        var res1 = await sdk.Unions.UnionNestedEnumsMultipartAsync(req1);
        Assert.Equal(HttpStatusCode.OK, res1.RawResponse.StatusCode);

        var enumMap = new Dictionary<string, Company.Product.Feature.Subnamespace.Models.Shared.Enum>()
        {
            { "key2", Company.Product.Feature.Subnamespace.Models.Shared.Enum.Two },
            { "key3", Company.Product.Feature.Subnamespace.Models.Shared.Enum.Three },
        };

        var req2 = new UnionNestedEnumsMultipartRequestBody() { Enums = Company.Product.Feature.Subnamespace.Models.Operations.Enums.CreateMapOfEnum(enumMap) };
        var serializedForm2 = RequestBodySerializer.Serialize(req2, "Request", "multipart", false, false);
        var form2 = await serializedForm2!.ReadAsStringAsync();
        Assert.Contains("Content-Disposition: form-data; name=enums", form2);
        Assert.Contains("{\"key2\":\"two\",\"key3\":\"three\"}", form2);

        var res2 = await sdk.Unions.UnionNestedEnumsMultipartAsync(req2);
        Assert.Equal(HttpStatusCode.OK, res2.RawResponse.StatusCode);
    }

    // Primary: unions-union-of-arrays
    [Fact]
    public async Task UnionOfArraysTest()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var arrayOf1 = new List<Company.Product.Feature.Subnamespace.Models.Shared.One> { new Company.Product.Feature.Subnamespace.Models.Shared.One { Foo = "foo0" }, new Company.Product.Feature.Subnamespace.Models.Shared.One { Foo = "foo1" } };
        var req1 = UnionOfArrays.CreateArrayOf1(arrayOf1);
        var res1 = await s.Unions.UnionOfArraysPostAsync(req1);

        Assert.Equal(HttpStatusCode.OK, res1.RawResponse.StatusCode);
        Assert.Equal(UnionOfArraysType.ArrayOf1, res1.Res!.Json.Type);
        Assert.Equal(2, res1.Res!.Json.ArrayOf1!.Count);
        Assert.Equal("foo0", res1.Res!.Json.ArrayOf1![0].Foo);
        Assert.Equal("foo1", res1.Res!.Json.ArrayOf1![1].Foo);

        var arrayOf2 = new List<UnionOfArrays2> {
            new UnionOfArrays2 { Bar = "bar0" },
            new UnionOfArrays2 { Bar = "bar1" }
        };
        var req2 = UnionOfArrays.CreateArrayOfUnionOfArrays2(arrayOf2);
        var res2 = await s.Unions.UnionOfArraysPostAsync(req2);

        Assert.Equal(HttpStatusCode.OK, res2.RawResponse.StatusCode);
        Assert.Equal(UnionOfArraysType.ArrayOfUnionOfArrays2, res2.Res!.Json.Type);
        Assert.Equal(2, res2.Res!.Json.ArrayOfUnionOfArrays2!.Count);
        Assert.Equal("bar0", res2.Res!.Json.ArrayOfUnionOfArrays2![0].Bar);
        Assert.Equal("bar1", res2.Res!.Json.ArrayOfUnionOfArrays2![1].Bar);

        var arrayOf3 = new List<Three> { new Three { Baz = "baz0" }, new Three { Baz = "baz1" } };
        var req3 = UnionOfArrays.CreateArrayOf3(arrayOf3);
        var res3 = await s.Unions.UnionOfArraysPostAsync(req3);

        Assert.Equal(HttpStatusCode.OK, res3.RawResponse.StatusCode);
        Assert.Equal(UnionOfArraysType.ArrayOf3, res3.Res!.Json.Type);
        Assert.Equal(2, res3.Res!.Json.ArrayOf3!.Count);
        Assert.Equal("baz0", res3.Res!.Json.ArrayOf3![0].Baz);
        Assert.Equal("baz1", res3.Res!.Json.ArrayOf3![1].Baz);
    }

    // Primary: unions-conflicting-discriminator-mapping-key
    [Fact]
    public async Task ConflictingDiscriminatorMappingKey()
    {
        var s = new SDK(apiKeyAuth: "Token YOUR_API_KEY");

        var taggedObj = new TaggedObject1
        {
            ImageURL = "https://example.com/image.png",
            Tag = Tag.Tag1
        };

        var req = new ConflictingDiscriminatorMappingKey
        {
            Prefix = StronglyTypedOneOfDiscriminatedObject.CreateTag1(taggedObj),
            PrefixTag2 = new PrefixTag2
            {
                Str = "test"
            }
        };

        var res = await s.Unions.ConflictingDiscriminatorMappingKeyAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.Equal(StronglyTypedOneOfDiscriminatedObjectType.Tag1, res.Res!.Json.Prefix!.Type);
        Assert.Equal(taggedObj.ImageURL, res.Res!.Json.Prefix!.TaggedObject1!.ImageURL);
        Assert.Equal(taggedObj.Tag, res.Res!.Json.Prefix!.TaggedObject1!.Tag);
        Assert.Equal("test", res.Res!.Json.PrefixTag2!.Str);
    }

    // Primary: unions-circular-reference-recursive-one-of
    [Fact]
    public async Task CircularReferenceRecursiveOneOf()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, apiKeyAuth: "Token YOUR_API_KEY");

        var payload = RecursiveOneOfValue.CreateArrayOfRecursiveOneOfValue(
            new List<RecursiveOneOfValue>
            {
                RecursiveOneOfValue.CreateStr("hello"),
                RecursiveOneOfValue.CreateMapOfRecursiveOneOfValue(
                    new Dictionary<string, RecursiveOneOfValue>
                    {
                        ["nested"] = RecursiveOneOfValue.CreateArrayOfRecursiveOneOfValue(
                            new List<RecursiveOneOfValue>
                            {
                                RecursiveOneOfValue.CreateStr("world"),
                            }
                        ),
                    }
                ),
            }
        );

        var res = await sdk.Unions.CircularReferenceRecursiveOneOfAsync(
            new CircularReferenceRecursiveOneOfRequestBody { Value = payload }
        );

        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        Assert.NotNull(res.Object);
        Assert.Equal(
            JsonConvert.SerializeObject(payload),
            JsonConvert.SerializeObject(res.Object!.Json.Value)
        );
    }

    // No primary equivalent — legacy-only test, INTENTIONALLY DIVERGENT from the primary
    // suite's smart-union expectations.
    //
    // Legacy candidate order: the legacy (non-smart) branch of union.cs.stmpl iterates
    // `sortTypeDefFieldCount .Local.Type.AssociatedTypes` (internal/js/template/engine.go),
    // which sorts class candidates by ASCENDING field count (classes before non-classes).
    // So SmartUnionMoreFieldsA (1 field: foo) is tried before SmartUnionMoreFieldsB
    // (2 fields: bar, foo), and the first candidate that deserializes strictly
    // (MissingMemberHandling.Error) wins — left-to-right first-match semantics.
    [Fact]
    public void LegacyUnion_LeftToRightFirstMatchSemantics()
    {
        // Ambiguous payload representable by BOTH members ({"foo"} is a strict subset of A
        // and of B): the first candidate in legacy order (A, fewest fields) wins immediately;
        // B is never considered.
        var first = JsonConvert.DeserializeObject<SmartUnionSelectsMoreMatchedFieldsJson>(
            "{\"foo\":\"test\"}"
        )!;
        Assert.Equal(
            SmartUnionSelectsMoreMatchedFieldsJsonType.SmartUnionMoreFieldsA,
            first.Type
        );
        Assert.NotNull(first.SmartUnionMoreFieldsA);
        Assert.Null(first.SmartUnionMoreFieldsB);
        Assert.Equal("test", first.SmartUnionMoreFieldsA!.Foo);

        // DIVERGENT case: payload {"foo":"test","bar":123}. The smart (populated-fields)
        // strategy scores bar:123 as an INEXACT match for B.Bar (string) and prefers A
        // (1 exact match, 0 inexact) over B (1 exact, 1 inexact). The LEGACY converter
        // instead rejects A in the strict pass (unknown member 'bar' raises
        // MissingMemberException) and accepts B, with Newtonsoft coercing 123 -> "123".
        var coerced = JsonConvert.DeserializeObject<SmartUnionSelectsMoreMatchedFieldsJson>(
            "{\"foo\":\"test\",\"bar\":123}"
        )!;
        Assert.Equal(
            SmartUnionSelectsMoreMatchedFieldsJsonType.SmartUnionMoreFieldsB,
            coerced.Type
        );
        Assert.Null(coerced.SmartUnionMoreFieldsA);
        Assert.NotNull(coerced.SmartUnionMoreFieldsB);
        Assert.Equal("test", coerced.SmartUnionMoreFieldsB!.Foo);
        Assert.Equal("123", coerced.SmartUnionMoreFieldsB!.Bar);
    }

    // Quaternary sets forwardCompatibleUnionsByDefault: false, so the Vehicle
    // discriminated union is generated CLOSED: no Unknown sentinel.

    private static Vehicle DeserializeVehicle(string json)
    {
        return JsonConvert.DeserializeObject<Vehicle>(
            json,
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;
    }

    // Closed counterpart for primary's open-union-known-variant
    [Fact]
    public void ClosedUnion_ParseKnownVariant()
    {
        var car = DeserializeVehicle("{\"vehicleType\":\"car\",\"wheelsType\":\"four\"}");
        Assert.Equal(VehicleType.Car.ToString(), car.Type.ToString());
        Assert.NotNull(car.Car);
    }

    // Closed counterpart for primary's open-union-unknown-discriminator
    [Fact]
    public void ClosedUnion_ThrowOnUnknownDiscriminator()
    {
        var ex = Assert.Throws<InvalidOperationException>(
            () => DeserializeVehicle("{\"vehicleType\":\"spaceship\",\"thrust\":9000}")
        );
        Assert.Contains("Could not deserialize", ex.Message);
    }

    // Closed counterpart for primary's open-union-missing-discriminator
    [Fact]
    public void ClosedUnion_ThrowOnMissingDiscriminator()
    {
        Assert.Throws<ArgumentNullException>(
            () => DeserializeVehicle("{\"wheelsType\":\"four\"}")
        );
    }

    // Date-time-shaped strings inside unrecognized open-union payloads must be captured raw.
    // This variant's legacy reader (presenceAwareJsonSerialization: false) would eagerly parse
    // them into DateTime (normalizing timezones and reformatting), so raw capture bypasses
    // date parsing via Utilities.LoadRawToken / ParseRawToken).

    private const string OffsetTimestamp = "2024-01-02T03:04:05+02:00";

    [Fact]
    public void OpenTagged_UnknownDiscriminator_ReplaysDateTimeVerbatim()
    {
        var v = JsonConvert.DeserializeObject<OpenDateTimeTaggedUnion>(
            "{\"kind\":\"paused\",\"occurredAt\":\"" + OffsetTimestamp + "\"}",
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;
        Assert.True(v.IsUnknown());
        var raw = (Newtonsoft.Json.Linq.JObject)v.UnknownRaw!;
        Assert.Equal(Newtonsoft.Json.Linq.JTokenType.String, raw["occurredAt"]!.Type);
        Assert.Equal(OffsetTimestamp, (string?)raw["occurredAt"]);
        Assert.Contains(OffsetTimestamp, JsonConvert.SerializeObject(v));
    }

    [Fact]
    public void OpenTagged_KnownVariant_Deserializes()
    {
        var v = JsonConvert.DeserializeObject<OpenDateTimeTaggedUnion>(
            "{\"kind\":\"started\",\"occurredAt\":\"2024-01-02T03:04:05Z\"}",
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;
        Assert.False(v.IsUnknown());
        Assert.NotNull(v.DateTimeEventStarted);
    }

    [Fact]
    public void OpenUntagged_DateShapedString_FallsBackToUnknown()
    {
        // A root-level date-shaped string matches no object variant, so the
        // left-to-right fallback captures it as Unknown. Root primitives are
        // read by the serializer before the union converter runs, so on this
        // legacy variant the reader has already materialized a DateTime:
        // UTC instants replay verbatim, but non-UTC offsets are normalized
        // (pre-existing reader limitation, same as untyped fields).
        var v = JsonConvert.DeserializeObject<OpenDateTimeUntaggedUnion>(
            "\"2024-01-02T03:04:05Z\"",
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;
        Assert.True(v.IsUnknown());
        Assert.Equal("2024-01-02T03:04:05Z", JsonConvert.SerializeObject(v).Trim('"'));
    }

    [Fact]
    public void OpenUntagged_KnownVariant_Deserializes()
    {
        var v = JsonConvert.DeserializeObject<OpenDateTimeUntaggedUnion>(
            "{\"betaAt\":\"2024-01-02T03:04:05Z\",\"count\":2}",
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;
        Assert.False(v.IsUnknown());
        Assert.NotNull(v.DateTimeRecordBeta);
    }

    [Fact]
    public async Task OpenUnionDateTimeEnvelope_UnknownPayloads_RoundTripVerbatim()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, apiKeyAuth: "Bearer testToken");

        // ParseRawToken keeps date-shaped strings as string tokens; JObject.Parse
        // would coerce them to DateTime before they ever reach the SDK.
        var unknownTagged = OpenDateTimeTaggedUnion.CreateUnknown(
            Utilities.ParseRawToken("{\"kind\":\"paused\",\"occurredAt\":\"" + OffsetTimestamp + "\"}"));
        var unknownUntagged = OpenDateTimeUntaggedUnion.CreateUnknown(
            Utilities.ParseRawToken("{\"gammaAt\":\"" + OffsetTimestamp + "\"}"));

        var res = await sdk.Unions.OpenUnionDateTimePostAsync(new OpenUnionDateTimeEnvelope {
            Tagged = unknownTagged,
            Untagged = unknownUntagged,
        });
        Assert.Equal(200, res.StatusCode);

        var tagged = res.Res!.OpenUnionDateTimeEnvelope.Tagged!;
        Assert.True(tagged.IsUnknown());
        var taggedRaw = (Newtonsoft.Json.Linq.JObject)tagged.UnknownRaw!;
        Assert.Equal("paused", (string?)taggedRaw["kind"]);
        Assert.Equal(OffsetTimestamp, (string?)taggedRaw["occurredAt"]);

        var untagged = res.Res!.OpenUnionDateTimeEnvelope.Untagged!;
        Assert.True(untagged.IsUnknown());
        Assert.Equal(OffsetTimestamp, (string?)((Newtonsoft.Json.Linq.JObject)untagged.UnknownRaw!)["gammaAt"]);
    }

    [Fact]
    public async Task OpenUnionDateTimeEnvelope_KnownVariants_RoundTrip()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, apiKeyAuth: "Bearer testToken");
        var instant = new DateTime(2024, 1, 2, 3, 4, 5, DateTimeKind.Utc);

        var res = await sdk.Unions.OpenUnionDateTimePostAsync(new OpenUnionDateTimeEnvelope {
            Tagged = OpenDateTimeTaggedUnion.CreateStarted(new DateTimeEventStarted {
                OccurredAt = instant,
            }),
            Untagged = OpenDateTimeUntaggedUnion.CreateDateTimeRecordBeta(new DateTimeRecordBeta {
                BetaAt = instant,
                Count = 2,
            }),
        });
        Assert.Equal(200, res.StatusCode);

        var tagged = res.Res!.OpenUnionDateTimeEnvelope.Tagged!;
        Assert.False(tagged.IsUnknown());
        Assert.Equal(instant, tagged.DateTimeEventStarted!.OccurredAt.ToUniversalTime());

        var untagged = res.Res!.OpenUnionDateTimeEnvelope.Untagged!;
        Assert.False(untagged.IsUnknown());
        Assert.Equal(2, untagged.DateTimeRecordBeta!.Count);
    }
}
