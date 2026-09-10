import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { UnionMapRequestBody } from "../sdk/models/operations/unionmap.js";
import { NullableOneOfRefInObject } from "../sdk/models/shared/nullableoneofrefinobject.js";
import { NullableOneOfTypeInObject } from "../sdk/models/shared/nullableoneoftypeinobject.js";
import { StronglyTypedOneOfObject } from "../sdk/models/shared/stronglytypedoneofobject.js";
import { StronglyTypedOneOfObjectWithNonStandardDiscriminatorName } from "../sdk/models/shared/stronglytypedoneofobjectwithnonstandarddiscriminatorname.js";
import { TypedObject1 } from "../sdk/models/shared/typedobject1.js";
import { TypedObject2 } from "../sdk/models/shared/typedobject2.js";
import { TypedObject3 } from "../sdk/models/shared/typedobject3.js";
import { TypedObjectNullableOneOf } from "../sdk/models/shared/typedobjectnullableoneof.js";
import { TypedObjectOneOf } from "../sdk/models/shared/typedobjectoneof.js";
import { WeaklyTypedOneOfObject } from "../sdk/models/shared/weaklytypedoneofobject.js";
import { ArrayOfDiscriminatedUnionsMap } from "../sdk/models/shared/arrayofdiscriminatedunionsmap.js";
import { NestedArrayOfDiscriminatedUnions } from "../sdk/models/shared/nestedarrayofdiscriminatedunions.js";
import { ConstObject1 } from "../sdk/models/shared/constobject1.js";

import {
  HTTPBIN_URL,
  recordTest,
  API_TEST_SERVICE_URL,
} from "./common_helpers.js";
import {
  createDeepObject,
  createDeepObjectWithType,
  createSimpleObject,
  createSimpleObjectWithType,
} from "./primary_helpers.js";

test("strongly typed one of post basic", async () => {
  recordTest("unions-strongly-typed-one-of-post-basic");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const req: StronglyTypedOneOfObject = {
    type: "simpleObjectWithType",
    str: "test",
    bool: true,
    int: 1,
    int32: 1,
    intEnum: 2,
    int32Enum: 55,
    num: 1.1,
    float32: 1.1,
    enum: "one",
    any: "any",
    date: "2020-01-01",
    dateTime: "2020-01-01T00:00:00.000Z",
    boolOpt: true,
    strOpt: "testOptional",
  };

  const result = await sdk.unions.stronglyTypedOneOfPost(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("one of made up of collections", async () => {
  recordTest("unions-collections-one-of-post");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const req = ["one", "two"];

  const result = await sdk.unions.collectionOneOfPost(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);

  const req2 = { "1": "one", "2": "two" };

  const result2 = await sdk.unions.collectionOneOfPost(req2);

  if (result2.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result2.httpMeta.response.status).toEqual(200);
  expect(result2.res.json).toEqual(req2);
});

test("strongly typed one of post with non standard discriminator name", async () => {
  recordTest(
    "unions-strongly-typed-one-of-post-with-non-standard-discriminator-name",
  );

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const req: StronglyTypedOneOfObjectWithNonStandardDiscriminatorName = {
    objType: "simpleObjectWithNonStandardTypeName",
    str: "test",
    bool: true,
    int: 1,
    int32: 1,
    intEnum: 2,
    int32Enum: 55,
    num: 1.1,
    float32: 1.1,
    enum: "one",
    any: "any",
    date: "2020-01-01",
    dateTime: "2020-01-01T00:00:00.000Z",
    boolOpt: true,
    strOpt: "testOptional",
  };

  const result =
    await sdk.unions.stronglyTypedOneOfPostWithNonStandardDiscriminatorName(
      req,
    );

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("strongly typed one of post deep", async () => {
  recordTest("unions-strongly-typed-one-of-post-deep");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const req: StronglyTypedOneOfObject = {
    any: createSimpleObject(),
    type: "deepObjectWithType",
    num: 1.1,
    bool: true,
    int: 1,
    str: "test",
    obj: createSimpleObject(),
    map: { key: createSimpleObject() },
    arr: [createSimpleObject(), createSimpleObject()],
  };

  const result = await sdk.unions.stronglyTypedOneOfPost(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("weakly typed one of post basic", async () => {
  recordTest("unions-weakly-typed-one-of-post-basic");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const obj = createSimpleObject();
  const req: WeaklyTypedOneOfObject = obj;

  const result = await sdk.unions.weaklyTypedOneOfPost(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(obj);
});

test("weakly typed one of post deep", async () => {
  recordTest("unions-weakly-typed-one-of-post-deep");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const req = createDeepObject();

  const result = await sdk.unions.weaklyTypedOneOfPost(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("typed object one of post obj1", async () => {
  recordTest("unions-typed-object-one-of-post-obj1");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const obj1: TypedObject1 = {
    type: "obj1",
    value: "test",
  };
  const req: TypedObjectOneOf = obj1;

  const result = await sdk.unions.typedObjectOneOfPost(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("typed object one of post obj2", async () => {
  recordTest("unions-typed-object-one-of-post-obj2");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const obj2: TypedObject2 = {
    type: "obj2",
    value: "test",
  };
  const req: TypedObjectOneOf = obj2;

  const result = await sdk.unions.typedObjectOneOfPost(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("typed object one of post obj3", async () => {
  recordTest("unions-typed-object-one-of-post-obj3");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const obj3: TypedObject3 = {
    type: "obj3",
    value: "test",
  };
  const req: TypedObjectOneOf = obj3;

  const result = await sdk.unions.typedObjectOneOfPost(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("typed object one of post null", async () => {
  recordTest("unions-typed-object-one-of-post-null");

  // The no-zod variant does no input validation, so `null` is forwarded to
  // the server as-is rather than rejected client-side with a
  // SDKValidationError. The server returns the wire shape and we confirm
  // no `res` body was decoded.
  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const req: TypedObjectOneOf = null as any;
  const result = await sdk.unions.typedObjectOneOfPost(req);
  expect(result.httpMeta.response.status).toEqual(200);
});

test("typed object nullable one of post obj1", async () => {
  recordTest("unions-typed-object-nullable-one-of-post-obj1");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const obj1: TypedObject1 = {
    type: "obj1",
    value: "test",
  };
  const req: TypedObjectOneOf = obj1;

  const result = await sdk.unions.typedObjectNullableOneOfPost(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("typed object nullable one of post obj2", async () => {
  recordTest("unions-typed-object-nullable-one-of-post-obj2");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const obj2: TypedObject2 = {
    type: "obj2",
    value: "test",
  };
  const req: TypedObjectNullableOneOf = obj2;

  const result = await sdk.unions.typedObjectNullableOneOfPost(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("typed object nullable one of post null", async () => {
  recordTest("unions-typed-object-nullable-one-of-post-null");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const result = await sdk.unions.typedObjectNullableOneOfPost(null);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toBeNull();
});

test("flattened typed object obj1", async () => {
  recordTest("unions-flattened-typed-object-post-obj1");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const obj: TypedObject1 = {
    type: "obj1",
    value: "one",
  };

  const result = await sdk.unions.flattenedTypedObjectPost(obj);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(obj);
});

test("nullable typed object post obj1", async () => {
  recordTest("unions-nullable-typed-object-post-obj1");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const obj: TypedObject1 = {
    type: "obj1",
    value: "one",
  };

  const result = await sdk.unions.nullableTypedObjectPost(obj);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(obj);
});

test("nullable typed object post null", async () => {
  recordTest("unions-nullable-typed-object-post-null");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const result = await sdk.unions.nullableTypedObjectPost(null);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toBeNull();
});

test("nullable one of schema post obj1", async () => {
  recordTest("unions-nullable-oneof-schema-post-obj1");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const obj: TypedObject1 = {
    type: "obj1",
    value: "one",
  };

  const result = await sdk.unions.nullableOneOfSchemaPost(obj);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(obj);
});

test("nullable one of schema post obj2", async () => {
  recordTest("unions-nullable-oneof-schema-post-obj2");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const obj: TypedObject2 = {
    type: "obj2",
    value: "two",
  };

  const result = await sdk.unions.nullableOneOfSchemaPost(obj);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(obj);
});

test("nullable one of schema post null", async () => {
  recordTest("unions-nullable-oneof-schema-post-null");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const result = await sdk.unions.nullableOneOfSchemaPost(null);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toBeNull();
});

test("nullable one of type in object", async () => {
  recordTest("unions-nullable-oneof-type-in-object-post");

  const testCases: Array<{
    name: string;
    input: NullableOneOfTypeInObject;
  }> = [
    {
      name: "Nullable fields set to null",
      input: {
        oneOfOne: true,
        nullableOneOfOne: null,
        nullableOneOfTwo: null,
      },
    },
    {
      name: "All fields set to non-null values",
      input: {
        nullableOneOfOne: true,
        nullableOneOfTwo: 2,
        oneOfOne: true,
      },
    },
  ];

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  for (const tc of testCases) {
    const result = await sdk.unions.nullableOneOfTypeInObjectPost(tc.input);

    expect(result.httpMeta.response.status, tc.name).toEqual(200);
    if (result.res == null) {
      expect.unreachable(
        tc.name + ": Server should have returned a valid response",
      );
    }
    expect(result.res.json, tc.name).toEqual(tc.input);
  }
});

test("nullable one of ref in object", async () => {
  recordTest("unions-nullable-oneof-ref-in-object-post");

  const testCases: Array<{
    name: string;
    input: NullableOneOfRefInObject;
  }> = [
    {
      name: "Nullable fields set to null",
      input: {
        oneOfOne: {
          type: "obj1",
          value: "one",
        },
        nullableOneOfOne: null,
        nullableOneOfTwo: null,
      },
    },
    {
      name: "All fields set to non-null values",
      input: {
        oneOfOne: {
          type: "obj1",
          value: "one",
        },
        nullableOneOfOne: {
          type: "obj1",
          value: "one",
        },
        nullableOneOfTwo: {
          type: "obj2",
          value: "two",
        },
      },
    },
  ];

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  for (const tc of testCases) {
    const result = await sdk.unions.nullableOneOfRefInObjectPost(tc.input);

    expect(result.httpMeta.response.status, tc.name).toEqual(200);
    if (result.res == null) {
      expect.unreachable(
        tc.name + ": Server should have returned a valid response",
      );
    }
    expect(result.res.json, tc.name).toEqual(tc.input);
  }
});

test("primitive type one of post string", async () => {
  recordTest("unions-primitive-type-one-of-post-string");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const result = await sdk.unions.primitiveTypeOneOfPost("test");

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual("test");
});

test("primitive type one of post integer", async () => {
  recordTest("unions-primitive-type-one-of-post-integer");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const result = await sdk.unions.primitiveTypeOneOfPost(111);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(111);
});

test("primitive type one of post number", async () => {
  recordTest("unions-primitive-type-one-of-post-number");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const result = await sdk.unions.primitiveTypeOneOfPost(22.2);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(22.2);
});

test("primitive type one of post boolean", async () => {
  recordTest("unions-primitive-type-one-of-post-boolean");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const result = await sdk.unions.primitiveTypeOneOfPost(true);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(true);
});

test("mixed type one of post string", async () => {
  recordTest("unions-mixed-type-one-of-post-string");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const result = await sdk.unions.mixedTypeOneOfPost("test");

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual("test");
});

test("mixed type one of post integer", async () => {
  recordTest("unions-mixed-type-one-of-post-integer");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const result = await sdk.unions.mixedTypeOneOfPost(111);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(111);
});

test("mixed type one of post object", async () => {
  recordTest("unions-mixed-type-one-of-post-object");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const obj = createSimpleObject();
  const result = await sdk.unions.mixedTypeOneOfPost(obj);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(obj);
});

test("date null union", async () => {
  recordTest("unions-date-null");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  // In no-zod, dates are sent as ISO date strings and round-tripped as such.
  const date = "2020-01-01";
  const result = await sdk.unions.unionDateNull(date);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(date);
});

test("datetime null union", async () => {
  recordTest("unions-datetime-null");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  // In no-zod, datetimes are sent as ISO strings.
  const date = "2020-01-01T00:00:00.000Z";
  const result = await sdk.unions.unionDateTimeNull(date);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(date);
});

test("datetime bigint union", async () => {
  recordTest("unions-datetime-bigint");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  // In no-zod, datetime values are ISO strings on the wire.
  const date = "2020-01-01T00:00:00.000Z";
  let result = await sdk.unions.unionDateTimeBigInt(date);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(date);

  const value = 9007199254740991;
  result = await sdk.unions.unionDateTimeBigInt(value);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(value);
});

test("union bigint str decimal", async () => {
  recordTest("unions-bigint-str-decimal");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  // In no-zod decimals are plain strings (or numbers if format=number);
  // bigints are emitted/consumed as strings on the wire.
  const dec = "3.141592653589793";
  let result = await sdk.unions.unionBigIntStrDecimal(dec);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(dec);

  const big = "9223372036854775807";

  result = await sdk.unions.unionBigIntStrDecimal(big);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(big);
});

test("union map", async () => {
  recordTest("unions-union-map");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const req: UnionMapRequestBody = {
    input: {
      str: "test",
      bool: true,
    },
  };

  const result = await sdk.unions.unionMap(req);
  expect(result.httpMeta.response.status).toEqual(200);
  expect(result.res).not.toBeNull();
  expect(result.res?.json.input["str"]).toEqual("test");
  expect(result.res?.json.input["bool"]).toEqual(true);
});

test("array of discriminated unions", async () => {
  recordTest("unions-array-of-discriminated-unions");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const req: Array<StronglyTypedOneOfObject> = [
    createSimpleObjectWithType(),
    createDeepObjectWithType(),
  ];

  const result = await sdk.unions.arrayOfDiscriminatedUnions(req);
  expect(result.httpMeta.response.status).toEqual(200);
  expect(result.res).not.toBeNull();
  expect(result.res?.json).toEqual(req);
  expect(result.res?.json).toHaveLength(2);
  expect(result.res?.json[0]?.type).toEqual("simpleObjectWithType");
  expect(result.res?.json[1]?.type).toEqual("deepObjectWithType");
});

test("array of discriminated unions map", async () => {
  recordTest("unions-array-of-discriminated-unions-map");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const req: ArrayOfDiscriminatedUnionsMap = {
    arrayMap: {
      item: [createSimpleObjectWithType(), createDeepObjectWithType()],
    },
  };

  const result = await sdk.unions.arrayOfDiscriminatedUnionsMap(req);
  expect(result.httpMeta.response.status).toEqual(200);
  expect(result.res).not.toBeNull();
  const item = result.res?.json.arrayMap["item"];
  expect(item).toBeDefined();
  expect(item).toHaveLength(2);
  expect(item![0]?.type).toEqual("simpleObjectWithType");
  expect(item![1]?.type).toEqual("deepObjectWithType");
});

test("nested array of discriminated unions", async () => {
  recordTest("unions-nested-array-of-discriminated-unions");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const simpleObject = createSimpleObjectWithType();
  const deepObject = createDeepObjectWithType();

  const req: NestedArrayOfDiscriminatedUnions = {
    nestedArray: [[simpleObject], [simpleObject, deepObject]],
  };

  const result = await sdk.unions.nestedArrayOfDiscriminatedUnions(req);
  expect(result.httpMeta.response.status).toEqual(200);
  expect(result.res).not.toBeNull();
  const array = result.res?.json.nestedArray;
  expect(array).toBeDefined();
  expect(array).toHaveLength(2);
  expect(array![0]).toHaveLength(1);
  expect(array![0]![0]?.type).toEqual("simpleObjectWithType");
  expect(array![1]).toHaveLength(2);
  expect(array![1]![0]?.type).toEqual("simpleObjectWithType");
  expect(array![1]![1]?.type).toEqual("deepObjectWithType");
});

test("const discriminator", async () => {
  recordTest("unions-const-discriminator");

  const sdk = new SDK();
  const req: ConstObject1 = { tag: "tag1", imageURL: "http://boo" };
  const res = await sdk.unions.constDiscriminatedOneOf(req);
  expect(res.httpMeta.response.status).toEqual(200);
  expect(res.res?.json).toEqual(req);
});

test("nested union should only count winning inner option's unrecognized", async () => {
  recordTest("smart-union-nested-union");
  // This test verifies that nested unions only count the winning inner option's
  // unrecognized values, not the accumulated count from all tried options.
  //
  // Outer union structure:
  //   Option A: { data: InnerUnion }
  //     where InnerUnion is:
  //       - cat:  { kind: OpenEnum["cat"] }                (1 open enum field)
  //       - dog:  { kind: OpenEnum["dog"] }                (1 open enum field)
  //       - bird: { kind: OpenEnum["bird"] }               (1 open enum field)
  //   Option B: { data: { kind: OpenEnum, name: OpenEnum } } (2 open enum fields)
  //
  // The api-test-service returns a response with unknown enum values:
  // { json: { data: { kind: "unknown", name: "also_unknown" } } }
  //
  // Expected: Option B should win because:
  //   - Option A: Inner union tries all variants (cat/dog/bird), but none match perfectly
  //     The response has 2 fields (kind, name), but each variant only has 1 field (kind)
  //     Best match would still count the extra 'name' field as unmatched
  //   - Option B: Has both kind and name fields as open enums, perfect structural match
  //     Counts 2 unrecognized (both kind and name are unknown enum values)
  //   - Option B wins because it has better field coverage despite more unrecognized enum values

  const sdk = new SDK({
    serverURL: API_TEST_SERVICE_URL,
  });

  const result = await sdk.unions.smartUnionNestedUnion();

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  // The response contains unknown enum values that were correctly parsed.
  // Option B should have won, meaning the data field contains:
  // - kind: "unknown" (an unrecognized enum value)
  // - name: "also_unknown" (an unrecognized enum value)
  const data = result.res.json.data as { kind: string; name: string };
  expect(data.kind).toEqual("unknown");
  expect(data.name).toEqual("also_unknown");
});

const circularRecursivePayload = ["hello", { nested: ["world"] }];

test("circularReferenceRecursiveOneOf", async () => {
  recordTest("unions-circular-reference-recursive-one-of");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.unions.circularReferenceRecursiveOneOf({
    value: circularRecursivePayload,
  });

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.object).toBeDefined();
  expect(res.object!.json.value).toEqual(circularRecursivePayload);
});
