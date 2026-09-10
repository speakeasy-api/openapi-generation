#nullable enable
using System;
using NUnit.Framework;
using UnityEngine.TestTools;
using Openapi;
using Openapi.Models.Shared;
using Openapi.Models.Operations;
using Openapi.Utils;
using System.Collections;
using System.Collections.Generic;
using System.Numerics;

public class UnionsShould
{
    [UnityTest]
    public IEnumerator StronglyTypedOneOfPost_Basic()
    {
        CommonHelpers.RecordTest("unions-strongly-typed-one-of-post-basic");
        yield return CommonHelpers.Await(async () =>
        {

            var s = new Openapi.SDK();

            var obj = new SimpleObjectWithType
            {
                Str = "test",
                Bool = true,
                Int = 1,
                Int32 = 1,
                IntEnum = SimpleObjectWithTypeIntEnum.Second,
                Int32Enum = SimpleObjectWithTypeInt32Enum.FiftyFive,
                Num = 1.1,
                Float32 = 1.1f,
                Enum = Openapi.Models.Shared.Enum.One,
                Any = "any",
                Date = DateOnly.FromDateTime(DateTime.Parse("2020-01-01")),
                DateTime = DateTime.Parse("2020-01-01T00:00:00.0000001Z").ToUniversalTime(),
                BoolOpt = true,
                StrOpt = "testOptional",
                IntOptNull = null,
                NumOptNull = null
            };

            var req = StronglyTypedOneOfObject.CreateSimpleObjectWithType(obj);
            using (var res = await s.Unions.StronglyTypedOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(StronglyTypedOneOfObjectType.SimpleObjectWithType.ToString(), res.Res?.Json.Type.ToString());
                CommonHelpers.AssertObjectEquivalent(res.Res?.Json.SimpleObjectWithType, obj);
            }
        });
    }

    [UnityTest]
    public IEnumerator StronglyTypedOneOfPostWithNonStandardDiscriminatorName()
    {
        CommonHelpers.RecordTest("unions-strongly-typed-one-of-post-with-non-standard-discriminator-name");
        yield return CommonHelpers.Await(async () =>
        {
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
                Date = DateOnly.FromDateTime(DateTime.Parse("2020-01-01")),
                DateTime = DateTime.Parse("2020-01-01T00:00:00.0000001Z").ToUniversalTime(),
                BoolOpt = true,
                StrOpt = "testOptional",
                IntOptNull = null,
                NumOptNull = null
            };

            var req = StronglyTypedOneOfObjectWithNonStandardDiscriminatorName.CreateSimpleObjectWithNonStandardTypeName(obj);

            using(var res = await s.Unions.StronglyTypedOneOfPostWithNonStandardDiscriminatorNameAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(StronglyTypedOneOfObjectWithNonStandardDiscriminatorNameType.SimpleObjectWithNonStandardTypeName.ToString(), res.Res?.Json.Type.ToString());
                CommonHelpers.AssertObjectEquivalent(res.Res?.Json.SimpleObjectWithNonStandardTypeName, obj);
            }
        });
    }

    [UnityTest]
    public IEnumerator StronglyTypedOneOfPost_Deep()
    {
        CommonHelpers.RecordTest("unions-strongly-typed-one-of-post-deep");
        yield return CommonHelpers.Await(async () =>
        {
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
                Map = new Dictionary<string, SimpleObject>
                {
                    { "key", Helpers.CreateSimpleObject() }
                },
                Num = 1.1,
                Obj = Helpers.CreateSimpleObject(),
                Str = "test",
                Type = StronglyTypedOneOfObjectType.DeepObjectWithType
            };

            var req = StronglyTypedOneOfObject.CreateDeepObjectWithType(obj);

            using(var res = await s.Unions.StronglyTypedOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(StronglyTypedOneOfObjectType.DeepObjectWithType, res.Res?.Json.Type);
                CommonHelpers.AssertObjectEquivalent(res.Res?.Json.DeepObjectWithType, obj);
            }
        });
    }

    [UnityTest]
    public IEnumerator WeaklyTypedOneOfPost_Basic()
    {
        CommonHelpers.RecordTest("unions-weakly-typed-one-of-post-basic");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var obj = Helpers.CreateSimpleObject();

            var req = WeaklyTypedOneOfObject.CreateSimpleObject(obj);

            using (var res = await s.Unions.WeaklyTypedOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(WeaklyTypedOneOfObjectType.SimpleObject, res.Res?.Json.Type);
                CommonHelpers.AssertObjectEquivalent(res.Res?.Json.SimpleObject, obj);
            }
        });
    }

    [UnityTest]
    public IEnumerator WeaklyTypedOneOfPost_Deep()
    {
        CommonHelpers.RecordTest("unions-weakly-typed-one-of-post-deep");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var obj = Helpers.CreateDeepObject();

            var req = WeaklyTypedOneOfObject.CreateDeepObject(obj);

            using (var res = await s.Unions.WeaklyTypedOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(WeaklyTypedOneOfObjectType.DeepObject, res.Res?.Json.Type);
                CommonHelpers.AssertObjectEquivalent(res.Res?.Json.DeepObject, obj);
            }
        });
    }

    [UnityTest]
    public IEnumerator TypedObjectOneOfPost_Obj1()
    {
        CommonHelpers.RecordTest("unions-typed-object-one-of-post-obj1");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var obj = new TypedObject1
            {
                Type = TypedObject1Type.Obj1
            };

            var req = TypedObjectOneOf.CreateTypedObject1(obj);

            using (var res = await s.Unions.TypedObjectOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(TypedObjectOneOfType.TypedObject1, res.Res?.Json.Type);
                CommonHelpers.AssertObjectEquivalent(res.Res?.Json.TypedObject1, obj);
            }
        });
    }

    [UnityTest]
    public IEnumerator TypedObjectOneOfPost_Obj2()
    {
        CommonHelpers.RecordTest("unions-typed-object-one-of-post-obj2");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var obj = new TypedObject2
            {
                Type = TypedObject2Type.Obj2
            };

            var req = TypedObjectOneOf.CreateTypedObject2(obj);

            using (var res = await s.Unions.TypedObjectOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(TypedObjectOneOfType.TypedObject2, res.Res?.Json.Type);
                CommonHelpers.AssertObjectEquivalent(res.Res?.Json.TypedObject2, obj);
            }
        });
    }

    [UnityTest]
    public IEnumerator TypedObjectOneOfPost_Obj3()
    {
        CommonHelpers.RecordTest("unions-typed-object-one-of-post-obj3");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var obj = new TypedObject3
            {
                Type = TypedObject3Type.Obj3
            };

            var req = TypedObjectOneOf.CreateTypedObject3(obj);

            using (var res = await s.Unions.TypedObjectOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(TypedObjectOneOfType.TypedObject3, res.Res?.Json.Type);
                CommonHelpers.AssertObjectEquivalent(res.Res?.Json.TypedObject3, obj);
            }
        });
    }

    [Test]
    public void TypedObjectOneOfPost_Null()
    {
        CommonHelpers.RecordTest("unions-typed-object-one-of-post-null");

        /// This is a no-op in C# because the type system prevents creating a null Union type.
        Assert.True(true);
    }

    [UnityTest]
    public IEnumerator TypedObjectNullableOneOfPost_Obj1()
    {
        CommonHelpers.RecordTest("unions-typed-object-nullable-one-of-post-obj1");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var obj = new TypedObject1
            {
                Type = TypedObject1Type.Obj1
            };

            var req = TypedObjectNullableOneOf.CreateTypedObject1(obj);

            using (var res = await s.Unions.TypedObjectNullableOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(TypedObjectOneOfType.TypedObject1.ToString(), res.Res?.Json?.Type.ToString());
                CommonHelpers.AssertObjectEquivalent(res.Res?.Json?.TypedObject1, obj);
            }
        });
    }

    [UnityTest]
    public IEnumerator TypedObjectNullableOneOfPost_Obj2()
    {
        CommonHelpers.RecordTest("unions-typed-object-nullable-one-of-post-obj2");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var obj = new TypedObject2
            {
                Type = TypedObject2Type.Obj2
            };

            var req = TypedObjectNullableOneOf.CreateTypedObject2(obj);

            using (var res = await s.Unions.TypedObjectNullableOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(TypedObjectOneOfType.TypedObject2.ToString(), res.Res?.Json?.Type.ToString());
                CommonHelpers.AssertObjectEquivalent(res.Res?.Json?.TypedObject2, obj);
            }
        });
    }

    [UnityTest]
    public IEnumerator TypedObjectNullableOneOfPost_Null()
    {
        CommonHelpers.RecordTest("unions-typed-object-nullable-one-of-post-null");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var req = TypedObjectNullableOneOf.CreateNull();

            using (var res = await s.Unions.TypedObjectNullableOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.Null(res.Res!.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator FlattenedTypedObject_Obj1()
    {
        CommonHelpers.RecordTest("unions-flattened-typed-object-post-obj1");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var obj = FlattenedTypedObject1.CreateTypedObject1(new TypedObject1
            {
                Value = "one",
                Type = TypedObject1Type.Obj1
            });

            var res = await s.Unions.FlattenedTypedObjectPostAsync(obj);
            Assert.AreEqual(200, res.StatusCode);
            CommonHelpers.AssertObjectEquivalent(res.Res!.Json, obj);
        });
    }

    [UnityTest]
    public IEnumerator NullableTypedObjectPost_Obj1()
    {
        CommonHelpers.RecordTest("unions-nullable-typed-object-post-obj1");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var obj = new TypedObject1
            {
                Value = "one",
                Type = TypedObject1Type.Obj1
            };

            using (var res = await s.Unions.NullableTypedObjectPostAsync(obj))
            {
                Assert.AreEqual(200, res.StatusCode);
                CommonHelpers.AssertObjectEquivalent(res.Res!.Json, obj);
            }
        });
    }

    [UnityTest]
    public IEnumerator NullableTypedObjectPost_Null()
    {
        CommonHelpers.RecordTest("unions-nullable-typed-object-post-null");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var res = await s.Unions.NullableTypedObjectPostAsync(null);
            Assert.AreEqual(200, res.StatusCode);
            Assert.Null(res.Res.Json);
        });
    }

    [UnityTest]
    public IEnumerator NullableOneOfSchemaPost_Obj1()
    {
        CommonHelpers.RecordTest("unions-nullable-oneof-schema-post-obj1");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var obj = new TypedObject1
            {
                Value = "one",
                Type = TypedObject1Type.Obj1
            };

            var req = NullableOneOfSchemaPostRequestBody.CreateTypedObject1(obj);

            using (var res = await s.Unions.NullableOneOfSchemaPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(NullableOneOfSchemaPostJsonType.TypedObject1.ToString(), res.Res.Json!.Type.ToString());
                CommonHelpers.AssertObjectEquivalent(res.Res!.Json.TypedObject1, obj);
            }
        });
    }

    [UnityTest]
    public IEnumerator NullableOneOfSchemaPost_Obj2()
    {
        CommonHelpers.RecordTest("unions-nullable-oneof-schema-post-obj2");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var obj = new TypedObject2
            {
                Value = "two",
                Type = TypedObject2Type.Obj2
            };

            var req = NullableOneOfSchemaPostRequestBody.CreateTypedObject2(obj);

            using (var res = await s.Unions.NullableOneOfSchemaPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(NullableOneOfSchemaPostJsonType.TypedObject2.ToString(), res.Res!.Json!.Type.ToString());
                CommonHelpers.AssertObjectEquivalent(res.Res!.Json!.TypedObject2, obj);
            }
        });
    }

    [UnityTest]
    public IEnumerator NullableOneOfSchemaPost_Null()
    {
        CommonHelpers.RecordTest("unions-nullable-oneof-schema-post-null");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            using (var res = await s.Unions.NullableOneOfSchemaPostAsync(null))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.Null(res.Res!.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator NullableOneOfTypeInObject()
    {
        CommonHelpers.RecordTest("unions-nullable-oneof-type-in-object-post");
        yield return CommonHelpers.Await(async () =>
        {
            var tests = new CommonHelpers.TestTableEntry[] {
                new CommonHelpers.TestTableEntry
                {
                    name = "Non-nullable field set only",
                    arg = new NullableOneOfTypeInObject
                    {
                        OneOfOne = true
                    },
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
                        NullableOneOfTwo =NullableOneOfTypeInObjectNullableOneOfTwo.CreateInteger(2),
                        OneOfOne = true
                    },
                    want = "{\"NullableOneOfOne\":true,\"NullableOneOfTwo\":2,\"OneOfOne\":true}"
                }
            };

            var s = new Openapi.SDK();
            foreach (var test in tests)
            {
                var req = test.arg;
                var serializedBody = RequestBodySerializer.Serialize(req, "Request", "json", false, false);

                using (var res = await s.Unions.NullableOneOfTypeInObjectPostAsync((NullableOneOfTypeInObject)req))
                {
                    Assert.AreEqual(200, res.StatusCode);
                    CommonHelpers.AssertObjectEquivalent(res.Res!.Json, req);
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator NullableOneOfRefInObject()
    {
        CommonHelpers.RecordTest("unions-nullable-oneof-ref-in-object-post");
        yield return CommonHelpers.Await(async () =>
        {

            var tests = new CommonHelpers.TestTableEntry[] {
                new CommonHelpers.TestTableEntry
                {
                    name = "Non-nullable field set only",
                    arg = new NullableOneOfRefInObject
                    {
                        OneOfOne = OneOfOne.CreateTypedObject1(new TypedObject1
                            {
                                Value = "one",
                                Type = TypedObject1Type.Obj1
                            }
                        )
                    },
                    want = "{\"NullableOneOfOne\":null,\"NullableOneOfTwo\":null,\"OneOfOne\":{\"type\":\"obj1\",\"value\":\"one\"}}"
                },
                new CommonHelpers.TestTableEntry
                {
                    name = "Nullable fields set to null",
                    arg = new NullableOneOfRefInObject
                    {
                        NullableOneOfOne = null,
                        NullableOneOfTwo = null,
                        OneOfOne = OneOfOne.CreateTypedObject1(new TypedObject1
                            {
                                Value = "one",
                                Type = TypedObject1Type.Obj1
                            }
                        )
                    },
                    want = "{\"NullableOneOfOne\":null,\"NullableOneOfTwo\":null,\"OneOfOne\":{\"type\":\"obj1\",\"value\":\"one\"}}"
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
                        NullableOneOfTwo = NullableOneOfTwo.CreateTypedObject2(new TypedObject2
                            {
                                Value = "two",
                                Type = TypedObject2Type.Obj2
                            }
                        ),
                        OneOfOne = OneOfOne.CreateTypedObject1(new TypedObject1
                            {
                                Value="",
                                Type = TypedObject1Type.Obj1
                            }
                        )
                    },
                    want = "{\"NullableOneOfOne\":{\"type\":\"obj1\",\"value\":\"one\"},\"NullableOneOfTwo\":{\"type\":\"obj2\",\"value\":\"two\"},\"OneOfOne\":{\"type\":\"obj1\",\"value\":\"\"}}",
                }
            };
            var s = new Openapi.SDK();
            foreach (var test in tests)
            {
                NullableOneOfRefInObject req = (NullableOneOfRefInObject)test.arg;
                var serializedBody = RequestBodySerializer.Serialize(req, "Request", "json", true, false);
                var json = serializedBody!.Body;
                Assert.AreEqual(test.want, json);
                using (var res = await s.Unions.NullableOneOfRefInObjectPostAsync(req))
                {
                    Assert.AreEqual(200, res.StatusCode);
                    CommonHelpers.AssertObjectEquivalent(res.Res!.Json, req);
                }
            }
        });
    }

    [UnityTest]
    public IEnumerator PrimitiveTypeOneOfPost_String()
    {
        CommonHelpers.RecordTest("unions-primitive-type-one-of-post-string");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var req = PrimitiveTypeOneOfPostRequestBody.CreateStr("test");

            using (var res = await s.Unions.PrimitiveTypeOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(PrimitiveTypeOneOfPostJsonType.Str.ToString(), res.Res!.Json.Type.ToString());
                Assert.AreEqual("test", res.Res!.Json.Str);
            }
        });
    }

    [UnityTest]
    public IEnumerator PrimitiveTypeOneOfPost_Integer()
    {
        CommonHelpers.RecordTest("unions-primitive-type-one-of-post-integer");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var req = PrimitiveTypeOneOfPostRequestBody.CreateInteger(111);

            using (var res = await s.Unions.PrimitiveTypeOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(PrimitiveTypeOneOfPostJsonType.Integer.ToString(), res.Res!.Json.Type.ToString());
                Assert.AreEqual(111, res.Res!.Json.Integer);
            }

        });
    }

    [UnityTest]
    public IEnumerator PrimitiveTypeOneOfPost_Number()
    {
        CommonHelpers.RecordTest("unions-primitive-type-one-of-post-number");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var req = PrimitiveTypeOneOfPostRequestBody.CreateNumber(22.2);

            using (var res = await s.Unions.PrimitiveTypeOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(PrimitiveTypeOneOfPostJsonType.Number.ToString(), res.Res!.Json.Type.ToString());
                Assert.AreEqual(22.2, res.Res.Json.Number);
            }

        });
    }

    [UnityTest]
    public IEnumerator PrimitiveTypeOneOfPost_Boolean()
    {
        CommonHelpers.RecordTest("unions-primitive-type-one-of-post-boolean");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var req = PrimitiveTypeOneOfPostRequestBody.CreateBoolean(true);

            var res = await s.Unions.PrimitiveTypeOneOfPostAsync(req);
            Assert.AreEqual(200, res.StatusCode);
            Assert.AreEqual(PrimitiveTypeOneOfPostJsonType.Boolean.ToString(), res.Res!.Json.Type.ToString());
            Assert.AreEqual(true, res.Res.Json.Boolean);
        });
    }

    [UnityTest]
    public IEnumerator MixedTypeOneOfPost_String()
    {
        CommonHelpers.RecordTest("unions-mixed-type-one-of-post-string");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var req = MixedTypeOneOfPostRequestBody.CreateStr("test");

            using (var res = await s.Unions.MixedTypeOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(MixedTypeOneOfPostJsonType.Str.ToString(), res.Res!.Json.Type.ToString());
                Assert.AreEqual("test", res.Res.Json.Str);
            }
        });
    }

    [UnityTest]
    public IEnumerator MixedTypeOneOfPost_Integer()
    {
        CommonHelpers.RecordTest("unions-mixed-type-one-of-post-integer");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var req = MixedTypeOneOfPostRequestBody.CreateInteger(111);

            using (var res = await s.Unions.MixedTypeOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(MixedTypeOneOfPostJsonType.Integer.ToString(), res.Res!.Json.Type.ToString());
                Assert.AreEqual(111, res.Res.Json.Integer);
            }
        });
    }

    [UnityTest]
    public IEnumerator MixedTypeOneOfPost_Object()
    {
        CommonHelpers.RecordTest("unions-mixed-type-one-of-post-object");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var obj = Helpers.CreateSimpleObject();

            var req = MixedTypeOneOfPostRequestBody.CreateSimpleObject(obj);

            using (var res = await s.Unions.MixedTypeOneOfPostAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(MixedTypeOneOfPostJsonType.SimpleObject.ToString(), res.Res!.Json.Type.ToString());
                CommonHelpers.AssertObjectEquivalent(res.Res.Json.SimpleObject, obj);

            }
        });
    }

    [UnityTest]
    public IEnumerator DateNullUnion()
    {
        CommonHelpers.RecordTest("unions-date-null");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var date = DateOnly.FromDateTime(DateTime.Parse("2020-01-01"));

            using (var res = await s.Unions.UnionDateNullAsync(date))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(date, res.Res!.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator DateTimeNullUnion()
    {
        CommonHelpers.RecordTest("unions-datetime-null");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var dateTime = System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime();

            using (var res = await s.Unions.UnionDateTimeNullAsync(dateTime))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(dateTime, res.Res!.Json);
            }
        });
    }

    [UnityTest]
    public IEnumerator DateTimeBigintUnion()
    {
        CommonHelpers.RecordTest("unions-datetime-bigint");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var datetime = UnionDateTimeBigIntRequestBody.CreateDateTime(System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime());
            using(var res = await s.Unions.UnionDateTimeBigIntAsync(datetime))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(UnionDateTimeBigIntRequestBodyType.DateTime.ToString(), res.Res!.Json.Type.ToString());
                Assert.AreEqual(System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime(), res.Res.Json.DateTime);
            }

            var bigint = UnionDateTimeBigIntRequestBody.CreateBigint(9007199254740991);
            using (var res = await s.Unions.UnionDateTimeBigIntAsync(bigint))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(UnionDateTimeBigIntRequestBodyType.Bigint.ToString(), res.Res!.Json.Type.ToString());
                Assert.AreEqual(new BigInteger(9007199254740991), res.Res.Json.Bigint);
            }
        });
    }

    [UnityTest]
    public IEnumerator UnionBigIntStrDecimal()
    {
        CommonHelpers.RecordTest("unions-bigint-str-decimal");
        yield return CommonHelpers.Await(async () =>
        {
            var s = new Openapi.SDK();

            var req = UnionBigIntStrDecimalRequestBody.CreateDecimal(3.141592653589793M);
            var json = Helpers.GetSerializedBodyJson(req);
            Assert.AreEqual("3.141592653589793", json);

            using (var res = await s.Unions.UnionBigIntStrDecimalAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(UnionBigIntStrDecimalRequestBodyType.Decimal.ToString(), res.Res!.Json.Type.ToString());
                Assert.AreEqual(3.14159265358979M, res.Res.Json.Decimal); // TODO decimal resolution loses last digit?
            }

            req = UnionBigIntStrDecimalRequestBody.CreateBigint(9223372036854775807);
            json = Helpers.GetSerializedBodyJson(req);
            Assert.AreEqual("\"9223372036854775807\"", json);

            using (var res = await s.Unions.UnionBigIntStrDecimalAsync(req))
            {
                Assert.AreEqual(200, res.StatusCode);
                Assert.AreEqual(UnionBigIntStrDecimalRequestBodyType.Bigint.ToString(), res.Res!.Json.Type.ToString());
                Assert.AreEqual(new BigInteger(9223372036854775807), res.Res.Json.Bigint);
            }
        });
    }
}
