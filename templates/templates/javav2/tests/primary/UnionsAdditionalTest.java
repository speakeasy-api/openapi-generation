package org.openapis.openapi;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonMappingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.exc.MismatchedInputException;

import com.fasterxml.jackson.databind.node.JsonNodeType;
import org.apache.hc.core5.http.HttpStatus;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.operations.CollectionOneOfPostResponse;
import org.openapis.openapi.models.operations.ConstDiscriminatedOneOfResponse;
import org.openapis.openapi.models.operations.DiscriminatedOneMultipleMembershipsResponse;
import org.openapis.openapi.models.operations.FlattenedTypedObjectPostResponse;
import org.openapis.openapi.models.operations.MixedTypeOneOfPostRequestBody;
import org.openapis.openapi.models.operations.MixedTypeOneOfPostResponse;
import org.openapis.openapi.models.operations.NullableOneOfSchemaPostRequestBody;
import org.openapis.openapi.models.operations.NullableOneOfSchemaPostResponse;
import org.openapis.openapi.models.operations.NullableTypedObjectPostResponse;
import org.openapis.openapi.models.operations.OneOfOverlappingObjectsRequestBody;
import org.openapis.openapi.models.operations.OneOfOverlappingObjectsResponse;
import org.openapis.openapi.models.operations.PrimitiveTypeOneOfPostRequestBody;
import org.openapis.openapi.models.operations.PrimitiveTypeOneOfPostResponse;
import org.openapis.openapi.models.operations.SmartUnionAllConstsRequestBody;
import org.openapis.openapi.models.operations.SmartUnionAnyFieldTypeRequestBody;
import org.openapis.openapi.models.operations.SmartUnionArrayFieldsRequestBody;
import org.openapis.openapi.models.operations.SmartUnionConstFieldDiscriminationRequestBody;
import org.openapis.openapi.models.operations.SmartUnionDeeplyNestedArrayRequestBody;
import org.openapis.openapi.models.operations.SmartUnionEmptyStringRequestBody;
import org.openapis.openapi.models.operations.SmartUnionNestedStructsRequestBody;
import org.openapis.openapi.models.operations.SmartUnionNestedUnionVsFlatStructRequestBody;
import org.openapis.openapi.models.operations.SmartUnionOpenEnumsAndSizeRequestBody;
import org.openapis.openapi.models.operations.SmartUnionOpenEnumsRequestBody;
import org.openapis.openapi.models.operations.SmartUnionOptionalPointerFieldsRequestBody;
import org.openapis.openapi.models.operations.SmartUnionOptionalPointerStructsRequestBody;
import org.openapis.openapi.models.operations.SmartUnionPrefersFewerUnmatchedFieldsRequestBody;
import org.openapis.openapi.models.operations.SmartUnionPreservesOrderOnTieRequestBody;
import org.openapis.openapi.models.operations.SmartUnionSelectsMoreMatchedFieldsRequestBody;
import org.openapis.openapi.models.operations.SmartUnionThreeWayFieldDiscriminationRequestBody;
import org.openapis.openapi.models.operations.SmartUnionUnionVsUnionRequestBody;
import org.openapis.openapi.models.operations.StronglyTypedOneOfPostResponse;
import org.openapis.openapi.models.operations.StronglyTypedOneOfPostWithNonStandardDiscriminatorNameResponse;
import org.openapis.openapi.models.operations.TypedObjectNullableOneOfPostResponse;
import org.openapis.openapi.models.operations.TypedObjectOneOfPostResponse;
import org.openapis.openapi.models.operations.UnionBigIntStrDecimalRequestBody;
import org.openapis.openapi.models.operations.UnionBigIntStrDecimalResponse;
import org.openapis.openapi.models.operations.UnionDateNullResponse;
import org.openapis.openapi.models.operations.UnionDateTimeBigIntRequestBody;
import org.openapis.openapi.models.operations.UnionDateTimeBigIntResponse;
import org.openapis.openapi.models.operations.UnionDateTimeNullResponse;
import org.openapis.openapi.models.operations.UnionMapRequestBody;
import org.openapis.openapi.models.operations.UnionMapResponse;
import org.openapis.openapi.models.operations.WeaklyTypedOneOfPostResponse;
import org.openapis.openapi.models.shared.AnyOfMultiMatch;
import org.openapis.openapi.models.shared.AnyOfMultiMatchMember1;
import org.openapis.openapi.models.shared.AnyOfMultiMatchMember2;

import org.openapis.openapi.models.shared.Car;
import org.openapis.openapi.models.shared.CollectionOneOfObject;
import org.openapis.openapi.models.shared.ConstObject1;
import org.openapis.openapi.models.shared.DeepObjectWithType;
import org.openapis.openapi.models.shared.DeepObjectWithTypeAny;
import org.openapis.openapi.models.shared.DiscriminatedOpenEnumUnion;
import org.openapis.openapi.models.shared.ObjectWithOpenEnumStatus1;
import org.openapis.openapi.models.shared.ObjectWithOpenEnumStatus2;
import org.openapis.openapi.models.shared.Enum;
import org.openapis.openapi.models.shared.FlattenedTypedObject1;
import org.openapis.openapi.models.shared.HasWheels;
import org.openapis.openapi.models.shared.NullableOneOfRefInObject;
import org.openapis.openapi.models.shared.NullableOneOfTwo;
import org.openapis.openapi.models.shared.NullableOneOfTypeInObject;
import org.openapis.openapi.models.shared.NullableOneOfTypeInObjectNullableOneOfTwo;
import org.openapis.openapi.models.shared.Obj1;
import org.openapis.openapi.models.shared.Obj2;
import org.openapis.openapi.models.shared.One;
import org.openapis.openapi.models.shared.OneOfOne;
import org.openapis.openapi.models.shared.OneOfPrimitives;
import org.openapis.openapi.models.shared.SimpleObjectWithNonStandardTypeName;
import org.openapis.openapi.models.shared.SimpleObjectWithNonStandardTypeNameInt32Enum;
import org.openapis.openapi.models.shared.SimpleObjectWithNonStandardTypeNameIntEnum;
import org.openapis.openapi.models.shared.SimpleObjectWithType;
import org.openapis.openapi.models.shared.SimpleObjectWithTypeInt32Enum;
import org.openapis.openapi.models.shared.SimpleObjectWithTypeIntEnum;
import org.openapis.openapi.models.shared.SmartUnionOpenEnumsDog;
import org.openapis.openapi.models.shared.SmartUnionOpenEnumsDogKind;
import org.openapis.openapi.models.shared.Three;
import org.openapis.openapi.models.shared.TypedObject1;
import org.openapis.openapi.models.shared.TypedObject1Type;
import org.openapis.openapi.models.shared.TypedObject2;
import org.openapis.openapi.models.shared.TypedObject2Type;
import org.openapis.openapi.models.shared.TypedObject3;
import org.openapis.openapi.models.shared.TypedObject3Type;
import org.openapis.openapi.models.shared.TypedObjectNullableOneOf;
import org.openapis.openapi.models.shared.TypedObjectOneOf;
import org.openapis.openapi.models.shared.UnionOfArrays;
import org.openapis.openapi.models.shared.UnionOfArrays2;
import org.openapis.openapi.models.shared.UnknownVehicle;
import org.openapis.openapi.models.shared.Vehicle;
import org.openapis.openapi.models.shared.WeaklyTypedOneOfObject;
import org.openapis.openapi.utils.JSON;

import java.math.BigDecimal;
import java.math.BigInteger;
import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.openapi.Helpers.checkRoundTrip;
import org.openapis.openapi.models.operations.CircularReferenceRecursiveOneOfRequestBody;
import org.openapis.openapi.models.operations.CircularReferenceRecursiveOneOfResponse;
import org.openapis.openapi.models.shared.RecursiveOneOfValue;

public class UnionsAdditionalTest {

    // @Test
    // test disable because we are overriding standard Jackson deserializers with custom strict ones
    public void demonstrateObjectMapperReadValueLaxity() throws JsonMappingException, JsonProcessingException {
        ObjectMapper m = JSON.getMapper();
        assertEquals(111L, m.readValue("111", Long.class));
        assertEquals(111L, m.readValue("\"111\"", Long.class));
        assertEquals(111.0, m.readValue("111", Double.class), 0.00001);
        assertEquals("111", m.readValue("111", String.class));
        assertTrue(m.readValue("111", Boolean.class));
        assertThrows(MismatchedInputException.class, () -> m.readValue("111.0", Boolean.class));
        assertEquals(111L, m.readValue("111.1", Long.class));
        assertEquals(111.1, m.readValue("111.1", Double.class), 0.00001);
        assertEquals("111.1", m.readValue("111.1", String.class));
        m.readValue("9007199254740991", OffsetDateTime.class);
        m.readValue("1234567", LocalDate.class);
        assertEquals(1000L, m.readValue("1.0e3", Long.class));
    }

    @Test
    void testStronglyTypedOneOfPostBasic() throws Exception {
        CommonHelpers.recordTest("unions-strongly-typed-one-of-post-basic");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        SimpleObjectWithType obj = SimpleObjectWithType.builder() //
                .str("test") //
                .bool(true) //
                .int_(1L) //
                .int32(1) //
                .intEnum(SimpleObjectWithTypeIntEnum.Second) //
                .int32Enum(SimpleObjectWithTypeInt32Enum.FIFTY_FIVE) //
                .num(1.1) //
                .float32(1.1f) //
                .enum_(Enum.ONE) //
                .any("any") //
                .date(LocalDate.of(2020, 1, 1)) //
                .dateTime(OffsetDateTime.of(2020, 1, 1, 0, 0, 0, 1, ZoneOffset.UTC)) //
                .boolOpt(true) //
                .strOpt("testOptional") //
                .type("simpleObjectWithType") //
                .build();

        StronglyTypedOneOfPostResponse res = s.unions().stronglyTypedOneOfPost() //
                .request(obj) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json() instanceof SimpleObjectWithType);
    }

    @Test
    void testCollectionOneOfPost() throws Exception {
        CommonHelpers.recordTest("unions-collections-one-of-post");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        List<Object> req = List.of("one", "two");

        CollectionOneOfPostResponse res = s.unions().collectionOneOfPost().request(CollectionOneOfObject.of(req)).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json().arrayOfObject().isPresent());
        assertEquals(req, res.res().get().json().arrayOfObject().get());

        Map<String, Object> req2 = Map.of("1", "one", "2", "two");


        CollectionOneOfPostResponse res2 = s.unions().collectionOneOfPost().request(CollectionOneOfObject.of(req2)).call();
        assertNotNull(res2);
        assertEquals(HttpStatus.SC_OK, res2.statusCode());
        assertTrue(res2.res().get().json().mapOfObject().isPresent());
        assertEquals(req2, res2.res().get().json().mapOfObject().get());
    }

    @Test
    void testStronglyTypedOneOfPostDeep() throws Exception {
        CommonHelpers.recordTest("unions-strongly-typed-one-of-post-deep");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        DeepObjectWithType obj = DeepObjectWithType.builder() //
                .any(DeepObjectWithTypeAny.of(Helpers.createSimpleObject())) //
                .arr(List.of(Helpers.createSimpleObject(), Helpers.createSimpleObject())) //
                .bool(true) //
                .int_(1L) //
                .map(Map.ofEntries(Map.entry("key", Helpers.createSimpleObject()))) //
                .num(1.1) //
                .obj(Helpers.createSimpleObject()) //
                .str("test") //
                .type("deepObjectWithType") //
                .build();

        StronglyTypedOneOfPostResponse res = s.unions().stronglyTypedOneOfPost() //
                .request(obj) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json() instanceof DeepObjectWithType);
    }

    @Test
    void testWeakly6TypedOneOfPostBasic() throws Exception {
        CommonHelpers.recordTest("unions-weakly-typed-one-of-post-basic");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        WeaklyTypedOneOfPostResponse res = s.unions().weaklyTypedOneOfPost() //
                .request(WeaklyTypedOneOfObject.of(Helpers.createSimpleObject())) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json().simpleObject().isPresent());
    }

    @Test
    void testWeaklyTypedOneOfPostDeep() throws Exception {
        CommonHelpers.recordTest("unions-weakly-typed-one-of-post-deep");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        WeaklyTypedOneOfPostResponse res = s.unions().weaklyTypedOneOfPost() //
                .request(WeaklyTypedOneOfObject.of(Helpers.createDeepObject())) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json().deepObject().isPresent());
    }

    @Test
    void testTypedObjectOneOfPostObj1() throws Exception {
        CommonHelpers.recordTest("unions-typed-object-one-of-post-obj1");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        TypedObject1 obj = TypedObject1.builder() //
                .type(TypedObject1Type.OBJ1).value("typedObject1").build();
        TypedObjectOneOfPostResponse res = s.unions().typedObjectOneOfPost() //
                .request(TypedObjectOneOf.of(obj)) //
                .call();
        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json().typedObject1().isPresent());
        assertEquals(obj, res.res().get().json().typedObject1().get());
    }

    @Test
    void testTypedObjectOneOfPostObj2() throws Exception {
        CommonHelpers.recordTest("unions-typed-object-one-of-post-obj2");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        TypedObject2 obj = TypedObject2.builder() //
                .type(TypedObject2Type.OBJ2).value("typedObject2").build();
        TypedObjectOneOfPostResponse res = s.unions().typedObjectOneOfPost() //
                .request(TypedObjectOneOf.of(obj)) //
                .call();
        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json().typedObject2().isPresent());
        assertEquals(obj, res.res().get().json().typedObject2().get());
    }

    @Test
    void testTypedObjectOneOfPostObj3() throws Exception {
        CommonHelpers.recordTest("unions-typed-object-one-of-post-obj3");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        TypedObject3 obj = TypedObject3.builder() //
                .type(TypedObject3Type.OBJ3).value("typedObject3").build();
        TypedObjectOneOfPostResponse res = s.unions().typedObjectOneOfPost() //
                .request(TypedObjectOneOf.of(obj)) //
                .call();
        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json().typedObject3().isPresent());
        assertEquals(obj, res.res().get().json().typedObject3().get());
    }

    @Test
    void testTypedObjectOneOfPostNull() throws Exception {
        CommonHelpers.recordTest("unions-typed-object-one-of-post-null");

        // cannot build request argument with null field
        assertThrows(IllegalArgumentException.class, () -> TypedObjectOneOf.of((TypedObject1) null));
    }

    @Test
    void testTestTypedObjectNullableOneOfPostNull() throws Exception {
        CommonHelpers.recordTest("unions-typed-object-nullable-one-of-post-null");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        TypedObjectNullableOneOfPostResponse res = s.unions().typedObjectNullableOneOfPost() //
                .request(null).call();
        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNull(res.res().get().json().get());
    }

    @Test
    void testTestTypedObjectNullableOneOfPostObj1() throws Exception {
        CommonHelpers.recordTest("unions-typed-object-nullable-one-of-post-obj1");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        TypedObject1 obj = TypedObject1.builder() //
                .type(TypedObject1Type.OBJ1).value("typedObject1").build();
        TypedObjectNullableOneOfPostResponse res = s.unions().typedObjectNullableOneOfPost() //
                .request(TypedObjectNullableOneOf.of(obj)).call();
        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json().get().typedObject1().isPresent());
        assertEquals(obj, res.res().get().json().get().typedObject1().get());
    }

    @Test
    void testTestTypedObjectNullableOneOfPostObj2() throws Exception {
        CommonHelpers.recordTest("unions-typed-object-nullable-one-of-post-obj2");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        TypedObject2 obj = TypedObject2.builder() //
                .type(TypedObject2Type.OBJ2).value("typedObject2").build();
        TypedObjectNullableOneOfPostResponse res = s.unions().typedObjectNullableOneOfPost() //
                .request(TypedObjectNullableOneOf.of(obj)).call();
        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json().get().typedObject2().isPresent());
        assertEquals(obj, res.res().get().json().get().typedObject2().get());
    }

    @Test
    void testFlattenedTypedObject_Obj1() throws Exception {
        CommonHelpers.recordTest("unions-flattened-typed-object-post-obj1");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        TypedObject1 obj = TypedObject1.builder() //
                .value("one") //
                .type(TypedObject1Type.OBJ1) //
                .build();

        FlattenedTypedObjectPostResponse res = s.unions().flattenedTypedObjectPost() //
                .request(FlattenedTypedObject1.of(obj)) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json().typedObject1().isPresent());
        assertEquals(obj, res.res().get().json().typedObject1().get());
    }

    @Test
    void testNullableTypedObjectPostNull() throws Exception {
        CommonHelpers.recordTest("unions-nullable-typed-object-post-null");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        NullableTypedObjectPostResponse res = s.unions().nullableTypedObjectPost() //
                .request(null) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNull(res.res().get().json().get());
    }

    @Test
    void testNullableTypedObjectPostObj1() throws Exception {
        CommonHelpers.recordTest("unions-nullable-typed-object-post-obj1");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        TypedObject1 obj = TypedObject1.builder() //
                .value("one") //
                .type(TypedObject1Type.OBJ1) //
                .build();

        NullableTypedObjectPostResponse res = s.unions().nullableTypedObjectPost() //
                .request(obj) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertEquals(obj, res.res().get().json().get());
    }

    @Test
    void testNullableOneOfSchemaPostNull() throws Exception {
        CommonHelpers.recordTest("unions-nullable-oneof-schema-post-null");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        NullableOneOfSchemaPostResponse res = s.unions().nullableOneOfSchemaPost() //
                .request(null) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNull(res.res().get().json().get());
    }

    @Test
    void testNullableOneOfSchemaPostObj1() throws Exception {
        CommonHelpers.recordTest("unions-nullable-oneof-schema-post-obj1");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        TypedObject1 obj = TypedObject1.builder() //
                .value("one") //
                .type(TypedObject1Type.OBJ1) //
                .build();

        NullableOneOfSchemaPostResponse res = s.unions().nullableOneOfSchemaPost() //
                .request(NullableOneOfSchemaPostRequestBody.of(obj)) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json().get().typedObject1().isPresent());
        assertEquals(obj, res.res().get().json().get().typedObject1().get());
    }

    @Test
    void testNullableOneOfSchemaPostObj2() throws Exception {
        CommonHelpers.recordTest("unions-nullable-oneof-schema-post-obj2");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        TypedObject2 obj = TypedObject2.builder() //
                .value("two") //
                .type(TypedObject2Type.OBJ2) //
                .build();

        NullableOneOfSchemaPostResponse res = s.unions().nullableOneOfSchemaPost() //
                .request(NullableOneOfSchemaPostRequestBody.of(obj)) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json().get().typedObject2().isPresent());
        assertEquals(obj, res.res().get().json().get().typedObject2().get());
    }

    @Test
    void testPrimitiveTypeOneOfPostString() throws Exception {
        CommonHelpers.recordTest("unions-primitive-type-one-of-post-string");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        PrimitiveTypeOneOfPostRequestBody req = PrimitiveTypeOneOfPostRequestBody.of("test");

        PrimitiveTypeOneOfPostResponse res = s.unions().primitiveTypeOneOfPost() //
                .request(req) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(req.string().isPresent());
        assertTrue(res.res().get().json().string().isPresent());
        assertEquals(req.string().get(), res.res().get().json().string().get());
    }

    @Test
    void testPrimitiveTypeOneOfPostInteger() throws Exception {
        CommonHelpers.recordTest("unions-primitive-type-one-of-post-integer");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        PrimitiveTypeOneOfPostRequestBody req = PrimitiveTypeOneOfPostRequestBody.of(111L);

        PrimitiveTypeOneOfPostResponse res = s.unions().primitiveTypeOneOfPost() //
                .request(req) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(req.asLong().isPresent());
        assertTrue(res.res().get().json().asLong().isPresent());
        assertEquals(req.asLong().get(), res.res().get().json().asLong().get());
    }

    @Test
    void testPrimitiveTypeOneOfPostNumber() throws Exception {
        CommonHelpers.recordTest("unions-primitive-type-one-of-post-number");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        PrimitiveTypeOneOfPostRequestBody req = PrimitiveTypeOneOfPostRequestBody.of(22.2);

        PrimitiveTypeOneOfPostResponse res = s.unions().primitiveTypeOneOfPost() //
                .request(req) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(req.asDouble().isPresent());
        assertTrue(res.res().get().json().asDouble().isPresent());
        assertEquals(req.asDouble().get(), res.res().get().json().asDouble().get());
    }

    @Test
    void testPrimitiveTypeOneOfPostBoolean() throws Exception {
        CommonHelpers.recordTest("unions-primitive-type-one-of-post-boolean");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        PrimitiveTypeOneOfPostRequestBody req = PrimitiveTypeOneOfPostRequestBody.of(true);

        PrimitiveTypeOneOfPostResponse res = s.unions().primitiveTypeOneOfPost() //
                .request(req) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(req.asBoolean().isPresent());
        assertTrue(res.res().get().json().asBoolean().isPresent());
        assertEquals(req.asBoolean().get(), res.res().get().json().asBoolean().get());
    }

    @Test
    void testMixedTypeOneOfPostString() throws Exception {
        CommonHelpers.recordTest("unions-mixed-type-one-of-post-string");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        MixedTypeOneOfPostRequestBody req = MixedTypeOneOfPostRequestBody.of("test");

        MixedTypeOneOfPostResponse res = s.unions().mixedTypeOneOfPost() //
                .request(req) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(req.string().isPresent());
        assertTrue(res.res().get().json().string().isPresent());
        assertEquals(req.string().get(), res.res().get().json().string().get());
    }

    @Test
    void testMixedTypeOneOfPostInteger() throws Exception {
        CommonHelpers.recordTest("unions-mixed-type-one-of-post-integer");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        MixedTypeOneOfPostRequestBody req = MixedTypeOneOfPostRequestBody.of(111L);

        MixedTypeOneOfPostResponse res = s.unions().mixedTypeOneOfPost() //
                .request(req) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(req.asLong().isPresent());
        assertTrue(res.res().get().json().asLong().isPresent());
        assertEquals(req.asLong().get(), res.res().get().json().asLong().get());
    }

    @Test
    void testMixedTypeOneOfPostObject() throws Exception {
        CommonHelpers.recordTest("unions-mixed-type-one-of-post-object");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        MixedTypeOneOfPostRequestBody req = MixedTypeOneOfPostRequestBody.of(Helpers.createSimpleObject());

        MixedTypeOneOfPostResponse res = s.unions().mixedTypeOneOfPost() //
                .request(req) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(req.simpleObject().isPresent());
        assertTrue(res.res().get().json().simpleObject().isPresent());
        assertEquals(req.simpleObject().get(), res.res().get().json().simpleObject().get());
    }

    @Test
    void testDateNullUnion() throws Exception {
        CommonHelpers.recordTest("unions-date-null");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        LocalDate date = LocalDate.of(2020, 01, 01);
        UnionDateNullResponse res = s.unions().unionDateNull() //
                .request(date) //
                .call();
        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertEquals(date, res.res().get().json().get());
    }

    @Test
    void testDateTimeNullUnion() throws Exception {
        CommonHelpers.recordTest("unions-datetime-null");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        OffsetDateTime time = OffsetDateTime.of(2020, 01, 01, 0, 0, 0, 0, ZoneOffset.UTC);
        UnionDateTimeNullResponse res = s.unions().unionDateTimeNull() //
                .request(time) //
                .call();
        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertEquals(time, res.res().get().json().get());
    }

    @Test
    void testDateTimeBigIntNullUnion() throws Exception {
        CommonHelpers.recordTest("unions-datetime-bigint");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        {
            OffsetDateTime time = OffsetDateTime.of(2020, 01, 01, 0, 0, 0, 0, ZoneOffset.UTC);
            UnionDateTimeBigIntResponse res = s.unions().unionDateTimeBigInt() //
                    .request(UnionDateTimeBigIntRequestBody.of(time)) //
                    .call();
            assertNotNull(res);
            assertEquals(HttpStatus.SC_OK, res.statusCode());
            assertTrue(res.res().get().json().offsetDateTime().isPresent());
            assertEquals(time, res.res().get().json().offsetDateTime().get());
        }
        {
            BigInteger n = BigInteger.valueOf(9007199254740991L);
            UnionDateTimeBigIntResponse res = s.unions().unionDateTimeBigInt() //
                    .request(UnionDateTimeBigIntRequestBody.of(n)) //
                    .call();
            assertNotNull(res);
            assertEquals(HttpStatus.SC_OK, res.statusCode());
            assertTrue(res.res().get().json().bigInteger().isPresent());
            assertEquals(n, res.res().get().json().bigInteger().get());
        }
    }

    @Test
    void testDateTimeBigIntStrDecimalUnion() throws Exception {
        CommonHelpers.recordTest("unions-bigint-str-decimal");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        {
            BigDecimal pi = BigDecimal.valueOf(3.141592653589793);
            UnionBigIntStrDecimalResponse res = s.unions().unionBigIntStrDecimal() //
                    .request(UnionBigIntStrDecimalRequestBody.of(pi)) //
                    .call();
            assertNotNull(res);
            assertEquals(HttpStatus.SC_OK, res.statusCode());
            assertTrue(res.res().get().json().bigDecimal().isPresent());
            assertEquals(pi, res.res().get().json().bigDecimal().get());
        }
        {
            BigInteger n = new BigInteger("9223372036854775807");
            UnionBigIntStrDecimalResponse res = s.unions().unionBigIntStrDecimal() //
                    .request(UnionBigIntStrDecimalRequestBody.of(n)) //
                    .call();
            assertNotNull(res);
            assertEquals(HttpStatus.SC_OK, res.statusCode());
            assertTrue(res.res().get().json().bigInteger().isPresent());
            assertEquals(n, res.res().get().json().bigInteger().get());
        }
    }

    @Test
    void testNullableOneOfTypeInObject() {
        CommonHelpers.recordTest("unions-nullable-oneof-type-in-object-post");
        {
            NullableOneOfTypeInObject o = NullableOneOfTypeInObject.builder() //
                    .oneOfOne(true) //
                    .build();
            checkRoundTrip(o);
            assertNull(o.nullableOneOfOne().get());
            assertNull(o.nullableOneOfTwo().get());
            // ensure required nullable fields are present with value null
            JsonNode node = Helpers.jsonTree(o);
            assertTrue(node.get("NullableOneOfOne").getNodeType() == JsonNodeType.NULL);
            assertTrue(node.get("NullableOneOfTwo").getNodeType() == JsonNodeType.NULL);
        }
        {
            NullableOneOfTypeInObject o = NullableOneOfTypeInObject.builder() //
                    .nullableOneOfOne(null) //
                    .nullableOneOfTwo(null) //
                    .oneOfOne(true) //
                    .build();
            checkRoundTrip(o);
            // ensure required nullable fields are present with value null
            JsonNode node = Helpers.jsonTree(o);
            assertTrue(node.get("NullableOneOfOne").getNodeType() == JsonNodeType.NULL);
            assertTrue(node.get("NullableOneOfTwo").getNodeType() == JsonNodeType.NULL);
        }
        {
            NullableOneOfTypeInObject o = NullableOneOfTypeInObject.builder() //
                    .nullableOneOfOne(true) //
                    .nullableOneOfTwo(NullableOneOfTypeInObjectNullableOneOfTwo.of(2L)) //
                    .oneOfOne(true) //
                    .build();
            checkRoundTrip(o);
        }
    }

    @Test
    void testNullableOneOfRefInObject() {
        CommonHelpers.recordTest("unions-nullable-oneof-ref-in-object-post");
        {
            NullableOneOfRefInObject o = NullableOneOfRefInObject.builder() //
                    .oneOfOne(OneOfOne.of( //
                            TypedObject1.builder() //
                                    .value("one") //
                                    .type(TypedObject1Type.OBJ1) //
                                    .build())) //
                    .build();
            checkRoundTrip(o);
            assertNull(o.nullableOneOfOne().get());
            assertNull(o.nullableOneOfTwo().get());
            // ensure required nullable fields are present with value null
            JsonNode node = Helpers.jsonTree(o);
            assertTrue(node.get("NullableOneOfOne").getNodeType() == JsonNodeType.NULL);
            assertTrue(node.get("NullableOneOfTwo").getNodeType() == JsonNodeType.NULL);
        }
        {
            NullableOneOfRefInObject o = NullableOneOfRefInObject.builder() //
                    .nullableOneOfOne(null) //
                    .nullableOneOfTwo(null) //
                    .oneOfOne(OneOfOne.of( //
                            TypedObject1.builder() //
                                    .value("one") //
                                    .type(TypedObject1Type.OBJ1) //
                                    .build())) //
                    .build();
            checkRoundTrip(o);
            // ensure required nullable fields are present with value null
            JsonNode node = Helpers.jsonTree(o);
            assertTrue(node.get("NullableOneOfOne").getNodeType() == JsonNodeType.NULL);
            assertTrue(node.get("NullableOneOfTwo").getNodeType() == JsonNodeType.NULL);
        }
        {
            NullableOneOfRefInObject o = NullableOneOfRefInObject.builder() //
                    .nullableOneOfOne(TypedObject1.builder().value("one") //
                            .type(TypedObject1Type.OBJ1) //
                            .build()) //
                    .nullableOneOfTwo(NullableOneOfTwo.of( //
                            TypedObject2.builder() //
                                    .value("two") //
                                    .type(TypedObject2Type.OBJ2) //
                                    .build())) //
                    .oneOfOne(OneOfOne.of( //
                            TypedObject1.builder() //
                                    .value("one") //
                                    .type(TypedObject1Type.OBJ1) //
                                    .build())) //
                    .build();
            checkRoundTrip(o);
        }
    }

    @Test
    void testStronglyTypedOneOfPostWithNonStandardDiscriminatorName() throws Exception {
        CommonHelpers.recordTest("unions-strongly-typed-one-of-post-with-non-standard-discriminator-name");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        SimpleObjectWithNonStandardTypeName obj = SimpleObjectWithNonStandardTypeName.builder() //
                .str("test") //
                .bool(true) //
                .int_(1L) //
                .int32(1) //
                .intEnum(SimpleObjectWithNonStandardTypeNameIntEnum.Second) //
                .int32Enum(SimpleObjectWithNonStandardTypeNameInt32Enum.FIFTY_FIVE) //
                .num(1.1) //
                .float32(1.1f) //
                .enum_(Enum.ONE) //
                .any("any") //
                .date(LocalDate.of(2020, 1, 1)) //
                .dateTime(OffsetDateTime.of(2020, 1, 1, 0, 0, 0, 1, ZoneOffset.UTC)) //
                .boolOpt(true) //
                .strOpt("testOptional") //
                .intOptNull(null) //
                .numOptNull(null) //
                .objType("simpleObjectWithNonStandardTypeName") //
                .build();

        StronglyTypedOneOfPostWithNonStandardDiscriminatorNameResponse res = s.unions().stronglyTypedOneOfPostWithNonStandardDiscriminatorName() //
                .request(obj) //
                .call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json() instanceof SimpleObjectWithNonStandardTypeName);
    }

    @Test
    void testDiscriminatorIsConst() throws Exception {
        CommonHelpers.recordTest("unions-const-discriminator");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        ConstObject1 request = ConstObject1.builder().imageURL("http://boo").build();
        ConstDiscriminatedOneOfResponse res = s.unions() //
                .constDiscriminatedOneOf() //
                .request(request) //
                .call();
        assertEquals(200, res.statusCode());
        assertEquals(request, res.res().get().json());
    }

    @Test
    void testObjectParticipatesInMultipleOneOfs() throws Exception {
        CommonHelpers.recordTest("unions-discriminated-multiple-memberships");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        // note that Car has only const fields
        Car request = Car.builder().build();
        assertTrue(request instanceof Vehicle);
        assertTrue(request instanceof HasWheels);
        assertEquals("car", request.vehicleType());
        assertEquals("four", request.wheelsType());
        DiscriminatedOneMultipleMembershipsResponse res = s.unions() //
                .discriminatedOneMultipleMemberships() //
                .request(request) //
                .call();
        assertEquals(200, res.statusCode());
        Vehicle answer = res.res().get().json();
        assertEquals(request, answer);
        assertEquals("four", ((HasWheels) answer).wheelsType());
    }

    @Test
    void testUnionMap() throws Exception {
        CommonHelpers.recordTest("unions-union-map");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        UnionMapResponse res = s.unions().unionMap().request(UnionMapRequestBody.builder().input(Map.of("str", OneOfPrimitives.of("test"), "bool", OneOfPrimitives.of(true))).build()).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertTrue(res.res().get().json().input().get("str").string().isPresent());
        assertEquals("test", res.res().get().json().input().get("str").string().get());
        assertTrue(res.res().get().json().input().get("bool").asBoolean().isPresent());
        assertEquals(true, res.res().get().json().input().get("bool").asBoolean().get());
    }

    @Test
    @Disabled
    // TODO This test has been added but fails because response matches both Obj1 and Obj2. 
    // The matching is actually correct because the default for additionalProperties is 
    // true in openapi (but is probably not what we are aiming for in speakeasy where I 
    // believe the default for additionalProperties is assumed to be false). Having said
    // that it still fails if additionalProperties on Obj1 is set to false.
    public void testUnionExtraJsonProperties() throws Exception {
        CommonHelpers.recordTest("unions-extra-json-properties");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        {
            OneOfOverlappingObjectsResponse res = s.unions() //
                    .oneOfOverlappingObjects() //
                    .request(OneOfOverlappingObjectsRequestBody.builder() //
                            .field1("test1") //
                            .field3(1.0) //
                            .build()) //
                    .call();
            assertNotNull(res);
            assertEquals(HttpStatus.SC_OK, res.statusCode());
            assertTrue(res.res().get().json().obj1().isPresent());
            Obj1 o = res.res().get().json().obj1().get();
            assertEquals("test1", o.field1());
        }
        {
            OneOfOverlappingObjectsResponse res = s.unions() //
                    .oneOfOverlappingObjects() //
                    .request(OneOfOverlappingObjectsRequestBody.builder() //
                            .field1("test2") //
                            .field2(true) //
                            .field3(1.0) //
                            .build()) //
                    .call();
            assertTrue(res.res().get().json().obj2().isPresent());
            Obj2 o = res.res().get().json().obj2().get();
            assertEquals("test2", o.field1());
            assertTrue(o.field2());
        }
    }

    @Test
    public void testUnionOfArraysNameOverride() {
        CommonHelpers.recordTest("unions-union-of-arrays");
        // test generated factory method names when there is an erasure conflict

        // use name override
        UnionOfArrays.ofFoos(List.of(One.builder().foo("hello").build()));

        // AST supplied name
        UnionOfArrays.ofUnionOfArrays2(List.of(UnionOfArrays2.builder().bar("hi").build()));

        // 1-based index
        UnionOfArrays.of3(List.of(Three.builder().baz("b").build()));
    }

    @Test
    public void testWeakUnionMultiMatchHeuristicIsApplied() throws JsonMappingException, JsonProcessingException {
        AnyOfMultiMatchMember2 a = AnyOfMultiMatchMember2.builder() //
                .name("bingo") //
                .description("over the moon") //
                .details("some details") //
                .build();
        String json = JSON.getMapper().writeValueAsString(a);
        // assert that json matches AnyOfMultiMatchMember1 as well
        JSON.getMapper().readValue(json, AnyOfMultiMatchMember1.class);
        // assert that deserialization to oneOf parent class results as AnyOfMultiMatchMember2
        AnyOfMultiMatch result = JSON.getMapper().readValue(json, AnyOfMultiMatch.class);
        assertTrue(result.anyOfMultiMatchMember2().isPresent());
    }

    // Open union Vehicle tests are in OpenUnionAdditionalTest.java

    @Test
    void testVehicleUnknownVariantSerializationUnwrapped() throws Exception {
        // Given an unknown vehicle variant (spaceship)
        String originalJson = "{\"vehicleType\":\"spaceship\",\"propulsion\":\"warp\",\"crew\":7}";

        // When unmarshalling to UnknownVehicle
        Vehicle vehicle = JSON.getMapper().readValue(originalJson, Vehicle.class);
        assertInstanceOf(UnknownVehicle.class, vehicle);

        // And then serializing back to JSON
        String serializedJson = JSON.getMapper().writeValueAsString(vehicle);

        // Then the serialized output should be unwrapped (just the raw payload without wrapper)
        JsonNode originalNode = JSON.getMapper().readTree(originalJson);
        JsonNode serializedNode = JSON.getMapper().readTree(serializedJson);

        // Verify the serialized JSON matches the original structure (unwrapped)
        assertEquals(originalNode, serializedNode, "Serialized UnknownVehicle should be unwrapped");

        // Verify all properties are present in serialized output
        assertTrue(serializedNode.has("vehicleType"));
        assertEquals("spaceship", serializedNode.get("vehicleType").asText());
        assertTrue(serializedNode.has("propulsion"));
        assertEquals("warp", serializedNode.get("propulsion").asText());
        assertTrue(serializedNode.has("crew"));
        assertEquals(7, serializedNode.get("crew").asInt());
    }

    @Test
    void testUnknownVariantDeserializationToJsonNode() throws Exception {
        // Test that unknown union variants fall back to JsonNode
        // This simulates API evolution where a new variant is added

        // Create JSON for an unknown variant (not TypedObject1, TypedObject2, or TypedObject3)
        String unknownJson = "{\"prop1\":\"obj4\",\"prop2\":\"unknown\",\"extraField\":123}";

        TypedObjectOneOf result = JSON.getMapper().readValue(unknownJson, TypedObjectOneOf.class);

        // Should fall back to JsonNode accessor
        assertTrue(result.asJson().isPresent(), "Unknown variant should be accessible via asJson()");
        assertFalse(result.typedObject1().isPresent(), "Should not match TypedObject1");
        assertFalse(result.typedObject2().isPresent(), "Should not match TypedObject2");
        assertFalse(result.typedObject3().isPresent(), "Should not match TypedObject3");

        // Verify JsonNode contains the data
        JsonNode node = result.asJson().get();
        assertEquals("obj4", node.get("prop1").asText());
        assertEquals("unknown", node.get("prop2").asText());
        assertEquals(123, node.get("extraField").asInt());
    }

    @Test
    void testUnknownPrimitiveVariant() throws Exception {
        // Test unknown variant with primitive union (string, boolean)
        // Simulate a new primitive type like number being added

        String unknownJson = "42";  // number, not string or boolean

        OneOfPrimitives result = JSON.getMapper().readValue(unknownJson, OneOfPrimitives.class);

        // Should fall back to JsonNode
        assertTrue(result.asJson().isPresent(), "Unknown primitive should be accessible via asJson()");
        assertFalse(result.string().isPresent(), "Should not match string");
        assertFalse(result.asBoolean().isPresent(), "Should not match boolean");

        // Verify JsonNode contains the number
        JsonNode node = result.asJson().get();
        assertTrue(node.isNumber());
        assertEquals(42, node.asInt());
    }

    @Test
    void testUnknownComplexObjectVariant() throws Exception {
        // Test unknown variant with complex object union

        String unknownJson = "{\"unknownField\":\"value\",\"nested\":{\"deep\":true}}";

        MixedTypeOneOfPostRequestBody result = JSON.getMapper().readValue(unknownJson, MixedTypeOneOfPostRequestBody.class);

        // Should fall back to JsonNode
        assertTrue(result.asJson().isPresent(), "Unknown object should be accessible via asJson()");
        assertFalse(result.string().isPresent(), "Should not match string");
        assertFalse(result.asLong().isPresent(), "Should not match long");
        assertFalse(result.simpleObject().isPresent(), "Should not match SimpleObject");

        // Verify JsonNode structure
        JsonNode node = result.asJson().get();
        assertTrue(node.isObject());
        assertEquals("value", node.get("unknownField").asText());
        assertTrue(node.get("nested").get("deep").asBoolean());
    }

    @Test
    void testUnknownArrayVariant() throws Exception {
        // Test unknown variant that doesn't match List<Object> or Map<String, Object>
        // Use a primitive string value which won't match either union member

        String unknownJson = "\"not an array or map\"";

        CollectionOneOfObject result = JSON.getMapper().readValue(unknownJson, CollectionOneOfObject.class);

        // Should fall back to JsonNode
        assertTrue(result.asJson().isPresent(), "Unknown variant should be accessible via asJson()");
        assertFalse(result.arrayOfObject().isPresent(), "Should not match List<Object>");
        assertFalse(result.mapOfObject().isPresent(), "Should not match Map<String, Object>");

        // Verify JsonNode contains the string
        JsonNode node = result.asJson().get();
        assertTrue(node.isTextual());
        assertEquals("not an array or map", node.asText());
    }

    @Test
    void testRoundTripWithUnknownVariant() throws Exception {
        // Test that unknown variants can be round-tripped through serialization

        String originalJson = "{\"futureType\":\"v2\",\"newFeature\":true,\"data\":[1,2,3]}";

        // Deserialize unknown variant
        TypedObjectOneOf union = JSON.getMapper().readValue(originalJson, TypedObjectOneOf.class);
        assertTrue(union.asJson().isPresent());

        // Serialize back
        String serialized = JSON.getMapper().writeValueAsString(union);

        // Deserialize again
        TypedObjectOneOf roundTripped = JSON.getMapper().readValue(serialized, TypedObjectOneOf.class);
        assertTrue(roundTripped.asJson().isPresent());

        // Verify data integrity
        JsonNode original = JSON.getMapper().readTree(originalJson);
        JsonNode final_ = roundTripped.asJson().get();
        assertEquals(original, final_);
    }

    @Test
    void testKnownVariantStillWorks() throws Exception {
        // Ensure that known variants still work correctly and don't use JsonNode fallback

        TypedObject1 obj1 = TypedObject1.builder().type(TypedObject1Type.OBJ1).value("test").build();

        String json = JSON.getMapper().writeValueAsString(obj1);
        TypedObjectOneOf union = JSON.getMapper().readValue(json, TypedObjectOneOf.class);

        // Should match the known type, not fall back to JsonNode
        assertTrue(union.typedObject1().isPresent());
        assertFalse(union.asJson().isPresent(), "Known variant should not use JsonNode fallback");
        assertEquals(obj1, union.typedObject1().get());
    }

    @Test
    void testUnknownVariantEquality() throws Exception {
        // Test that two unknown variants with same data are equal

        String json1 = "{\"unknownType\":\"test\",\"data\":123}";
        String json2 = "{\"unknownType\":\"test\",\"data\":123}";

        TypedObjectOneOf union1 = JSON.getMapper().readValue(json1, TypedObjectOneOf.class);
        TypedObjectOneOf union2 = JSON.getMapper().readValue(json2, TypedObjectOneOf.class);

        assertEquals(union1, union2);
        assertEquals(union1.hashCode(), union2.hashCode());
    }

    @Test
    void testUnknownVariantToString() throws Exception {
        // Test that toString works for unknown variants

        String unknownJson = "{\"type\":\"future\",\"value\":\"data\"}";
        TypedObjectOneOf union = JSON.getMapper().readValue(unknownJson, TypedObjectOneOf.class);

        String str = union.toString();
        assertNotNull(str);
        assertTrue(str.contains("TypedObjectOneOf"));
    }

    // Smart Union Tests

    @Test
    void testSmartUnionOpenEnums() throws Exception {
        CommonHelpers.recordTest("smart-union-open-enums");

        // Test basic discriminator with OpenEnum
        // Union of objects with "kind" enum field (values: "cat" or "dog")
        // Payload contains kind="dog"
        // Expected: matches the dog variant
        String json = "{\"kind\":\"dog\"}";

        SmartUnionOpenEnumsRequestBody result = JSON.getMapper().readValue(json, SmartUnionOpenEnumsRequestBody.class);

        assertNotNull(result);
        assertTrue(result.smartUnionOpenEnumsDog().isPresent());
        assertEquals(SmartUnionOpenEnumsDogKind.DOG, result.smartUnionOpenEnumsDog().get().kind());
    }

    @Test
    void testSmartUnionOpenEnumsAndSize() throws Exception {
        CommonHelpers.recordTest("smart-union-open-enums-and-size");

        // Test discriminator with additional field and unrecognized enum value
        // Union of: object with "kind" enum + "name" field vs object with only "kind" enum
        // Payload contains kind="bat" (unrecognized) and name="asdf"
        // Expected: matches the variant with name field despite unrecognized enum value
        String json = "{\"kind\":\"bat\",\"name\":\"asdf\"}";

        SmartUnionOpenEnumsAndSizeRequestBody result = JSON.getMapper().readValue(json, SmartUnionOpenEnumsAndSizeRequestBody.class);

        assertNotNull(result);
        assertTrue(result.smartUnionOpenEnumsAndSizeCatWithName().isPresent());
        // The union should deserialize successfully and match the variant with more fields
        JsonNode resultJson = JSON.getMapper().valueToTree(result);
        assertEquals("bat", resultJson.get("kind").asText());
        assertEquals("asdf", resultJson.get("name").asText());
    }

    @Test
    void testSmartUnionDeeplyNestedArray() throws Exception {
        CommonHelpers.recordTest("smart-union-deeply-nested-array");

        // Test array with more nested fields wins
        // Union of: array of objects with nested "x.a" vs array of objects with nested "x.a" and "x.b"
        // Payload contains nested object with both "a" and "b" fields
        // Expected: matches the variant with more fields (includes "b")
        String json = "[{\"x\":{\"a\":\"\",\"b\":false}}]";

        SmartUnionDeeplyNestedArrayRequestBody result = JSON.getMapper().readValue(json, SmartUnionDeeplyNestedArrayRequestBody.class);

        assertNotNull(result);
        assertTrue(result.arrayOfSmartUnionDeeplyNestedArrayObjectB().isPresent());
        // The union should deserialize successfully and match the variant with more nested fields
        JsonNode resultJson = JSON.getMapper().valueToTree(result);
        assertTrue(resultJson.isArray());
        assertEquals(1, resultJson.size());
        JsonNode firstElement = resultJson.get(0);
        assertTrue(firstElement.has("x"));
        JsonNode xField = firstElement.get("x");
        assertTrue(xField.has("a"));
        assertTrue(xField.has("b"));
        assertEquals("", xField.get("a").asText());
        assertEquals(false, xField.get("b").asBoolean());
    }

    @Test
    void testSmartUnionEmptyString() throws Exception {
        CommonHelpers.recordTest("smart-union-empty-string");

        // Test simple object field discriminator
        // Union of: object with field "a" vs object with field "b"
        // Payload contains field "b" with empty string
        // Expected: matches the variant with field "b"
        String json = "{\"b\":\"\"}";

        SmartUnionEmptyStringRequestBody result = JSON.getMapper().readValue(json, SmartUnionEmptyStringRequestBody.class);

        assertNotNull(result);
        assertTrue(result.smartUnionEmptyStringObjectB().isPresent());
        // The union should deserialize successfully and match the variant with field "b"
        JsonNode resultJson = JSON.getMapper().valueToTree(result);
        assertTrue(resultJson.has("b"));
        assertEquals("", resultJson.get("b").asText());
    }

    @Test
    void testSmartUnionNestedUnion() throws Exception {
        CommonHelpers.recordTest("smart-union-nested-union");
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

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        var res = s.unions().smartUnionNestedUnion().call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get(), "Server should have returned a valid response");

        // The response contains unknown enum values that were correctly parsed.
        // Option B should have won, meaning the data field contains:
        // - kind: "unknown" (an unrecognized enum value)
        // - name: "also_unknown" (an unrecognized enum value)
        var responseJson = res.res().get().json();
        assertNotNull(responseJson);
        assertTrue(responseJson.smartUnionNestedOuterBWrapper().isPresent());

        // Verify it's Option B that won by checking the deserialized structure
        JsonNode jsonNode = JSON.getMapper().valueToTree(responseJson);
        assertTrue(jsonNode.has("data"));
        JsonNode dataNode = jsonNode.get("data");
        assertEquals("unknown", dataNode.get("kind").asText());
        assertEquals("also_unknown", dataNode.get("name").asText());
    }

    @Test
    void testSmartUnionSelectsMoreMatchedFields() throws Exception {
        CommonHelpers.recordTest("smart-union-selects-more-matched-fields");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        String json = "{\"foo\":\"test\",\"bar\":\"value\"}";
        SmartUnionSelectsMoreMatchedFieldsRequestBody req = JSON.getMapper().readValue(json, SmartUnionSelectsMoreMatchedFieldsRequestBody.class);

        var res = s.unions().smartUnionSelectsMoreMatchedFields().request(req).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get().json());
        assertTrue(res.res().get().json().smartUnionMoreFieldsB().isPresent());

        JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
        assertEquals("test", responseJson.get("foo").asText());
        assertEquals("value", responseJson.get("bar").asText());
    }

    @Test
    void testSmartUnionNestedStructs() throws Exception {
        CommonHelpers.recordTest("smart-union-nested-structs");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        String json = "{\"nested\":{\"value\":\"test\",\"extra\":\"data\"}}";
        SmartUnionNestedStructsRequestBody req = JSON.getMapper().readValue(json, SmartUnionNestedStructsRequestBody.class);

        var res = s.unions().smartUnionNestedStructs().request(req).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get().json());
        assertTrue(res.res().get().json().smartUnionNestedStructsB().isPresent());

        JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
        assertTrue(responseJson.has("nested"));
        assertEquals("test", responseJson.get("nested").get("value").asText());
        assertEquals("data", responseJson.get("nested").get("extra").asText());
    }

    @Test
    void testSmartUnionConstFieldDiscrimination() throws Exception {
        CommonHelpers.recordTest("smart-union-const-field-discrimination");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        String json = "{\"b\":\"1\",\"c\":\"1\"}";
        SmartUnionConstFieldDiscriminationRequestBody req = JSON.getMapper().readValue(json, SmartUnionConstFieldDiscriminationRequestBody.class);

        var res = s.unions().smartUnionConstFieldDiscrimination().request(req).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get().json());
        assertTrue(res.res().get().json().smartUnionConstFieldC().isPresent());

        JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
        assertEquals("1", responseJson.get("b").asText());
        assertEquals("1", responseJson.get("c").asText());
    }

    @Test
    void testSmartUnionThreeWayFieldDiscrimination() throws Exception {
        CommonHelpers.recordTest("smart-union-three-way-field-discrimination");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        String json = "{\"b\":\"test\",\"c\":\"value\"}";
        SmartUnionThreeWayFieldDiscriminationRequestBody req = JSON.getMapper().readValue(json, SmartUnionThreeWayFieldDiscriminationRequestBody.class);

        var res = s.unions().smartUnionThreeWayFieldDiscrimination().request(req).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get().json());
        assertTrue(res.res().get().json().smartUnionThreeWayC().isPresent());

        JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
        assertEquals("test", responseJson.get("b").asText());
        assertEquals("value", responseJson.get("c").asText());
    }

    @Test
    void testSmartUnionPrefersFewerUnmatchedFields() throws Exception {
        CommonHelpers.recordTest("smart-union-prefers-fewer-unmatched-fields");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        String json = "{\"foo\":\"test\"}";
        SmartUnionPrefersFewerUnmatchedFieldsRequestBody req = JSON.getMapper().readValue(json, SmartUnionPrefersFewerUnmatchedFieldsRequestBody.class);

        var res = s.unions().smartUnionPrefersFewerUnmatchedFields().request(req).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get().json());
        assertTrue(res.res().get().json().smartUnionFewerUnmatchedB().isPresent());

        JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
        assertEquals("test", responseJson.get("foo").asText());
    }

    @Test
    void testSmartUnionPreservesOrderOnTie() throws Exception {
        CommonHelpers.recordTest("smart-union-preserves-order-on-tie");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        String json = "{\"foo\":\"test\"}";
        SmartUnionPreservesOrderOnTieRequestBody req = JSON.getMapper().readValue(json, SmartUnionPreservesOrderOnTieRequestBody.class);

        var res = s.unions().smartUnionPreservesOrderOnTie().request(req).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get().json());
        assertTrue(res.res().get().json().smartUnionTieA().isPresent());

        JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
        assertEquals("test", responseJson.get("foo").asText());
    }

    @Test
    void testSmartUnionArrayFields() throws Exception {
        CommonHelpers.recordTest("smart-union-array-fields");

        String json = "[{\"name\":\"a\",\"value\":\"1\"},{\"name\":\"b\",\"value\":\"2\"}]";
        SmartUnionArrayFieldsRequestBody result = JSON.getMapper().readValue(json, SmartUnionArrayFieldsRequestBody.class);

        assertNotNull(result);
        assertTrue(result.arrayOfSmartUnionArrayItemB().isPresent());
        JsonNode resultJson = JSON.getMapper().valueToTree(result);
        assertTrue(resultJson.isArray());
        assertEquals(2, resultJson.size());
        assertEquals("a", resultJson.get(0).get("name").asText());
        assertEquals("1", resultJson.get(0).get("value").asText());
        assertEquals("b", resultJson.get(1).get("name").asText());
        assertEquals("2", resultJson.get(1).get("value").asText());
    }

    @Test
    void testSmartUnionOptionalPointerFields() throws Exception {
        CommonHelpers.recordTest("smart-union-optional-pointer-fields");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        String json = "{\"foo\":\"test\",\"bar\":\"value\"}";
        SmartUnionOptionalPointerFieldsRequestBody req = JSON.getMapper().readValue(json, SmartUnionOptionalPointerFieldsRequestBody.class);

        var res = s.unions().smartUnionOptionalPointerFields().request(req).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get().json());
        assertTrue(res.res().get().json().smartUnionOptionalPointerB().isPresent());

        JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
        assertEquals("test", responseJson.get("foo").asText());
        assertEquals("value", responseJson.get("bar").asText());
    }

    @Test
    void testSmartUnionOptionalPointerStructs() throws Exception {
        CommonHelpers.recordTest("smart-union-optional-pointer-structs");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        String json = "{\"nested\":{\"name\":\"test\",\"value\":\"data\"}}";
        SmartUnionOptionalPointerStructsRequestBody req = JSON.getMapper().readValue(json, SmartUnionOptionalPointerStructsRequestBody.class);

        var res = s.unions().smartUnionOptionalPointerStructs().request(req).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get().json());
        assertTrue(res.res().get().json().smartUnionOptionalPointerStructsB().isPresent());

        JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
        assertTrue(responseJson.has("nested"));
        assertEquals("test", responseJson.get("nested").get("name").asText());
        assertEquals("data", responseJson.get("nested").get("value").asText());
    }

    @Test
    void testSmartUnionAllConsts() throws Exception {
        CommonHelpers.recordTest("smart-union-all-consts");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        String json = "{\"b\":\"B\",\"c\":\"C\"}";
        SmartUnionAllConstsRequestBody req = JSON.getMapper().readValue(json, SmartUnionAllConstsRequestBody.class);

        var res = s.unions().smartUnionAllConsts().request(req).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get().json());
        assertTrue(res.res().get().json().smartUnionAllConstsB().isPresent());

        JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
        assertEquals("B", responseJson.get("b").asText());
        assertEquals("C", responseJson.get("c").asText());
    }

    @Test
    void testSmartUnionAnyFieldType() throws Exception {
        CommonHelpers.recordTest("smart-union-any-field-type");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        String json = "{\"b\":\"asdf\"}";
        SmartUnionAnyFieldTypeRequestBody req = JSON.getMapper().readValue(json, SmartUnionAnyFieldTypeRequestBody.class);

        var res = s.unions().smartUnionAnyFieldType().request(req).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get().json());
        assertTrue(res.res().get().json().smartUnionAnyFieldB().isPresent());

        JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
        assertEquals("asdf", responseJson.get("b").asText());
    }

    @Test
    void testSmartUnionNestedUnionVsFlatStruct() throws Exception {
        CommonHelpers.recordTest("smart-union-nested-union-vs-flat-struct");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        String json = "{\"data\":{\"x\":\"\",\"y\":\"\"}}";
        SmartUnionNestedUnionVsFlatStructRequestBody req = JSON.getMapper().readValue(json, SmartUnionNestedUnionVsFlatStructRequestBody.class);

        var res = s.unions().smartUnionNestedUnionVsFlatStruct().request(req).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get().json());
        assertTrue(res.res().get().json().smartUnionNestedUnionVsFlatStructSmartUnionNestedVsFlatOuterB().isPresent());

        JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
        assertTrue(responseJson.has("data"));
        assertEquals("", responseJson.get("data").get("x").asText());
        assertEquals("", responseJson.get("data").get("y").asText());
    }

    @Test
    void testSmartUnionUnionVsUnion() throws Exception {
        CommonHelpers.recordTest("smart-union-union-vs-union");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        String json = "{\"a\":\"\",\"b\":\"\",\"c\":\"\"}";
        SmartUnionUnionVsUnionRequestBody req = JSON.getMapper().readValue(json, SmartUnionUnionVsUnionRequestBody.class);

        var res = s.unions().smartUnionUnionVsUnion().request(req).call();

        assertNotNull(res);
        assertEquals(HttpStatus.SC_OK, res.statusCode());
        assertNotNull(res.res().get().json());
        assertTrue(res.res().get().json().smartUnionUnionVsUnionUnions2().isPresent());

        JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
        assertEquals("", responseJson.get("a").asText());
        assertEquals("", responseJson.get("b").asText());
        assertEquals("", responseJson.get("c").asText());
    }

    @Test
    void testUnionsDiscriminatedOpenEnum() throws Exception {
        CommonHelpers.recordTest("unions-discriminated-open-enum");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        // Step 1: active status
        {
            String json = "{\"status\":\"active\",\"userId\":\"user-123\",\"activeAt\":\"2024-01-15T10:30:00Z\"}";
            DiscriminatedOpenEnumUnion req = JSON.getMapper().readValue(json, DiscriminatedOpenEnumUnion.class);

            var res = s.unions().discriminatedOpenEnum().request(req).call();

            assertNotNull(res);
            assertEquals(HttpStatus.SC_OK, res.statusCode());
            assertNotNull(res.res().get().json());
            assertInstanceOf(ObjectWithOpenEnumStatus1.class, res.res().get().json());

            JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
            assertEquals("active", responseJson.get("status").asText());
            assertEquals("user-123", responseJson.get("userId").asText());
        }

        // Step 2: inactive status
        {
            String json = "{\"status\":\"inactive\",\"reason\":\"User requested deactivation\",\"inactiveSince\":\"2024-01-10T15:45:00Z\"}";
            DiscriminatedOpenEnumUnion req = JSON.getMapper().readValue(json, DiscriminatedOpenEnumUnion.class);

            var res = s.unions().discriminatedOpenEnum().request(req).call();

            assertNotNull(res);
            assertEquals(HttpStatus.SC_OK, res.statusCode());
            assertNotNull(res.res().get().json());
            assertInstanceOf(ObjectWithOpenEnumStatus2.class, res.res().get().json());

            JsonNode responseJson = JSON.getMapper().valueToTree(res.res().get().json());
            assertEquals("inactive", responseJson.get("status").asText());
            assertEquals("User requested deactivation", responseJson.get("reason").asText());
        }
    }

    @Test
    void testCircularReferenceRecursiveOneOf() throws Exception {
        CommonHelpers.recordTest("unions-circular-reference-recursive-one-of");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        RecursiveOneOfValue payload = RecursiveOneOfValue.of(List.of(
                RecursiveOneOfValue.of("hello"),
                RecursiveOneOfValue.of(Map.of(
                        "nested", RecursiveOneOfValue.of(List.of(
                                RecursiveOneOfValue.of("world")
                        ))
                ))
        ));

        CircularReferenceRecursiveOneOfResponse res = s.unions()
                .circularReferenceRecursiveOneOf(
                        new CircularReferenceRecursiveOneOfRequestBody(payload));

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.object().isPresent());
        assertEquals(payload, res.object().get().json().value());
    }
}
