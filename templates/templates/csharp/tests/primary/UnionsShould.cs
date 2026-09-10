using System;
using System.Collections.Generic;
using System.Net;
using System.Threading.Tasks;
using System.Web;
using FluentAssertions;
using Newtonsoft.Json;
using Openapi.Models.Errors;
using Openapi.Models.Operations;
using Openapi.Models.Shared;
using Openapi.Utils;
using Xunit;

public class UnionsShould
{
    [Fact]
    public async Task StronglyTypedOneOfPost_Basic()
    {
        CommonHelpers.RecordTest("unions-strongly-typed-one-of-post-basic");
        var s = new Openapi.SDK();

        var obj = Helpers.CreateSimpleObjectWithType();

        var req = StronglyTypedOneOfObject.CreateSimpleObjectWithType(obj);

        var res = await s.Unions.StronglyTypedOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            StronglyTypedOneOfObjectType.SimpleObjectWithType,
            res.Res.Json.Type.ToString()
        );
        Helpers.AssertSimpleObjectWithType(res.Res.Json.SimpleObjectWithType);
    }

    [Fact]
    public async Task StronglyTypedNullableOneOfPost()
    {
        CommonHelpers.RecordTest("unions-strongly-typed-nullable-one-of-post");
        var s = new Openapi.SDK();

        // option 1 - NestedEnumArray
        var obj1 = new Schemas()
        {
            Enums = new List<Openapi.Models.Shared.Enum>()
            {
                Openapi.Models.Shared.Enum.One,
                Openapi.Models.Shared.Enum.Two,
            }
        };

        var req = StronglyTypedNullableOneOfObject.CreateNestedEnumArray(obj1);
        var res = await s.Unions.StronglyTypedNullableOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            StronglyTypedNullableOneOfObjectType.NestedEnumArray,
            res.Res.Json.Type.ToString()
        );
        res.Res.GetJsonNestedEnumArray().Should().BeEquivalentTo(obj1);

        // option 2 - NestedEnumMap
        var obj2 = new NestedEnumMapSchemas()
        {
            Enums = new Dictionary<string, Openapi.Models.Shared.Enum>()
            {
                { "key2", Openapi.Models.Shared.Enum.Two },
                { "key3", Openapi.Models.Shared.Enum.Three },
            },
        };

        req = StronglyTypedNullableOneOfObject.CreateNestedEnumMap(obj2);
        res = await s.Unions.StronglyTypedNullableOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            StronglyTypedNullableOneOfObjectType.NestedEnumMap,
            res.Res.Json.Type.ToString()
        );

        // option 3 - null
        req = StronglyTypedNullableOneOfObject.CreateNull();
        res = await s.Unions.StronglyTypedNullableOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Null(res.Res.Json);
    }

    [Fact]
    public async Task CollectionOneOfPost()
    {
        CommonHelpers.RecordTest("unions-collections-one-of-post");
        var s = new Openapi.SDK();

        var req = CollectionOneOfObject.CreateArrayOfAny(new List<object> { "one", "two" });

        var res = await s.Unions.CollectionOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        res.Res.Json.ArrayOfAny.Should().BeEquivalentTo(req.ArrayOfAny);

        var req2 = CollectionOneOfObject.CreateMapOfAny(
            new Dictionary<string, object>() { { "1", "one" }, { "2", "two" } }
        );

        var res2 = await s.Unions.CollectionOneOfPostAsync(req2);
        Assert.Equal(HttpStatusCode.OK, res2.HttpMeta.Response.StatusCode);
        res.Res.Json.MapOfAny.Should().BeEquivalentTo(req.MapOfAny);
    }

    [Fact]
    public async Task StronglyTypedOneOfPostWithNonStandardDiscriminatorName()
    {
        CommonHelpers.RecordTest(
            "unions-strongly-typed-one-of-post-with-non-standard-discriminator-name"
        );
        var s = new Openapi.SDK();

        var obj = new SimpleObjectWithNonStandardTypeName
        {
            Str = "test",
            Bool = true,
            Int = 1,
            Int32 = 1,
            IntEnum = SimpleObjectWithNonStandardTypeNameIntEnum.Second,
            Int32Enum = SimpleObjectWithNonStandardTypeNameInt32Enum.FiftyFive,
            Num = 1.1,
            Float32 = 1.1f,
            Enum = Openapi.Models.Shared.Enum.One,
            Any = "any",
            Date = DateOnly.FromDateTime(System.DateTime.Parse("2020-01-01")),
            DateTime = DateTime.Parse("2020-01-01T00:00:00.0000001Z").ToUniversalTime(),
            BoolOpt = true,
            StrOpt = "testOptional",
            IntOptNull = null,
            NumOptNull = null
        };

        var req =
            StronglyTypedOneOfObjectWithNonStandardDiscriminatorName.CreateSimpleObjectWithNonStandardTypeName(
                obj
            );

        var res = await s.Unions.StronglyTypedOneOfPostWithNonStandardDiscriminatorNameAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            StronglyTypedOneOfObjectWithNonStandardDiscriminatorNameType.SimpleObjectWithNonStandardTypeName,
            res.Res.Json.Type.ToString()
        );
        res.Res.Json.SimpleObjectWithNonStandardTypeName.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task StronglyTypedOneOfPost_Deep()
    {
        CommonHelpers.RecordTest("unions-strongly-typed-one-of-post-deep");
        var s = new Openapi.SDK();

        var obj = new DeepObjectWithType
        {
            Any = DeepObjectWithTypeAny.CreateSimpleObject(Helpers.CreateSimpleObject()),
            Arr = new List<SimpleObject>
            {
                Helpers.CreateSimpleObject(),
                Helpers.CreateSimpleObject()
            },
            Bool = true,
            Int = 1,
            Map = new Dictionary<string, SimpleObject> { { "key", Helpers.CreateSimpleObject() } },
            Num = 1.1,
            Obj = Helpers.CreateSimpleObject(),
            Str = "test",
            Type = "deepObjectWithType"
        };

        var req = StronglyTypedOneOfObject.CreateDeepObjectWithType(obj);

        var res = await s.Unions.StronglyTypedOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(StronglyTypedOneOfObjectType.DeepObjectWithType, res.Res.Json.Type.ToString());
        res.Res.Json.DeepObjectWithType.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task WeaklyTypedOneOfPost_Basic()
    {
        CommonHelpers.RecordTest("unions-weakly-typed-one-of-post-basic");
        var s = new Openapi.SDK();

        var obj = Helpers.CreateSimpleObject();

        var req = WeaklyTypedOneOfObject.CreateSimpleObject(obj);

        var res = await s.Unions.WeaklyTypedOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(WeaklyTypedOneOfObjectType.SimpleObject, res.Res.Json.Type.ToString());
        res.Res.Json.SimpleObject.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task WeaklyTypedOneOfPost_Deep()
    {
        CommonHelpers.RecordTest("unions-weakly-typed-one-of-post-deep");
        var s = new Openapi.SDK();

        var obj = Helpers.CreateDeepObject();

        var req = WeaklyTypedOneOfObject.CreateDeepObject(obj);

        var res = await s.Unions.WeaklyTypedOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(WeaklyTypedOneOfObjectType.DeepObject, res.Res.Json.Type.ToString());
        res.Res.Json.DeepObject.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task TypedObjectOneOfPost_Obj1()
    {
        CommonHelpers.RecordTest("unions-typed-object-one-of-post-obj1");
        var s = new Openapi.SDK();

        var obj = new TypedObject1 { Type = TypedObject1Type.Obj1, Value = "v1" };

        var req = TypedObjectOneOf.CreateObj1(obj);

        var res = await s.Unions.TypedObjectOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(TypedObjectOneOfType.Obj1, res.Res.Json.Type.ToString());
        Assert.NotNull(res.Res.Json.TypedObject1);
        Assert.Equal("v1", res.Res.Json.TypedObject1.Value);
        res.Res.Json.TypedObject1.Should().BeEquivalentTo(obj);

        // Omitting spec-required `value` field causes response validation error
        var missingValue = new TypedObject1 { Type = TypedObject1Type.Obj1 };
        var missingReq = TypedObjectOneOf.CreateObj1(missingValue);
        await Assert.ThrowsAsync<Openapi.Models.Errors.ResponseValidationException>(
            () => s.Unions.TypedObjectOneOfPostAsync(missingReq)
        );
    }

    [Fact]
    public async Task TypedObjectOneOfPost_Obj2()
    {
        CommonHelpers.RecordTest("unions-typed-object-one-of-post-obj2");
        var s = new Openapi.SDK();

        var obj = new TypedObject2 { Type = TypedObject2Type.Obj2, Value = "v2" };

        var req = TypedObjectOneOf.CreateObj2(obj);

        var res = await s.Unions.TypedObjectOneOfPostAsync(req);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(TypedObjectOneOfType.Obj2, res.Res.Json.Type.ToString());
        Assert.NotNull(res.Res.Json.TypedObject2);
        Assert.Equal("v2", res.Res.Json.TypedObject2.Value);
        res.Res.Json.TypedObject2.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task TypedObjectOneOfPost_Obj3()
    {
        CommonHelpers.RecordTest("unions-typed-object-one-of-post-obj3");
        var s = new Openapi.SDK();

        var obj = new TypedObject3 { Type = TypedObject3Type.Obj3, Value = "v3" };

        var req = TypedObjectOneOf.CreateObj3(obj);

        var res = await s.Unions.TypedObjectOneOfPostAsync(req);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(TypedObjectOneOfType.Obj3, res.Res.Json.Type.ToString());
        Assert.NotNull(res.Res.Json.TypedObject3);
        Assert.Equal("v3", res.Res.Json.TypedObject3.Value);
        res.Res.Json.TypedObject3.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public void TypedObjectOneOfPost_Null()
    {
        CommonHelpers.RecordTest("unions-typed-object-one-of-post-null");

        // The type system prevents creating a TypedObjectOneOf with Type = "null";
        var method = typeof(TypedObjectOneOf).GetMethod("CreateNull");
        Assert.Null(method);

        // Deserialization should fail since TypedObjectOneOf is non-nullable. TypedObjectOneOf
        // is generated open so its converter tolerates null (Unknown fallback): the failure
        // surfaces from the envelope's Required.Always presence validation on the `json` property.
        var ex = Assert.Throws<JsonSerializationException>(
            () => ResponseBodyDeserializer.DeserializeNotNull<TypedObjectOneOfPostRes>(
                "{\"json\": null}",
                NullValueHandling.Ignore
            )
        );

        Assert.StartsWith("Required property 'json' expects a value but got null", ex.Message);
    }

    [Fact]
    public void TypedObjectOneOfPost_BodyLessThrowsOnSerialize()
    {
        CommonHelpers.RecordTest("unions-typed-object-one-of-post-body-less-throws");

        // A discriminated union built with only the discriminator set (no variant body)
        // must not silently serialize a required field to null; it must throw.
        var bodyLess = new TypedObjectOneOf(TypedObjectOneOfType.Obj1);

        var ex = Assert.Throws<InvalidOperationException>(
            () => JsonConvert.SerializeObject(bodyLess)
        );
        Assert.Contains("no variant value was set", ex.Message);
    }

    [Fact]
    public async Task TypedObjectNullableOneOfPost_Obj1()
    {
        CommonHelpers.RecordTest("unions-typed-object-nullable-one-of-post-obj1");
        var s = new Openapi.SDK();

        var obj = new TypedObject1 { Type = TypedObject1Type.Obj1, Value = "v1" };

        var req = TypedObjectNullableOneOf.CreateObj1(obj);

        var res = await s.Unions.TypedObjectNullableOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(TypedObjectOneOfType.Obj1, res.Res.Json.Type.ToString());
        Assert.NotNull(res.Res.Json.TypedObject1);
        Assert.Equal("v1", res.Res.Json.TypedObject1.Value);
        res.Res.Json.TypedObject1.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task TypedObjectNullableOneOfPost_Obj2()
    {
        CommonHelpers.RecordTest("unions-typed-object-nullable-one-of-post-obj2");
        var s = new Openapi.SDK();

        var obj = new TypedObject2 { Type = TypedObject2Type.Obj2, Value = "v2" };

        var req = TypedObjectNullableOneOf.CreateObj2(obj);

        var res = await s.Unions.TypedObjectNullableOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(TypedObjectOneOfType.Obj2, res.Res.Json.Type.ToString());
        Assert.NotNull(res.Res.Json.TypedObject2);
        Assert.Equal("v2", res.Res.Json.TypedObject2.Value);
        res.Res.Json.TypedObject2.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task TypedObjectNullableOneOfPost_Null()
    {
        CommonHelpers.RecordTest("unions-typed-object-nullable-one-of-post-null");
        var s = new Openapi.SDK();

        var req = TypedObjectNullableOneOf.CreateNull();

        var res = await s.Unions.TypedObjectNullableOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Null(res.Res.Json);
    }

    [Fact]
    public async Task FlattenedTypedObject_Obj1()
    {
        CommonHelpers.RecordTest("unions-flattened-typed-object-post-obj1");
        var s = new Openapi.SDK();

        var obj = FlattenedTypedObject1.CreateTypedObject1(
            new TypedObject1 { Value = "one", Type = TypedObject1Type.Obj1 }
        );

        var res = await s.Unions.FlattenedTypedObjectPostAsync(obj);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        res.Res.Json.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task NullableTypedObjectPost_Obj1()
    {
        CommonHelpers.RecordTest("unions-nullable-typed-object-post-obj1");
        var s = new Openapi.SDK();

        var obj = new TypedObject1 { Value = "one", Type = TypedObject1Type.Obj1 };

        var res = await s.Unions.NullableTypedObjectPostAsync(obj);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        res.Res.Json.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task NullableTypedObjectPost_Null()
    {
        CommonHelpers.RecordTest("unions-nullable-typed-object-post-null");
        var s = new Openapi.SDK();

        var res = await s.Unions.NullableTypedObjectPostAsync(null);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Null(res.Res.Json);
    }

    [Fact]
    public async Task NullableOneOfSchemaPost_Obj1()
    {
        CommonHelpers.RecordTest("unions-nullable-oneof-schema-post-obj1");
        var s = new Openapi.SDK();

        var obj = new TypedObject1 { Value = "one", Type = TypedObject1Type.Obj1 };

        var req = NullableOneOfSchemaPostRequestBody.CreateObj1(obj);

        var res = await s.Unions.NullableOneOfSchemaPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(NullableOneOfSchemaPostJsonType.Obj1, res.Res.Json.Type.ToString());
        res.Res.Json.TypedObject1.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task NullableOneOfSchemaPost_Obj2()
    {
        CommonHelpers.RecordTest("unions-nullable-oneof-schema-post-obj2");
        var s = new Openapi.SDK();

        var obj = new TypedObject2 { Value = "two", Type = TypedObject2Type.Obj2 };

        var req = NullableOneOfSchemaPostRequestBody.CreateObj2(obj);

        var res = await s.Unions.NullableOneOfSchemaPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(NullableOneOfSchemaPostJsonType.Obj2, res.Res.Json.Type.ToString());
        res.Res.Json.TypedObject2.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task NullableOneOfSchemaPost_Null()
    {
        CommonHelpers.RecordTest("unions-nullable-oneof-schema-post-null");
        var s = new Openapi.SDK();

        var res = await s.Unions.NullableOneOfSchemaPostAsync(null);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Null(res.Res.Json);
    }

    [Fact]
    public async Task NullableOneOfTypeInObject()
    {
        CommonHelpers.RecordTest("unions-nullable-oneof-type-in-object-post");

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

        var s = new Openapi.SDK();
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
            Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
            res.Res.Json.Should().BeEquivalentTo(req);
        }
    }

    [Fact]
    public async Task NullableOneOfRefInObject()
    {
        CommonHelpers.RecordTest("unions-nullable-oneof-ref-in-object-post");

        var tests = new CommonHelpers.TestTableEntry[]
        {
            new CommonHelpers.TestTableEntry
            {
                name = "Non-nullable field set only",
                arg = new NullableOneOfRefInObject
                {
                    OneOfOne = OneOfOne.CreateTypedObject1(
                        new TypedObject1 { Value = "one", Type = TypedObject1Type.Obj1 }
                    )
                },
                want =
                    "{\"NullableOneOfOne\":null,\"NullableOneOfTwo\":null,\"OneOfOne\":{\"type\":\"obj1\",\"value\":\"one\"}}"
            },
            new CommonHelpers.TestTableEntry
            {
                name = "Nullable fields set to null",
                arg = new NullableOneOfRefInObject
                {
                    NullableOneOfOne = null,
                    NullableOneOfTwo = null,
                    OneOfOne = OneOfOne.CreateTypedObject1(
                        new TypedObject1 { Value = "one", Type = TypedObject1Type.Obj1 }
                    )
                },
                want =
                    "{\"NullableOneOfOne\":null,\"NullableOneOfTwo\":null,\"OneOfOne\":{\"type\":\"obj1\",\"value\":\"one\"}}"
            },
            new CommonHelpers.TestTableEntry
            {
                name = "All fields set to non-null values",
                arg = new NullableOneOfRefInObject
                {
                    NullableOneOfOne = new TypedObject1
                    {
                        Value = "one",
                        Type = TypedObject1Type.Obj1
                    },
                    NullableOneOfTwo = NullableOneOfTwo.CreateObj2(
                        new TypedObject2 { Value = "two", Type = TypedObject2Type.Obj2 }
                    ),
                    OneOfOne = OneOfOne.CreateTypedObject1(
                        new TypedObject1 { Value = "", Type = TypedObject1Type.Obj1 }
                    )
                },
                want =
                    "{\"NullableOneOfOne\":{\"type\":\"obj1\",\"value\":\"one\"},\"NullableOneOfTwo\":{\"type\":\"obj2\",\"value\":\"two\"},\"OneOfOne\":{\"type\":\"obj1\",\"value\":\"\"}}",
            }
        };
        var s = new Openapi.SDK();
        foreach (var test in tests)
        {
            NullableOneOfRefInObject req = (NullableOneOfRefInObject)test.arg;
            var serializedBody = RequestBodySerializer.Serialize(
                req,
                "Request",
                "json",
                true,
                false
            );
            var json = await serializedBody.ReadAsStringAsync();
            Assert.Equal(test.want, json);
            var res = await s.Unions.NullableOneOfRefInObjectPostAsync(req);
            Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
            res.Res.Json.Should().BeEquivalentTo(req);
        }
    }

    [Fact]
    public async Task PrimitiveTypeOneOfPost_String()
    {
        CommonHelpers.RecordTest("unions-primitive-type-one-of-post-string");
        var s = new Openapi.SDK();

        var req = PrimitiveTypeOneOfPostRequestBody.CreateStr("test");

        var res = await s.Unions.PrimitiveTypeOneOfPostAsync(req);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(PrimitiveTypeOneOfPostJsonType.Str, res.Res.Json.Type.ToString());
        Assert.Equal("test", res.Res.Json.Str);
    }

    [Fact]
    public async Task PrimitiveTypeOneOfPost_Integer()
    {
        CommonHelpers.RecordTest("unions-primitive-type-one-of-post-integer");
        var s = new Openapi.SDK();

        var req = PrimitiveTypeOneOfPostRequestBody.CreateInteger(111);

        var res = await s.Unions.PrimitiveTypeOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(PrimitiveTypeOneOfPostJsonType.Integer, res.Res.Json.Type.ToString());
        Assert.Equal(111, res.Res.Json.Integer);
    }

    [Fact]
    public async Task PrimitiveTypeOneOfPost_Number()
    {
        CommonHelpers.RecordTest("unions-primitive-type-one-of-post-number");
        var s = new Openapi.SDK();

        var req = PrimitiveTypeOneOfPostRequestBody.CreateNumber(22.2);

        var res = await s.Unions.PrimitiveTypeOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(PrimitiveTypeOneOfPostJsonType.Number, res.Res.Json.Type.ToString());
        Assert.Equal(22.2, res.Res.Json.Number);
    }

    [Fact]
    public async Task PrimitiveTypeOneOfPost_Boolean()
    {
        CommonHelpers.RecordTest("unions-primitive-type-one-of-post-boolean");
        var s = new Openapi.SDK();

        var req = PrimitiveTypeOneOfPostRequestBody.CreateBoolean(true);

        var res = await s.Unions.PrimitiveTypeOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(PrimitiveTypeOneOfPostJsonType.Boolean, res.Res.Json.Type.ToString());
        Assert.Equal(true, res.Res.Json.Boolean);
    }

    [Fact]
    public async Task MixedTypeOneOfPost_String()
    {
        CommonHelpers.RecordTest("unions-mixed-type-one-of-post-string");
        var s = new Openapi.SDK();

        var req = MixedTypeOneOfPostRequestBody.CreateStr("test");

        var res = await s.Unions.MixedTypeOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(MixedTypeOneOfPostJsonType.Str, res.Res.Json.Type.ToString());
        Assert.Equal("test", res.Res.Json.Str);
    }

    [Fact]
    public async Task MixedTypeOneOfPost_Integer()
    {
        CommonHelpers.RecordTest("unions-mixed-type-one-of-post-integer");
        var s = new Openapi.SDK();

        var req = MixedTypeOneOfPostRequestBody.CreateInteger(111);

        var res = await s.Unions.MixedTypeOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(MixedTypeOneOfPostJsonType.Integer, res.Res.Json.Type.ToString());
        Assert.Equal(111, res.Res.Json.Integer);
    }

    [Fact]
    public async Task MixedTypeOneOfPost_Object()
    {
        CommonHelpers.RecordTest("unions-mixed-type-one-of-post-object");
        var s = new Openapi.SDK();

        var obj = Helpers.CreateSimpleObject();

        var req = MixedTypeOneOfPostRequestBody.CreateSimpleObject(obj);

        var res = await s.Unions.MixedTypeOneOfPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(MixedTypeOneOfPostJsonType.SimpleObject, res.Res.Json.Type.ToString());
        res.Res.Json.SimpleObject.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task DateNullUnion()
    {
        CommonHelpers.RecordTest("unions-date-null");
        var s = new Openapi.SDK();

        var date = DateOnly.FromDateTime(System.DateTime.Parse("2020-01-01"));

        var res = await s.Unions.UnionDateNullAsync(date);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(date, res.Res.Json);
    }

    [Fact]
    public async Task DateTimeNullUnion()
    {
        CommonHelpers.RecordTest("unions-datetime-null");
        var s = new Openapi.SDK();

        var dateTime = System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime();

        var res = await s.Unions.UnionDateTimeNullAsync(dateTime);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(dateTime, res.Res.Json);
    }

    [Fact]
    public async Task DateTimeBigintUnion()
    {
        CommonHelpers.RecordTest("unions-datetime-bigint");

        // NB: The date-time payload arrives as a string token and is read with
        // DateParseHandling.None), so the DateTime member scores Inexact rather
        // than Matched. The winner is still DateTime since the bigint member fails
        // to deserialize the string and drops out.
        var s = new Openapi.SDK();

        var req = UnionDateTimeBigIntRequestBody.CreateDateTime(
            System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime()
        );
        var res = await s.Unions.UnionDateTimeBigIntAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(UnionDateTimeBigIntRequestBodyType.DateTime, res.Res.Json.Type.ToString());
        Assert.Equal(
            System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime(),
            res.Res.Json.DateTime
        );

        var nextRes = await s.Unions.UnionDateTimeBigIntAsync(
            UnionDateTimeBigIntRequestBody.CreateBigint(9007199254740991)
        );
        Assert.Equal(HttpStatusCode.OK, nextRes.HttpMeta.Response.StatusCode);
        Assert.Equal(UnionDateTimeBigIntRequestBodyType.Bigint, nextRes.Res.Json.Type);
        Assert.Equal(9007199254740991, nextRes.Res.Json.Bigint);
    }

    [Fact]
    public async Task UnionBigIntStrDecimal()
    {
        CommonHelpers.RecordTest("unions-bigint-str-decimal");
        var s = new Openapi.SDK();

        var req = UnionBigIntStrDecimalRequestBody.CreateDecimal(3.141592653589793M);
        var json = await Helpers.GetSerializedBodyJson(req);
        Assert.Equal("3.141592653589793", json);

        var res = await s.Unions.UnionBigIntStrDecimalAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(UnionBigIntStrDecimalRequestBodyType.Decimal, res.Res.Json.Type);
        Assert.Equal(3.141592653589793M, res.Res.Json.Decimal);

        req = UnionBigIntStrDecimalRequestBody.CreateBigint(9223372036854775807);
        json = await Helpers.GetSerializedBodyJson(req);
        Assert.Equal("\"9223372036854775807\"", json);

        res = await s.Unions.UnionBigIntStrDecimalAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(UnionBigIntStrDecimalRequestBodyType.Bigint, res.Res.Json.Type);
        Assert.Equal(9223372036854775807, res.Res.Json.Bigint);
    }

    [Fact]
    public async Task UnionMap()
    {
        CommonHelpers.RecordTest("unions-union-map");
        var s = new Openapi.SDK();

        var res = await s.Unions.UnionMapAsync(new UnionMapRequestBody
        {
            Input = new Dictionary<string, OneOfPrimitives>{
                    {"str", OneOfPrimitives.CreateStr("test")},
                    {"bool", OneOfPrimitives.CreateBoolean(true)}
                }
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("test", res.Res.Json.Input["str"].Str);
        Assert.Equal(true, res.Res.Json.Input["bool"].Boolean);
    }

    [Fact]
    public async Task ConstDiscriminator()
    {
        CommonHelpers.RecordTest("unions-const-discriminator");
        var s = new Openapi.SDK();
        var obj = new ConstObject1 { ImageURL = "http://boo" };
        var req = ConstDiscriminatedOneOf.CreateTag1(obj);
        var res = await s.Unions.ConstDiscriminatedOneOfAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        res.Res.Json.ConstObject1.Should().BeEquivalentTo(obj);
    }

    [Fact]
    public async Task StronglyTypedOneOfDiscriminatedPost()
    {
        CommonHelpers.RecordTest("unions-strongly-typed-one-of-discriminated-post");
        var s = new Openapi.SDK();

        var obj = new TaggedObject1
        {
            ImageURL = "http://example.com/image.png",
            Tag = Openapi.Models.Shared.Tag.Tag1
        };
        var req = StronglyTypedOneOfDiscriminatedObject.CreateTag1(obj);

        var res = await s.Unions.StronglyTypedOneOfDiscriminatedPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        res.Res.Json.TaggedObject1.Should().BeEquivalentTo(obj);

        var obj2 = new TaggedObject2
        {
            ProfileId = "abc123",
            Tag = TaggedObject2Tag.Tag2
        };
        var res2 = await s.Unions.StronglyTypedOneOfDiscriminatedPostAsync(
            StronglyTypedOneOfDiscriminatedObject.CreateTag2(obj2));
        Assert.Equal(HttpStatusCode.OK, res2.HttpMeta.Response.StatusCode);
        res2.Res.Json.TaggedObject2.Should().BeEquivalentTo(obj2);

        var obj3 = new TaggedObject3
        {
            Phone = "+15555550123"
        };
        var res3 = await s.Unions.StronglyTypedOneOfDiscriminatedPostAsync(
            StronglyTypedOneOfDiscriminatedObject.CreateTag3(obj3));
        Assert.Equal(HttpStatusCode.OK, res3.HttpMeta.Response.StatusCode);
        res3.Res.Json.TaggedObject3.Should().BeEquivalentTo(obj3);
    }

    [Fact]
    public async Task UnionExtraJsonProperties()
    {
        CommonHelpers.RecordTest("unions-extra-json-properties");
        var s = new Openapi.SDK();

        var res = await s.Unions.OneOfOverlappingObjectsAsync(new OneOfOverlappingObjectsRequestBody
        {
            Field1 = "test1", // common to Obj1 and Obj2
            Field3 = 1  // not part of Obj1 nor Obj2
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(OneOfOverlappingObjectsType.Obj1, res.Res.Json.Type);
        Assert.Equal("test1", res.Res.Json.Obj1.Field1);


        res = await s.Unions.OneOfOverlappingObjectsAsync(new OneOfOverlappingObjectsRequestBody
        {
            Field1 = "test2",  // common to Obj1 and Obj2
            Field2 = true,  // exclusive to Obj2
            Field3 = 1  // not part of Obj1 nor Obj2
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(OneOfOverlappingObjectsType.Obj2, res.Res.Json.Type);
        Assert.Equal("test2", res.Res.Json.Obj2.Field1);
        Assert.True(res.Res.Json.Obj2.Field2);
    }

    [Fact]
    public async Task UnionNestedEnumsForm()
    {
        CommonHelpers.RecordTest("unions-nested-enums-form");

        var sdk = new Openapi.SDK();

        var obj1 = new NestedEnumArray()
        {
            Tags = "one,two",
            Enums = new List<Openapi.Models.Shared.Enum>()
            {
                Openapi.Models.Shared.Enum.One,
                Openapi.Models.Shared.Enum.Two,
            }
        };

        var req1 = UnionNestedEnumsFormRequestBody.CreateNestedEnumArray(obj1);
        var serializedForm1 = RequestBodySerializer.Serialize(req1, "Request", "form", false, false);
        var form1 = HttpUtility.UrlDecode(await serializedForm1.ReadAsStringAsync());
        Assert.Equal("enums[]=one&enums[]=two&tags=one,two", form1);

        var res1 = await sdk.Unions.UnionNestedEnumsFormAsync(req1);
        Assert.Equal(HttpStatusCode.OK, res1.HttpMeta.Response.StatusCode);

        var obj2 = new NestedEnumMap()
        {
            Tags = "two,three",
            Enums = new Dictionary<string, Openapi.Models.Shared.Enum>()
            {
                { "key2", Openapi.Models.Shared.Enum.Two },
                { "key3", Openapi.Models.Shared.Enum.Three },
            },
        };

        var req2 = UnionNestedEnumsFormRequestBody.CreateNestedEnumMap(obj2);
        var serializedForm2 = RequestBodySerializer.Serialize(req2, "Request", "form", false, false);
        var form2 = HttpUtility.UrlDecode(await serializedForm2.ReadAsStringAsync());
        Assert.Equal("enums={\"key2\":\"two\",\"key3\":\"three\"}&tags=two,three", form2);

        var res2 = await sdk.Unions.UnionNestedEnumsFormAsync(req2);
        Assert.Equal(HttpStatusCode.OK, res2.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task UnionNestedEnumsMultipart()
    {
        CommonHelpers.RecordTest("unions-nested-enums-multipart");

        var sdk = new Openapi.SDK();

        var enumArray = new List<Openapi.Models.Shared.Enum>()
        {
            Openapi.Models.Shared.Enum.One,
            Openapi.Models.Shared.Enum.Two,
        };

        var req1 = new UnionNestedEnumsMultipartRequestBody() { Enums = Enums.CreateArrayOfEnum(enumArray) };
        var serializedForm1 = RequestBodySerializer.Serialize(req1, "Request", "multipart", false, false);
        var form1 = await serializedForm1.ReadAsStringAsync();
        Assert.Contains("Content-Disposition: form-data; name=enums", form1);
        Assert.Contains("[\"one\",\"two\"]", form1);

        var res1 = await sdk.Unions.UnionNestedEnumsMultipartAsync(req1);
        Assert.Equal(HttpStatusCode.OK, res1.HttpMeta.Response.StatusCode);

        var enumMap = new Dictionary<string, Openapi.Models.Shared.Enum>()
        {
            { "key2", Openapi.Models.Shared.Enum.Two },
            { "key3", Openapi.Models.Shared.Enum.Three },
        };

        var req2 = new UnionNestedEnumsMultipartRequestBody() { Enums = Enums.CreateMapOfEnum(enumMap) };
        var serializedForm2 = RequestBodySerializer.Serialize(req2, "Request", "multipart", false, false);
        var form2 = await serializedForm2.ReadAsStringAsync();
        Assert.Contains("Content-Disposition: form-data; name=enums", form2);
        Assert.Contains("{\"key2\":\"two\",\"key3\":\"three\"}", form2);

        var res2 = await sdk.Unions.UnionNestedEnumsMultipartAsync(req2);
        Assert.Equal(HttpStatusCode.OK, res2.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task UnionOfArraysTest()
    {
        CommonHelpers.RecordTest("unions-union-of-arrays");
        var s = new Openapi.SDK();

        var arrayOf1 = new List<Openapi.Models.Shared.One> { new Openapi.Models.Shared.One { Foo = "foo0" }, new Openapi.Models.Shared.One { Foo = "foo1" } };
        var req1 = UnionOfArrays.CreateArrayOf1(arrayOf1);
        var res1 = await s.Unions.UnionOfArraysPostAsync(req1);

        Assert.Equal(HttpStatusCode.OK, res1.HttpMeta.Response.StatusCode);
        Assert.Equal(UnionOfArraysType.ArrayOf1, res1.Res.Json.Type);
        Assert.Equal(2, res1.Res.Json.ArrayOf1.Count);
        Assert.Equal("foo0", res1.Res.Json.ArrayOf1[0].Foo);
        Assert.Equal("foo1", res1.Res.Json.ArrayOf1[1].Foo);

        var arrayOf2 = new List<UnionOfArrays2> {
            new UnionOfArrays2 { Bar = "bar0" },
            new UnionOfArrays2 { Bar = "bar1" }
        };
        var req2 = UnionOfArrays.CreateArrayOfUnionOfArrays2(arrayOf2);
        var res2 = await s.Unions.UnionOfArraysPostAsync(req2);

        Assert.Equal(HttpStatusCode.OK, res2.HttpMeta.Response.StatusCode);
        Assert.Equal(UnionOfArraysType.ArrayOfUnionOfArrays2, res2.Res.Json.Type);
        Assert.Equal(2, res2.Res.Json.ArrayOfUnionOfArrays2.Count);
        Assert.Equal("bar0", res2.Res.Json.ArrayOfUnionOfArrays2[0].Bar);
        Assert.Equal("bar1", res2.Res.Json.ArrayOfUnionOfArrays2[1].Bar);

        var arrayOf3 = new List<Three> { new Three { Baz = "baz0" }, new Three { Baz = "baz1" } };
        var req3 = UnionOfArrays.CreateArrayOf3(arrayOf3);
        var res3 = await s.Unions.UnionOfArraysPostAsync(req3);

        Assert.Equal(HttpStatusCode.OK, res3.HttpMeta.Response.StatusCode);
        Assert.Equal(UnionOfArraysType.ArrayOf3, res3.Res.Json.Type);
        Assert.Equal(2, res3.Res.Json.ArrayOf3.Count);
        Assert.Equal("baz0", res3.Res.Json.ArrayOf3[0].Baz);
        Assert.Equal("baz1", res3.Res.Json.ArrayOf3[1].Baz);
    }

    [Fact]
    public async Task NestedDiscUnionTest()
    {
        CommonHelpers.RecordTest("unions-nested-discriminated-union");

        var s = new Openapi.SDK();

        var req1 = NestedDiscUnion.CreateA(
            TypeA.CreateTypeA1(new TypeA1() { Value = "something" }));
        var res1 = await s.Unions.NestedDiscUnionAsync(req1);

        Assert.Equal(HttpStatusCode.OK, res1.HttpMeta.Response.StatusCode);
        Assert.Equal(NestedDiscUnionType.A, res1.Res.Json.Type);
        Assert.Equal(TypeAType.TypeA1, res1.Res.Json.TypeA.Type);
        Assert.Equal("something", res1.Res.Json.TypeA.TypeA1.Value);

        var req2 = NestedDiscUnion.CreateA(
            TypeA.CreateTypeA2(new TypeA2() { Count = 42 }));
        var res2 = await s.Unions.NestedDiscUnionAsync(req2);

        Assert.Equal(HttpStatusCode.OK, res2.HttpMeta.Response.StatusCode);
        Assert.Equal(NestedDiscUnionType.A, res2.Res.Json.Type);
        Assert.Equal(TypeAType.TypeA2, res2.Res.Json.TypeA.Type);
        Assert.Equal(42, res2.Res.Json.TypeA.TypeA2.Count);

        var req3 = NestedDiscUnion.CreateB(new TypeB() { Data = "something else" });
        var res3 = await s.Unions.NestedDiscUnionAsync(req3);

        Assert.Equal(HttpStatusCode.OK, res3.HttpMeta.Response.StatusCode);
        Assert.Equal(NestedDiscUnionType.B, res3.Res.Json.Type);
        Assert.Equal("something else", res3.Res.Json.TypeB.Data);
    }

    [Fact]
    public async Task ConflictingDiscriminatorMappingKey()
    {
        CommonHelpers.RecordTest("unions-conflicting-discriminator-mapping-key");
        var s = new Openapi.SDK();

        var taggedObj = new TaggedObject1
        {
            ImageURL = "https://example.com/image.png",
            Tag = Openapi.Models.Shared.Tag.Tag1
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
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(StronglyTypedOneOfDiscriminatedObjectType.Tag1, res.Res.Json.Prefix.Type);
        Assert.Equal(taggedObj.ImageURL, res.Res.Json.Prefix.TaggedObject1.ImageURL);
        Assert.Equal(taggedObj.Tag, res.Res.Json.Prefix.TaggedObject1.Tag);
        Assert.Equal("test", res.Res.Json.PrefixTag2.Str);
    }

    [Fact]
    public async Task CircularReferenceRecursiveOneOf()
    {
        CommonHelpers.RecordTest("unions-circular-reference-recursive-one-of");

        var sdk = new Openapi.SDK(serverUrl: Helpers.HttpBinUrl);

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

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Object);
        Assert.Equal(
            JsonConvert.SerializeObject(payload),
            JsonConvert.SerializeObject(res.Object!.Json.Value)
        );
    }

    [Fact]
    public async Task ArrayOfDiscriminatedUnions()
    {
        CommonHelpers.RecordTest("unions-array-of-discriminated-unions");
        var sdk = new Openapi.SDK(serverUrl: Helpers.HttpBinUrl);

        var simpleObject = Helpers.CreateSimpleObjectWithType();

        var req = new List<StronglyTypedOneOfObject>
        {
            StronglyTypedOneOfObject.CreateSimpleObjectWithType(simpleObject),
        };

        var res = await sdk.Unions.ArrayOfDiscriminatedUnionsAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.NotNull(res.Res.Json);
        Assert.Single(res.Res.Json);
        Assert.Equal(StronglyTypedOneOfObjectType.SimpleObjectWithType, res.Res.Json[0].Type.ToString());
        Helpers.AssertSimpleObjectWithType(res.Res.Json[0].SimpleObjectWithType);
    }

    [Fact]
    public async Task ArrayOfDiscriminatedUnionsMap()
    {
        CommonHelpers.RecordTest("unions-array-of-discriminated-unions-map");
        var sdk = new Openapi.SDK(serverUrl: Helpers.HttpBinUrl);

        var simpleObject = Helpers.CreateSimpleObjectWithType();

        var req = new ArrayOfDiscriminatedUnionsMap
        {
            ArrayMap = new Dictionary<string, List<StronglyTypedOneOfObject>>
            {
                ["item"] = new List<StronglyTypedOneOfObject>
                {
                    StronglyTypedOneOfObject.CreateSimpleObjectWithType(simpleObject),
                },
            },
        };

        var res = await sdk.Unions.ArrayOfDiscriminatedUnionsMapAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.NotNull(res.Res.Json.ArrayMap);
        Assert.True(res.Res.Json.ArrayMap.ContainsKey("item"));
        Assert.Single(res.Res.Json.ArrayMap["item"]);
        Assert.Equal(
            StronglyTypedOneOfObjectType.SimpleObjectWithType,
            res.Res.Json.ArrayMap["item"][0].Type.ToString()
        );
        Helpers.AssertSimpleObjectWithType(res.Res.Json.ArrayMap["item"][0].SimpleObjectWithType);
    }

    [Fact]
    public async Task NestedArrayOfDiscriminatedUnions()
    {
        CommonHelpers.RecordTest("unions-nested-array-of-discriminated-unions");
        var sdk = new Openapi.SDK(serverUrl: Helpers.HttpBinUrl);

        var simpleObject = Helpers.CreateSimpleObjectWithType();

        var req = new NestedArrayOfDiscriminatedUnions
        {
            NestedArray = new List<List<StronglyTypedOneOfObject>>
            {
                new List<StronglyTypedOneOfObject>
                {
                    StronglyTypedOneOfObject.CreateSimpleObjectWithType(simpleObject),
                },
                new List<StronglyTypedOneOfObject>
                {
                    StronglyTypedOneOfObject.CreateSimpleObjectWithType(simpleObject),
                },
            },
        };

        var res = await sdk.Unions.NestedArrayOfDiscriminatedUnionsAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.NotNull(res.Res.Json.NestedArray);
        Assert.Equal(2, res.Res.Json.NestedArray.Count);
        Assert.Single(res.Res.Json.NestedArray[0]);
        Assert.Single(res.Res.Json.NestedArray[1]);
        Helpers.AssertSimpleObjectWithType(res.Res.Json.NestedArray[0][0].SimpleObjectWithType);
        Helpers.AssertSimpleObjectWithType(res.Res.Json.NestedArray[1][0].SimpleObjectWithType);
    }

    [Fact]
    public async Task MixedUnionTypes_Array()
    {
        CommonHelpers.RecordTest("unions-mixed-union-types");
        var sdk = new Openapi.SDK(serverUrl: Helpers.HttpBinUrl);

        var bike1 = new Bike { Colour = "white" };
        var bike2 = new Bike { Colour = "brown" };
        var req = MixedUnionTypes.CreateArrayOfBike(new List<Bike> { bike1, bike2 });

        var res = await sdk.Unions.MixedUnionTypesAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(MixedUnionTypesType.ArrayOfBike, res.Res.Json.Type.ToString());
        Assert.NotNull(res.Res.Json.ArrayOfBike);
        Assert.Equal(2, res.Res.Json.ArrayOfBike.Count);
        res.Res.Json.ArrayOfBike[0].Should().BeEquivalentTo(bike1);
        res.Res.Json.ArrayOfBike[1].Should().BeEquivalentTo(bike2);

        // Test 2: single Bike
        var req2 = MixedUnionTypes.CreateBike(bike1);
        var res2 = await sdk.Unions.MixedUnionTypesAsync(req2);
        Assert.Equal(HttpStatusCode.OK, res2.HttpMeta.Response.StatusCode);
        Assert.NotNull(res2.Res);
        Assert.Equal(MixedUnionTypesType.Bike, res2.Res.Json.Type.ToString());
        Assert.NotNull(res2.Res.Json.Bike);
        res2.Res.Json.Bike.Should().BeEquivalentTo(bike1);
    }

    [Fact]
    public async Task UnionMapOptional()
    {
        CommonHelpers.RecordTest("unions-optional-union-map");
        var sdk = new Openapi.SDK(serverUrl: Helpers.HttpBinUrl);

        var req = new UnionMapOptionalRequestBody
        {
            Input = new Dictionary<string, OneOfPrimitives>
            {
                ["str"] = OneOfPrimitives.CreateStr("test"),
                ["bool"] = OneOfPrimitives.CreateBoolean(true),
            },
        };

        var res = await sdk.Unions.UnionMapOptionalAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.NotNull(res.Res.Json.Input);
        Assert.True(res.Res.Json.Input.ContainsKey("str"));
        Assert.True(res.Res.Json.Input.ContainsKey("bool"));
        Assert.Equal(OneOfPrimitivesType.Str, res.Res.Json.Input["str"].Type.ToString());
        Assert.Equal("test", res.Res.Json.Input["str"].Str);
        Assert.Equal(OneOfPrimitivesType.Boolean, res.Res.Json.Input["bool"].Type.ToString());
        Assert.Equal(true, res.Res.Json.Input["bool"].Boolean);

        // null input
        var req2 = new UnionMapOptionalRequestBody { Input = null };
        var res2 = await sdk.Unions.UnionMapOptionalAsync(req2);
        Assert.Equal(HttpStatusCode.OK, res2.HttpMeta.Response.StatusCode);
        Assert.NotNull(res2.Res);
        Assert.Null(res2.Res.Json.Input);
    }

    [Fact]
    public async Task OneOfBooleanAndStringEnum_Boolean()
    {
        CommonHelpers.RecordTest("unions-one-of-boolean-and-string-enum-with-response-boolean");
        var sdk = new Openapi.SDK(serverUrl: Helpers.HttpBinUrl);

        var req = new OneOfBooleanAndStringEnumRequestBody
        {
            Active = Active.CreateBoolean(true),
        };

        var res = await sdk.Unions.OneOfBooleanAndStringEnumAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.True(res.Res.Json.Active.IsSet);
        Assert.False(res.Res.Json.Active.IsNull);
        var active = res.Res.Json.Active.Value;
        Assert.NotNull(active);
        Assert.Equal(ActiveType.Boolean, active.Type.ToString());
        Assert.Equal(true, active.Boolean);
    }

    [Fact]
    public void UnionEnumNestedInArrayInUnion()
    {
        CommonHelpers.RecordTest("union-enum-nested-in-array-in-union");

        // Local serialization test: array of enums inside a union renders correctly.
        var enumArray = OneOfCollectionEnum.CreateArrayOfNestedenum(
            new List<Nestedenum> { Nestedenum.Abc }
        );
        var val = new OneOfCollectionEnumRes { Json = enumArray };
        var json = JsonConvert.SerializeObject(val, Utilities.GetDefaultJsonSerializerSettings());
        Assert.Equal("{\"json\":[\"abc\"]}", json);
    }

    [Fact]
    public async Task OneOfBooleanAndStringEnum_Enum()
    {
        CommonHelpers.RecordTest("unions-one-of-boolean-and-string-enum-with-response-enum");
        var sdk = new Openapi.SDK(serverUrl: Helpers.HttpBinUrl);

        var req = new OneOfBooleanAndStringEnumRequestBody
        {
            Active = Active.CreateOneOfBooleanAndStringEnum2(OneOfBooleanAndStringEnum2.True),
        };

        var res = await sdk.Unions.OneOfBooleanAndStringEnumAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.True(res.Res.Json.Active.IsSet);
        Assert.False(res.Res.Json.Active.IsNull);
        var active = res.Res.Json.Active.Value;
        Assert.NotNull(active);
        Assert.Equal(OneOfBooleanAndStringEnumActiveType.OneOfBooleanAndStringEnumUnions2, active.Type.ToString());
        Assert.Equal(OneOfBooleanAndStringEnumUnions2.True, active.OneOfBooleanAndStringEnumUnions2);
    }

    [Fact]
    public void UnionEnumArrayDeserialization()
    {
        CommonHelpers.RecordTest("union-enum-array-deserialization");

        // Deserialize an array of enums wrapped in a union response.
        var json = "{\"json\":[\"abc\",\"def\"]}";
        var res = JsonConvert.DeserializeObject<OneOfCollectionEnumRes>(
            json,
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;
        Assert.NotNull(res.Json);
        Assert.NotNull(res.Json.ArrayOfNestedenum);
        Assert.Equal(2, res.Json.ArrayOfNestedenum.Count);
        Assert.Equal(Nestedenum.Abc, res.Json.ArrayOfNestedenum[0]);
        Assert.Equal(Nestedenum.Def, res.Json.ArrayOfNestedenum[1]);

        // Round-trip: serialize then deserialize.
        var original = new OneOfCollectionEnumRes
        {
            Json = OneOfCollectionEnum.CreateArrayOfNestedenum(
                new List<Nestedenum> { Nestedenum.Abc, Nestedenum.Def }
            ),
        };
        var serialized = JsonConvert.SerializeObject(original, Utilities.GetDefaultJsonSerializerSettings());
        var deserialized = JsonConvert.DeserializeObject<OneOfCollectionEnumRes>(
            serialized,
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;
        Assert.Equal(2, deserialized.Json.ArrayOfNestedenum.Count);
        Assert.Equal(Nestedenum.Abc, deserialized.Json.ArrayOfNestedenum[0]);
        Assert.Equal(Nestedenum.Def, deserialized.Json.ArrayOfNestedenum[1]);
    }

    [Fact]
    public void SmartUnionNullableCollectionItem()
    {
        CommonHelpers.RecordTest("smart-union-nullable-collection-item");

        var settings = Utilities.GetDefaultJsonDeserializerSettings();

        // When a union member holds a collection of a non-nullable nested union, a null
        // element/value in the payload makes that nested union fail to deserialize. Resolving
        // the outer union must treat that as "this member does not match" and move on to the
        // next member, exactly as it would for any other type mismatch — a single failing
        // member must never abort resolution for the whole union.
        //
        // Array case: SmartUnionArrayWithUnionItems.items is List<union>,
        // SmartUnionArrayWithNullableItems.items is List<string?>. Payload {"items":[null]}
        // only fits SmartUnionArrayWithNullableItems, so it must win.
        var arr = JsonConvert.DeserializeObject<SmartUnionNullableArrayItem>(
            "{\"items\":[null]}",
            settings
        )!;
        Assert.Equal(
            SmartUnionNullableArrayItemType.SmartUnionArrayWithNullableItems.ToString(),
            arr.Type.ToString()
        );
        Assert.NotNull(arr.SmartUnionArrayWithNullableItems);
        Assert.Null(arr.SmartUnionArrayWithUnionItems);
        Assert.NotNull(arr.SmartUnionArrayWithNullableItems.Items);
        Assert.Single(arr.SmartUnionArrayWithNullableItems.Items!);
        Assert.Null(arr.SmartUnionArrayWithNullableItems.Items![0]);

        // Map case: same situation with a dictionary value instead of an array element.
        // SmartUnionMapWithUnionValues.entries is Dictionary<string, union>,
        // SmartUnionMapWithNullableValues.entries is Dictionary<string, string?>. Payload
        // {"entries":{"k":null}} only fits SmartUnionMapWithNullableValues, so it must win.
        var map = JsonConvert.DeserializeObject<SmartUnionNullableMapItem>(
            "{\"entries\":{\"k\":null}}",
            settings
        )!;
        Assert.Equal(
            SmartUnionNullableMapItemType.SmartUnionMapWithNullableValues.ToString(),
            map.Type.ToString()
        );
        Assert.NotNull(map.SmartUnionMapWithNullableValues);
        Assert.Null(map.SmartUnionMapWithUnionValues);
        Assert.NotNull(map.SmartUnionMapWithNullableValues.Entries);
        Assert.True(map.SmartUnionMapWithNullableValues.Entries!.ContainsKey("k"));
        Assert.Null(map.SmartUnionMapWithNullableValues.Entries!["k"]);
    }

    [Fact]
    public void SmartUnionWrappedComplexFields()
    {
        CommonHelpers.RecordTest("smart-union-wrapped-complex-fields");

        // SmartUnionWrappedComplexMember's fields are all optional+nullable COMPLEX types
        // (object / array / map), so presence-aware serialization wraps each in
        // OptionalNullable<T>. Scoring must unwrap the boxed wrapper before recursing;
        // a still-wrapped value crashes the entire pick (TargetException on the object
        // field, InvalidCastException on the array field) instead of resolving.
        var union = JsonConvert.DeserializeObject<SmartUnionWrappedComplex>(
            "{\"meta\":{\"name\":\"x\"},\"tags\":[\"a\",\"b\"],\"extras\":{\"k\":\"v\"}}",
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;

        Assert.Equal(
            SmartUnionWrappedComplexType.SmartUnionWrappedComplexMember.ToString(),
            union.Type.ToString()
        );
        Assert.NotNull(union.SmartUnionWrappedComplexMember);
        Assert.Null(union.SmartUnionWrappedPlainNote);

        var member = union.SmartUnionWrappedComplexMember!;
        Assert.Equal("x", member.Meta.Value!.Name);
        Assert.Equal(new List<string> { "a", "b" }, member.Tags.Value);
        Assert.Equal("v", member.Extras.Value!["k"]);

        // Explicit null for every wrapped field: the wrapper converter consumes the nulls
        // (IsSet && IsNull) and the member still wins over the note-only sibling.
        var allNull = JsonConvert.DeserializeObject<SmartUnionWrappedComplex>(
            "{\"meta\":null,\"tags\":null,\"extras\":null}",
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;
        Assert.NotNull(allNull.SmartUnionWrappedComplexMember);
        Assert.True(allNull.SmartUnionWrappedComplexMember!.Meta.IsNull);
        Assert.True(allNull.SmartUnionWrappedComplexMember!.Tags.IsNull);
        Assert.True(allNull.SmartUnionWrappedComplexMember!.Extras.IsNull);

        // The sibling still wins its own payload.
        var note = JsonConvert.DeserializeObject<SmartUnionWrappedComplex>(
            "{\"note\":\"hi\"}",
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;
        Assert.NotNull(note.SmartUnionWrappedPlainNote);
        Assert.Equal("hi", note.SmartUnionWrappedPlainNote!.Note);
    }

    [Fact]
    public async Task SmartUnionOpenEnumsAndSize()
    {
        CommonHelpers.RecordTest("smart-union-open-enums-and-size");

        var sdk = new Openapi.SDK(serverUrl: Helpers.HttpBinUrl);

        // Unrecognized enum value "bat" plus a "name" field. Neither variant's enum
        // matches, so smart resolution falls back to structural matching:
        // CatWithName has { kind, name }, DogSimple only { kind } — the payload's
        // extra "name" makes CatWithName the better (more-populated) match.
        var req = SmartUnionOpenEnumsAndSizeRequestBody.CreateSmartUnionOpenEnumsAndSizeCatWithName(
            new SmartUnionOpenEnumsAndSizeCatWithName {
                Kind = SmartUnionOpenEnumsAndSizeCatWithNameKind.Of("bat"),
                Name = "asdf",
            }
        );

        var res = await sdk.Unions.SmartUnionOpenEnumsAndSizeAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.Equal(
            SmartUnionOpenEnumsAndSizeJsonType.SmartUnionOpenEnumsAndSizeCatWithName.ToString(),
            res.Res.Json.Type.ToString()
        );
        Assert.NotNull(res.Res.Json.SmartUnionOpenEnumsAndSizeCatWithName);
        // Unrecognized enum value preserved, name echoed.
        Assert.Equal("bat", res.Res.Json.SmartUnionOpenEnumsAndSizeCatWithName.Kind.Value);
        Assert.Equal("asdf", res.Res.Json.SmartUnionOpenEnumsAndSizeCatWithName.Name);
    }

    [Fact]
    public async Task SmartUnionNestedUnion()
    {
        CommonHelpers.RecordTest("smart-union-nested-union");

        // Nested unions must count only the winning inner option's unrecognized values,
        // not the accumulated count across tried options.
        //   OuterA: { data: InnerUnion(cat|dog|bird) } — each inner variant has only { kind }.
        //   OuterBWrapper: { data: { kind, name } } — both open enums.
        // The api-test-service returns { json: { data: { kind: "unknown", name: "also_unknown" } } }.
        // OuterBWrapper wins on field coverage (kind AND name present) despite carrying two unrecognized
        // enum values; OuterA's inner variants each leave "name" unmatched.
        var sdk = new Openapi.SDK();

        var res = await sdk.Unions.SmartUnionNestedUnionAsync();
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res);
        Assert.NotNull(res.Res.Json.SmartUnionNestedOuterBWrapper);

        var data = res.Res.Json.SmartUnionNestedOuterBWrapper.Data;
        Assert.Equal("unknown", data.Kind.Value);
        Assert.Equal("also_unknown", data.Name.Value);
    }

    [Fact]
    public void SmartUnionNestedUnionVsNestedUnion()
    {
        CommonHelpers.RecordTest("smart-union-nested-union-vs-nested-union");

        NestedUnion Deser(string json) =>
            JsonConvert.DeserializeObject<NestedUnion>(
                json,
                Utilities.GetDefaultJsonDeserializerSettings()
            )!;

        // Both outer members wrap a nested union: PaymentUnion -> PaymentMethod
        // (CardPayment | TransferPayment), ShipmentUnion -> ShipmentMethod
        // (AirShipment | SeaShipment).

        // Clean selection, one member at a time: the other member's required
        // wrapper key is absent, so it is rejected before its nested union runs.
        var payment = Deser("{\"selectedEvent\":{\"paymentMethod\":{\"cardNumber\":\"4111\"}}}");
        Assert.Equal("PaymentUnion", payment.SelectedEvent.Type.ToString());
        Assert.Equal("4111", payment.SelectedEvent.PaymentUnion!.PaymentMethod.CardPayment!.CardNumber);

        var shipping = Deser("{\"selectedEvent\":{\"shipmentMethod\":{\"flightNumber\":\"UA1\"}}}");
        Assert.Equal("ShipmentUnion", shipping.SelectedEvent.Type.ToString());
        Assert.Equal("UA1", shipping.SelectedEvent.ShipmentUnion!.ShipmentMethod.AirShipment!.FlightNumber);

        // Nested no-match must not abort the outer union: PaymentUnion is declared
        // first and passes its pre-check (paymentMethod present), but its nested
        // union matches neither CardPayment (needs cardNumber) nor
        // TransferPayment (needs accountNumber) and throws. The outer union
        // must skip it and select ShipmentUnion — selection is order-independent.
        var skipPayment = Deser("{\"selectedEvent\":{\"paymentMethod\":{\"unknown\":\"x\"},\"shipmentMethod\":{\"flightNumber\":\"UA1\"}}}");
        Assert.Equal("ShipmentUnion", skipPayment.SelectedEvent.Type.ToString());
        Assert.Null(skipPayment.SelectedEvent.PaymentUnion);

        // Symmetric: ShipmentUnion's nested union fails while PaymentUnion resolves,
        // so the later-declared failing member is skipped.
        var skipShipping = Deser("{\"selectedEvent\":{\"paymentMethod\":{\"cardNumber\":\"4111\"},\"shipmentMethod\":{\"unknown\":\"x\"}}}");
        Assert.Equal("PaymentUnion", skipShipping.SelectedEvent.Type.ToString());
        Assert.Null(skipShipping.SelectedEvent.ShipmentUnion);
    }

    [Fact]
    public void SmartUnionNullableUnionNullableFields()
    {
        CommonHelpers.RecordTest("smart-union-nullable-union-nullable-fields");

        // Both members accept {"value": null}: ObjectWithNullableInteger types `value` as a
        // nullable value type (int?), ObjectWithNullableObject as a nullable reference type
        // (object). Explicit null must score as a match for the value type just as
        // for the reference type; otherwise ObjectWithNullableInteger is penalized (Inexact)
        // and the scorer flips to the later-declared ObjectWithNullableObject. ObjectWithNullableInteger
        // is declared first, so it must win the tie.
        var union = JsonConvert.DeserializeObject<NullableUnionWithNullableFields>(
            "{\"value\":null}",
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;

        Assert.Equal("ObjectWithNullableInteger", union.Type.ToString());
        Assert.NotNull(union.ObjectWithNullableInteger);
        Assert.Null(union.ObjectWithNullableObject);
        Assert.True(union.ObjectWithNullableInteger!.Value.IsNull);

        // The union carries a `null` member, so a top-level null payload resolves
        // to the null variant rather than either object member.
        var nullUnion = JsonConvert.DeserializeObject<NullableUnionWithNullableFields>(
            "null",
            Utilities.GetDefaultJsonDeserializerSettings()
        );
        Assert.Null(nullUnion);
    }

    [Fact]
    public void SmartUnionScoring_DoesNotDoubleCountObjectFieldPresence()
    {
        var candidates = new UnionCandidatePool<string>();
        candidates.Add<ShallowTwoFields>("Shallow", 0, "shallow");
        candidates.Add<DeepSingleLeaf>("Deep", 1, "deep");

        var json = Newtonsoft.Json.Linq.JToken.Parse(
            "{\"a\":{\"b\":{\"c\":{\"d\":\"leaf\"}}},\"z\":\"present\"}"
        );

        var (winner, _) = candidates.PickBest(
            json,
            Utilities.GetDefaultJsonDeserializerSettings()
        );

        Assert.Equal("shallow", winner?.TypeName);
    }

    private class ShallowTwoFields
    {
        [JsonProperty("a")]
        public object A { get; set; }

        [JsonProperty("z")]
        public string Z { get; set; }
    }

    private class DeepSingleLeaf
    {
        [JsonProperty("a")]
        public DeepA A { get; set; }
    }

    private class DeepA
    {
        [JsonProperty("b")]
        public DeepB B { get; set; }
    }

    private class DeepB
    {
        [JsonProperty("c")]
        public DeepC C { get; set; }
    }

    private class DeepC
    {
        [JsonProperty("d")]
        public string D { get; set; }
    }

    [Fact]
    public void ScoreUnion_AdditionalPropertiesWins()
    {
        var pool = new UnionCandidatePool<int>();
        pool.Add<ScoreOnlyA>("A", 0, 0);
        pool.Add<ScoreBWithExtras>("B", 1, 1);
        // C# does not yet spread additionalProperties on the wire - extras are
        // nested under a literal "additionalProperties" key to be captured. Once wire format
        // is fixed switch this payload to: {"a":"","b":"","c":""}
        Assert.Equal(
            typeof(ScoreBWithExtras),
            PickWinnerType(pool, "{\"a\":\"\",\"b\":\"\",\"additionalProperties\":{\"c\":\"\"}}")
        );
    }

    [Fact]
    public void ScoreUnion_AdditionalPropertiesVsExactField()
    {
        var pool = new UnionCandidatePool<int>();
        pool.Add<ScoreIdWithExtras>("A", 0, 0);
        pool.Add<ScoreFoo>("B", 1, 1);

        // ScoreFoo wins on an exact field match; ScoreIdWithExtras only has an absent field plus
        // an empty catch-all, so an exact match must outrank additionalProperties capture.
        Assert.Equal(typeof(ScoreFoo), PickWinnerType(pool, "{\"foo\":\"\"}"));
    }

    [Fact]
    public void ScoreUnion_AllCandidatesFailPreCheck_ReturnsNoWinner()
    {
        // Every candidate declares a Required.Always field the payload omits, so all are
        // rejected at PreCheck before scoring. PickBest must yield no winner; the generated
        // union converter turns this into a DeserializationException (fallback path).
        var pool = new UnionCandidatePool<int>();
        pool.Add<ReqAlwaysFoo>("A", 0, 0);
        pool.Add<ReqAlwaysBar>("B", 1, 1);

        Assert.Null(PickWinnerType(pool, "{\"other\":\"x\"}"));
        // Explicit null for a Required.Always field also fails PreCheck.
        Assert.Null(PickWinnerType(pool, "{\"foo\":null}"));
    }

    [Fact]
    public void SmartUnion_AllCandidatesFail_OpenUnionFallsBackToUnknown()
    {
        // Same no-winner condition as ScoreUnion_AllCandidatesFailPreCheck_ReturnsNoWinner
        // but through actual generated union's ReadJson rather than a hand-rolled pool.
        // Both members of SmartUnionOpenEnumsJson (cat | dog) require `kind` so an empty object
        // eliminates both at PreCheck. SmartUnionOpenEnumsJson is a response union and the
        // primary config sets forwardCompatibleUnionsByDefault: tagged-and-untagged, so the
        // no-winner case falls back to the Unknown sentinel instead of throwing.
        var res = JsonConvert.DeserializeObject<SmartUnionOpenEnumsJson>(
            "{}",
            Utilities.GetDefaultJsonDeserializerSettings()
        )!;
        Assert.True(res.IsUnknown());
        Assert.Equal(SmartUnionOpenEnumsJsonType.Unknown.ToString(), res.Type.ToString());
        Assert.NotNull(res.UnknownRaw);
    }

    [Fact]
    public void SmartUnion_AllCandidatesFail_ClosedUnionStillThrows()
    {
        // Closed-union counterpart: the request-side cat | dog union is request-only, so the
        // response default does not open it and the no-winner case must keep throwing.
        var ex = Assert.Throws<ResponseBodyDeserializer.DeserializationException>(
            () => JsonConvert.DeserializeObject<SmartUnionOpenEnumsRequestBody>(
                "{}",
                Utilities.GetDefaultJsonDeserializerSettings()
            )
        );
        Assert.Contains("SmartUnionOpenEnumsRequestBody", ex.Message);
    }

    [Fact]
    public void SmartUnion_IntegerOpenEnum_ScoresByTokenShapeAndKnownness()
    {
        // HeroWidth is IOpenEnum<long>; scoring must rank it by token shape and
        // knownness like string open enums, not blanket-match every token.
        var settings = Utilities.GetDefaultJsonDeserializerSettings();

        var pool = new UnionCandidatePool<int>();
        pool.Add<HeroWidth>("A", 0, 0);
        pool.Add<string>("B", 1, 1);

        // "720" coerces to long and is a known value, but the string token
        // shape alone demotes HeroWidth so string sibling wins.
        var (stringWinner, stringValue) = pool.PickBest(
            Newtonsoft.Json.Linq.JToken.Parse("\"720\""), settings);
        Assert.Equal(typeof(string), stringWinner?.Type);
        Assert.Equal("720", stringValue);

        // A known integer value does win over the string sibling.
        var (knownWinner, knownValue) = pool.PickBest(
            Newtonsoft.Json.Linq.JToken.Parse("720"), settings);
        Assert.Equal(typeof(HeroWidth), knownWinner?.Type);
        Assert.True(((HeroWidth)knownValue!).IsKnown());

        // An unknown value still deserializes but ranks below an
        // exact primitive sibling.
        var numericPool = new UnionCandidatePool<int>();
        numericPool.Add<HeroWidth>("A", 0, 0);
        numericPool.Add<long>("B", 1, 1);
        var (unknownWinner, unknownValue) = numericPool.PickBest(
            Newtonsoft.Json.Linq.JToken.Parse("999"), settings);
        Assert.Equal(typeof(long), unknownWinner?.Type);
        Assert.Equal(999L, unknownValue);

        // Plain long candidate gets eliminated at Precheck (lossy coercion)
        // Open enum only gets demoted, so it wins as the sole surviving candidate.
        var (fallbackWinner, fallbackValue) = numericPool.PickBest(
            Newtonsoft.Json.Linq.JToken.Parse("\"720\""), settings);
        Assert.Equal(typeof(HeroWidth), fallbackWinner?.Type);
        Assert.Equal(HeroWidth.Of(720), fallbackValue);
    }

    [Fact]
    public async Task DiscriminatedOpenEnum()
    {
        CommonHelpers.RecordTest("unions-discriminated-open-enum");
        var s = new Openapi.SDK();
        var settings = Utilities.GetDefaultJsonDeserializerSettings();

        // The union discriminates on `status`, an open enum on both variants.
        var activeReq = JsonConvert.DeserializeObject<DiscriminatedOpenEnumUnion>(
            "{\"status\":\"active\",\"userId\":\"user-123\",\"activeAt\":\"2024-01-15T10:30:00Z\"}",
            settings
        )!;
        var activeRes = await s.Unions.DiscriminatedOpenEnumAsync(activeReq);
        Assert.Equal(HttpStatusCode.OK, activeRes.HttpMeta.Response.StatusCode);
        var activeJson = activeRes.Res!.Json;
        Assert.Equal(
            DiscriminatedOpenEnumUnionType.Active.ToString(),
            activeJson.Type.ToString()
        );
        Assert.NotNull(activeJson.ObjectWithOpenEnumStatus1);
        Assert.Equal("user-123", activeJson.ObjectWithOpenEnumStatus1!.UserId);

        var inactiveReq = JsonConvert.DeserializeObject<DiscriminatedOpenEnumUnion>(
            "{\"status\":\"inactive\",\"reason\":\"User requested deactivation\",\"inactiveSince\":\"2024-01-10T15:45:00Z\"}",
            settings
        )!;
        var inactiveRes = await s.Unions.DiscriminatedOpenEnumAsync(inactiveReq);
        Assert.Equal(HttpStatusCode.OK, inactiveRes.HttpMeta.Response.StatusCode);
        var inactiveJson = inactiveRes.Res!.Json;
        Assert.Equal(
            DiscriminatedOpenEnumUnionType.Inactive.ToString(),
            inactiveJson.Type.ToString()
        );
        Assert.NotNull(inactiveJson.ObjectWithOpenEnumStatus2);
        Assert.Equal(
            "User requested deactivation",
            inactiveJson.ObjectWithOpenEnumStatus2!.Reason
        );
    }

    [Fact]
    public async Task DiscriminatedMultipleMemberships()
    {
        CommonHelpers.RecordTest("unions-discriminated-multiple-memberships");
        var s = new Openapi.SDK();

        // Car participates in both Vehicle (vehicleType discriminator) and
        // HasWheels (wheelsType discriminator); the same instance is usable in
        // both unions and carries both const discriminator fields.
        var car = new Car();
        var vehicle = Vehicle.CreateCar(car);
        var hasWheels = HasWheels.CreateFour(car);
        Assert.NotNull(hasWheels.Car);

        var res = await s.Unions.DiscriminatedOneMultipleMembershipsAsync(vehicle);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var answer = res.Res!.Json;
        Assert.Equal(VehicleType.Car.ToString(), answer.Type.ToString());
        Assert.NotNull(answer.Car);
        Assert.Equal("car", answer.Car!.VehicleType);
        Assert.Equal("four", answer.Car!.WheelsType);
    }

    [Fact]
    public void ScoreUnion_PopulatedPrimitivePreferredOverCoerced()
    {
        // A string payload matches the string variant exactly (Matched); the int variant can
        // only take it as a type mismatch (Inexact). Exact type must outrank coercion.
        var pool = new UnionCandidatePool<int>();
        pool.Add<ScoreStrV>("A", 0, 0);
        pool.Add<ScoreIntV>("B", 1, 1);

        Assert.Equal(typeof(ScoreStrV), PickWinnerType(pool, "{\"v\":\"5\"}"));
        // An empty string is still a populated string token, so the string variant wins.
        Assert.Equal(typeof(ScoreStrV), PickWinnerType(pool, "{\"v\":\"\"}"));
    }

    [Fact]
    public void ScoreUnion_TopLevelLossyPrimitiveEliminated()
    {
        // Payload 1.5 is a Float token. Newtonsoft would coerce it into `long` (banker's
        // rounding -> 2), letting a candidate Go rejects survive with a corrupted value.
        // The double candidate must win with the exact value; the long must be eliminated.
        var pool = new UnionCandidatePool<int>();
        pool.Add<long>("A", 0, 0);
        pool.Add<double>("B", 1, 1);

        var (winner, value) = pool.PickBest(
            Newtonsoft.Json.Linq.JToken.Parse("1.5"),
            new JsonSerializerSettings { Converters = Utilities.GetDefaultJsonDeserializers() }
        );
        Assert.Equal(typeof(double), winner?.Type);
        Assert.Equal("B", winner?.PropertyName);
        Assert.Equal(1.5, value);

        // With no lossless candidate, the lossy long must be eliminated, not coerced to 2.
        var lossyPool = new UnionCandidatePool<int>();
        lossyPool.Add<long>("A", 0, 0);
        lossyPool.Add<ScoreFoo>("B", 1, 1);
        Assert.Null(PickWinnerType(lossyPool, "1.5"));
    }

    [Fact]
    public void ScoreUnion_TopLevelStringToNumericEliminated()
    {
        // "7" is a String token; Newtonsoft would parse it into long. String -> numeric is
        // lossy coercion, so the long candidate must be eliminated (no winner here).
        var pool = new UnionCandidatePool<int>();
        pool.Add<long>("A", 0, 0);
        pool.Add<ScoreFoo>("B", 1, 1);
        Assert.Null(PickWinnerType(pool, "\"7\""));
    }

    [Fact]
    public void ScoreUnion_TopLevelIntegerWideningAllowed()
    {
        // Integer token 1 widens losslessly into double (JSON does not distinguish 1 from 1.0).
        // The double candidate must still be selectable.
        var pool = new UnionCandidatePool<int>();
        pool.Add<double>("A", 0, 0);
        Assert.Equal(typeof(double), PickWinnerType(pool, "1"));
    }

    [Fact]
    public void ScoreUnion_ConstFieldDiscrimination()
    {
        // The scorer discriminates by field presence, not const value: the variant whose
        // fields are all present (b and c) outranks the two that leave a field unmatched.
        // NOTE: C# does not validate const value on read, unlike Go (json.go const check)
        // and Python (validate_const), so const-only-discriminated unions can pick a
        // different winner cross-target. Full const-on-read support is not yet implemented.
        var pool = new UnionCandidatePool<int>();
        pool.Add<ConstAB>("A", 0, 0);
        pool.Add<ConstAC>("B", 1, 1);
        pool.Add<ConstBC>("C", 2, 2);

        Assert.Equal(typeof(ConstBC), PickWinnerType(pool, "{\"b\":\"1\",\"c\":\"1\"}"));
    }

    [Fact]
    public void ScoreUnion_MoreMatchedFieldsWins()
    {
        // Both variants match on `a`; the two-field variant also matches `b`, so its higher
        // Matched count wins the primary comparison.
        var pool = new UnionCandidatePool<int>();
        pool.Add<ScoreOneField>("A", 0, 0);
        pool.Add<ScoreTwoFields>("B", 1, 1);

        Assert.Equal(typeof(ScoreTwoFields), PickWinnerType(pool, "{\"a\":\"x\",\"b\":\"y\"}"));
    }

    [Fact]
    public void ScoreUnion_PureTie_FirstDeclaredWins()
    {
        // Structurally identical variants score equally; the final tiebreak is DeclarationIndex,
        // so the earlier-declared variant must win.
        var pool = new UnionCandidatePool<int>();
        pool.Add<TieA>("A", 0, 0);
        pool.Add<TieB>("B", 1, 1);

        Assert.Equal(typeof(TieA), PickWinnerType(pool, "{\"a\":\"x\"}"));
    }

    [Fact]
    public void ScoreUnion_NestedArrayElementScoring()
    {
        // Scoring recurses into array elements: the variant whose element type matches more
        // fields per item (x and y) beats the one matching only x.
        var pool = new UnionCandidatePool<int>();
        pool.Add<ArrShallow>("A", 0, 0);
        pool.Add<ArrDeep>("B", 1, 1);

        Assert.Equal(typeof(ArrDeep), PickWinnerType(pool, "{\"items\":[{\"x\":\"1\",\"y\":\"2\"}]}"));
    }

    [Fact]
    public void ScoreUnion_OpenEnumUnrecognized_LosesToExactString()
    {
        // Open enum (declared first) takes any string, but an unrecognized value scores
        // Matched+Inexact; the exact string member scores Matched only. On the Matched tie
        // the string wins despite its later declaration — matching Go's ranking.
        var pool = new UnionCandidatePool<int>();
        pool.Add<EnumUsedInRequestExplicitlyOpen>("A", 0, 0);
        pool.Add<string>("B", 1, 1);

        Assert.Equal(typeof(string), PickWinnerType(pool, "\"delta\""));
    }

    [Fact]
    public void ScoreUnion_OpenEnumRecognized_WinsTieOnDeclarationOrder()
    {
        // A declared member scores Matched only (no Inexact), tying the string member; the
        // earlier-declared open enum then wins on DeclarationIndex.
        var pool = new UnionCandidatePool<int>();
        pool.Add<EnumUsedInRequestExplicitlyOpen>("A", 0, 0);
        pool.Add<string>("B", 1, 1);

        Assert.Equal(typeof(EnumUsedInRequestExplicitlyOpen), PickWinnerType(pool, "\"alpha\""));
    }

    [Fact]
    public void ScoreUnion_NestedOpenEnumUnrecognized_LosesToStringMember()
    {
        // Scoring recurses into object members: the `kind` open enum takes "delta" as
        // Matched+Inexact, while the `kind` string member takes it as Matched. The string
        // member wins despite the enum member being declared first.
        var pool = new UnionCandidatePool<int>();
        pool.Add<ScoreKindEnum>("A", 0, 0);
        pool.Add<ScoreKindStr>("B", 1, 1);

        Assert.Equal(typeof(ScoreKindStr), PickWinnerType(pool, "{\"kind\":\"delta\"}"));
    }

    private static System.Type PickWinnerType(UnionCandidatePool<int> pool, string payload)
    {
        var (winner, _) = pool.PickBest(
            Newtonsoft.Json.Linq.JToken.Parse(payload),
            new JsonSerializerSettings { Converters = Utilities.GetDefaultJsonDeserializers() }
        );
        return winner?.Type;
    }

    private class ReqAlwaysFoo
    {
        [JsonProperty("foo", Required = Required.Always)]
        public string Foo { get; set; } = "";
    }

    private class ReqAlwaysBar
    {
        [JsonProperty("bar", Required = Required.Always)]
        public string Bar { get; set; } = "";
    }

    private class ScoreStrV
    {
        [JsonProperty("v")]
        public string V { get; set; }
    }

    // EnumUsedInRequestExplicitlyOpen is a real generated open enum (components.yaml):
    // declared members alpha/beta/gamma, x-speakeasy-unknown-values: allow. It implements
    // IOpenEnum via enum-open.cs.stmpl, so these tests exercise the generated type.
    private class ScoreKindEnum
    {
        [JsonProperty("kind")]
        public EnumUsedInRequestExplicitlyOpen Kind { get; set; }
    }

    private class ScoreKindStr
    {
        [JsonProperty("kind")]
        public string Kind { get; set; }
    }

    private class ScoreIntV
    {
        [JsonProperty("v")]
        public int V { get; set; }
    }

    private class ConstAB
    {
        [JsonProperty("a")]
        public string A { get; set; }

        [JsonProperty("b")]
        public string B { get; set; }
    }

    private class ConstAC
    {
        [JsonProperty("a")]
        public string A { get; set; }

        [JsonProperty("c")]
        public string C { get; set; }
    }

    private class ConstBC
    {
        [JsonProperty("b")]
        public string B { get; set; }

        [JsonProperty("c")]
        public string C { get; set; }
    }

    private class ScoreOneField
    {
        [JsonProperty("a")]
        public string A { get; set; }
    }

    private class ScoreTwoFields
    {
        [JsonProperty("a")]
        public string A { get; set; }

        [JsonProperty("b")]
        public string B { get; set; }
    }

    private class TieA
    {
        [JsonProperty("a")]
        public string A { get; set; }
    }

    private class TieB
    {
        [JsonProperty("a")]
        public string A { get; set; }
    }

    private class ArrShallow
    {
        [JsonProperty("items")]
        public List<ArrShallowElem> Items { get; set; }
    }

    private class ArrShallowElem
    {
        [JsonProperty("x")]
        public string X { get; set; }
    }

    private class ArrDeep
    {
        [JsonProperty("items")]
        public List<ArrDeepElem> Items { get; set; }
    }

    private class ArrDeepElem
    {
        [JsonProperty("x")]
        public string X { get; set; }

        [JsonProperty("y")]
        public string Y { get; set; }
    }

    private class ScoreOnlyA
    {
        [JsonProperty("a")]
        public string A { get; set; }
    }

    private class ScoreBWithExtras
    {
        [JsonProperty("b")]
        public string B { get; set; }

        [JsonProperty("additionalProperties")]
        public Dictionary<string, string> AdditionalProperties { get; set; }
    }

    private class ScoreIdWithExtras
    {
        [JsonProperty("id")]
        public string Id { get; set; }

        [JsonProperty("additionalProperties")]
        public Dictionary<string, string> AdditionalProperties { get; set; }
    }

    private class ScoreFoo
    {
        [JsonProperty("foo")]
        public string Foo { get; set; }
    }
}
