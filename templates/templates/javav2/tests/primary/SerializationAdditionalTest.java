package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.openapi.Helpers.checkRoundTrip;

import java.math.BigDecimal;
import java.math.BigInteger;
import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.Optional;

import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.operations.HeaderParamsObjectHeaders;
import org.openapis.openapi.models.shared.Color.ColorEnum;
import org.openapis.openapi.models.shared.ConstEnumInt;
import org.openapis.openapi.models.shared.ConstEnumStr;
import org.openapis.openapi.models.shared.DefaultEnumInt;
import org.openapis.openapi.models.shared.DefaultEnumStr;
import org.openapis.openapi.models.shared.DefaultsAndConsts;
import org.openapis.openapi.models.shared.Obj1;
import org.openapis.openapi.utils.JSON;
import org.openapis.openapi.utils.Utils;
import org.openapitools.jackson.nullable.JsonNullable;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonMappingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.exc.ValueInstantiationException;

/**
 * Tests serialization and deserialization locally only (no interaction with
 * external http servers)
 */
public class SerializationAdditionalTest {
    
    private static final ObjectMapper m = JSON.getMapper();

    @Test
    public void testSimpleObjectRoundTrip() {
        checkRoundTrip(Helpers.createSimpleObject());
    }
    
    @Test
    public void testDeepObjectRoundTrip() {
        checkRoundTrip(Helpers.createDeepObject());
    }
    
    @Test
    public void testDateSerializesInDesiredFormat() throws JsonProcessingException {
        String json = Helpers.json(Helpers.createSimpleObject());
        assertEquals("2020-01-01", m.readTree(json).get("date").asText());
    }
    
    @Test
    public void testDefaultOptionalSerialization() throws JsonProcessingException {
        assertEquals("null", JSON.getMapper().writeValueAsString(Optional.empty()));
    }
    
    @Test
    public void testDefaultsUsedWhenFieldsNotSet() {
        DefaultsAndConsts a = DefaultsAndConsts.builder() //
                .normalField("normal") //
                .build();
        assertEquals(9007199254740991L, a.defaultBigInt().get().longValue());
        assertEquals("9223372036854775807", a.defaultBigIntStr().get().toString());
        assertTrue(a.defaultBool().get());
        assertEquals("2020-01-01", a.defaultDate().get().toString());
        assertEquals("2020-01-01T00:00Z", a.defaultDateTime().get().toString());
        assertEquals(3.141592653589793, a.defaultDecimal().get().doubleValue(), 0.00000000001);
        assertEquals(new BigDecimal("3.141592653589793238462643383279"), a.defaultDecimalStr().get());
        assertEquals(DefaultEnumInt.TWO, a.defaultEnumInt().get());
        assertEquals(DefaultEnumStr.TWO, a.defaultEnumStr().get());
        assertEquals(123, a.defaultInt().get());
        assertEquals(123.456, a.defaultNum().get(), 0.0000001);
        assertEquals("default", a.defaultStr().get());
        assertEquals(JsonNullable.of(null), a.defaultStrNullable());

        // assert with mutator methods present for defaults
        a.withDefaultBigInt(BigInteger.valueOf(1234L));
        assertEquals(1234L, a.defaultBigInt().get().longValue());

        // assert with mutator methods not present for consts

        try {
            DefaultsAndConsts.class.getMethod("withConstBigIntStr", String.class);
            Assertions.fail();
        } catch (NoSuchMethodException e) {
            // ok
        }

        // assert constant values
        assertEquals(9007199254740991L, a.constBigInt().longValue());
        assertEquals("9223372036854775807", a.constBigIntStr().toString());
        assertTrue(a.constBool());
        assertEquals(LocalDate.of(2020, 1, 1), a.constDate());
        assertEquals(OffsetDateTime.of(2020, 1, 1, 0, 0, 0, 0, ZoneOffset.UTC), a.constDateTime());
        assertEquals(3.141592653589793, a.constDecimal().doubleValue(), 0.000000000001);
        assertEquals(new BigDecimal("3.141592653589793238462643383279"), a.constDecimalStr());
        assertEquals(ConstEnumInt.TWO, a.constEnumInt());
        assertEquals(ConstEnumStr.TWO, a.constEnumStr());
        assertEquals(123L, a.constInt());
        assertEquals(123.456, a.constNum(), 0.000001);
        assertEquals("const", a.constStr());
    }
    
    @Test
    public void testHeadersRoundTrip() {
        HeaderParamsObjectHeaders a = HeaderParamsObjectHeaders.builder()
                .xHeaderObj("hello")
                .xHeaderObjExplode("there")
                .build();
        checkRoundTrip(a);
    }
    
    @Test
    public void testLocalDateSerialization() throws JsonProcessingException {
        assertEquals("\"2024-12-23\"", JSON.getMapper().writeValueAsString(Optional.of(LocalDate.of(2024, 12, 23))));
    }
    
    @Test
    public void testEnumReservedWords() {
        // really more of a compilation test (to ensure that enum member RETURN does not
        // get escaped to RETURN_)
        assertEquals("return", ColorEnum.RETURN.value());
        assertEquals("class", ColorEnum.CLASS.value());
    }
    
    @Test
    public void testGEN507WrongCaseInFieldShouldThrow() throws JsonMappingException, JsonProcessingException {
        {
            String json = "{\"field1\":\"thing\"}";
            Utils.mapper().readValue(json, Obj1.class);
        }
        {
            String json = "{\"Field1\":\"thing\"}";
            assertThrows(ValueInstantiationException.class, () -> Utils.mapper().readValue(json, Obj1.class));
        }
    }

    @Test
    public void testBooleanFieldsWithIsPrefixSerialization() throws JsonProcessingException {
        // Test that boolean fields with "is" prefix serialize correctly without duplicating properties
        // Jackson's default behavior strips "is" from getter names (isActive() -> "active")
        // This test ensures @JsonProperty prevents both "isActive" and "active" from appearing
        org.openapis.openapi.models.shared.BooleanSerializationTest obj =
            org.openapis.openapi.models.shared.BooleanSerializationTest.builder()
                .isActive(true)
                .isEnabled(false)
                .isPublic(true)
                .normalBoolean(false)
                .name("test")
                .build();

        String serialized = Utils.mapper().writeValueAsString(obj);
        com.fasterxml.jackson.databind.JsonNode jsonNode = Utils.mapper().readTree(serialized);

        // Assert that exactly 5 fields are present in JSON (no duplicates like "active", "enabled", "public")
        assertEquals(5, jsonNode.size());

        // Assert that JSON uses the correct field names from the spec
        assertTrue(jsonNode.has("isActive"));
        assertEquals(true, jsonNode.get("isActive").asBoolean());

        assertTrue(jsonNode.has("isEnabled"));
        assertEquals(false, jsonNode.get("isEnabled").asBoolean());

        assertTrue(jsonNode.has("isPublic"));
        assertEquals(true, jsonNode.get("isPublic").asBoolean());

        assertTrue(jsonNode.has("normalBoolean"));
        assertEquals(false, jsonNode.get("normalBoolean").asBoolean());

        assertTrue(jsonNode.has("name"));
        assertEquals("test", jsonNode.get("name").asText());

        // Assert that getter-derived names are NOT present (Jackson should not auto-detect from getters)
        Assertions.assertFalse(jsonNode.has("active"), "Should not have 'active' field from isActive() getter");
        Assertions.assertFalse(jsonNode.has("enabled"), "Should not have 'enabled' field from isEnabled() getter");
        Assertions.assertFalse(jsonNode.has("public"), "Should not have 'public' field from isPublic() getter");

        // Verify deserialization from JSON with correct field names
        String json = "{\"isActive\":false,\"isEnabled\":true,\"isPublic\":false,\"normalBoolean\":true,\"name\":\"test2\"}";
        org.openapis.openapi.models.shared.BooleanSerializationTest deserialized =
            Utils.mapper().readValue(json, org.openapis.openapi.models.shared.BooleanSerializationTest.class);

        assertEquals(false, deserialized.isActive());
        assertEquals(true, deserialized.isEnabled());
        assertEquals(Optional.of(false), deserialized.isPublic());
        assertEquals(Optional.of(true), deserialized.normalBoolean());
        assertEquals("test2", deserialized.name());
    }

}
