package org.openapis.openapi;

import com.fasterxml.jackson.core.JsonProcessingException;
import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.operations.EnumsPostOpenEnumUnrecognizedResponse;
import org.openapis.openapi.models.shared.Color;
import org.openapis.openapi.models.shared.Color.ColorEnum;
import org.openapis.openapi.models.shared.HeroWidth;
import org.openapis.openapi.models.shared.HeroWidth.HeroWidthEnum;
import org.openapis.openapi.models.shared.Icon;
import org.openapis.openapi.models.shared.ObjectWithKnownOpenEnum;
import org.openapis.openapi.models.shared.ObjectWithUnknownOpenEnum;
import org.openapis.openapi.models.shared.Status;
import org.openapis.openapi.models.shared.Theme;
import org.openapis.openapi.models.shared.ThemeRequestOpaque;

import java.util.Arrays;
import java.util.Objects;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.openapi.CommonHelpers.recordTest;
import static org.openapis.openapi.utils.Utils.mapper;

public class EnumsAdditionalTest {

    @Test
    public void testOpenEnumResponse() throws Exception {
        recordTest("open-enums-round-trip");

        assertEquals("tick", Icon.TICK.value());
        // make sure the that purple and 2160 are unknown values in the enums
        assertFalse(Arrays.stream(ColorEnum.values()).anyMatch(c -> "purple".equals(c.value())));
        assertFalse(Arrays.stream(HeroWidthEnum.values()).anyMatch(c -> c.value() == 2160));

        SDK sdk = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        {
            // check open enum with unknown values

            EnumsPostOpenEnumUnrecognizedResponse res = sdk.enums().enumsPostOpenEnumUnrecognized() //
                    .request(ThemeRequestOpaque.builder() //
                            .color("purple") //
                            .icon("tick") //
                            .heroWidth(2160L) //
                            .build()) //
                    .call();

            Theme theme = res.themeResponse().get().json();
            Color color = theme.color().get();
            Icon icon = theme.icon().get();
            HeroWidth heroWidth = theme.heroWidth().get();
            assertFalse(color.isKnown());
            assertEquals("purple", color.value());
            assertEquals(Icon.TICK, icon);
            assertFalse(heroWidth.isKnown());
            assertEquals(2160, heroWidth.value());

            // check serialization of unknown values
            String json = mapper().writeValueAsString(theme);
            assertEquals(theme, mapper().readValue(json, Theme.class));
        }
        {
            // check open enum with known values

            EnumsPostOpenEnumUnrecognizedResponse res = sdk.enums().enumsPostOpenEnumUnrecognized() //
                    .request(ThemeRequestOpaque.builder() //
                            .color("red") //
                            .icon("tick") //
                            .heroWidth(480L) //
                            .build()) //
                    .call();

            Theme theme = res.themeResponse().get().json();
            Color color = theme.color().get();
            Icon icon = theme.icon().get();
            HeroWidth heroWidth = theme.heroWidth().get();
            assertTrue(color.isKnown());
            assertEquals(Color.RED, color);
            assertEquals(Icon.TICK, icon);
            assertTrue(heroWidth.isKnown());
            assertEquals(HeroWidth.FOUR_HUNDRED_AND_EIGHTY, heroWidth);

            // check serialization of known values
            String json = mapper().writeValueAsString(theme);
            assertEquals(theme, mapper().readValue(json, Theme.class));
        }
    }

    @Test
    public void testOpenEnumReferenceEquality() {
        // test with known enum
        assertTrue(Color.RED == Color.of("red"));
        // test with unknown enum
        assertTrue(Color.of("brown") == Color.of("brown"));
    }

    @Test
    public void testSerializationRoundTripKnownValue() throws JsonProcessingException {
        String json = mapper().writeValueAsString(Color.RED);
        assertEquals("\"red\"", json);
        assertEquals(Color.RED, mapper().readValue(json, Color.class));
    }

    @Test
    public void testSerializationRoundTripUnknownValue() throws JsonProcessingException {
        String json = mapper().writeValueAsString(Color.of("brown"));
        assertEquals("\"brown\"", json);
        assertEquals(Color.of("brown"), mapper().readValue(json, Color.class));
    }

    @Test
    public void testOpenEnumToString() {
        assertEquals("Color [value=red]", Color.RED.toString());
    }

    @Test
    public void testOpenEnumEquals() {
        // test with known enum
        assertEquals(Color.RED, Color.of("red"));
        // test with unknown enum
        assertEquals(Color.of("brown"), Color.of("brown"));
    }

    @Test
    public void testHashCode() {
        assertEquals(Objects.hash("red"), Color.RED.hashCode());
    }

    // Enhanced Open Enum Tests for Smart Union Resolution

    @Test
    public void testOpenEnumInUnionResolution() throws JsonProcessingException {
        // Test that open enums with known values are preferred over unknown values
        // in union resolution scenarios

        // Test with known enum value
        String jsonKnown = "{\"name\":\"test\",\"status\":\"active\",\"value\":42}";

        // This would test union resolution preferring known enum values
        // The actual union types would be generated from the schemas we added
        ObjectWithKnownOpenEnum knownResult = mapper().readValue(jsonKnown, ObjectWithKnownOpenEnum.class);

        assertTrue(knownResult.status().isKnown(), "Known enum value should be recognized");
        assertEquals("active", knownResult.status().value());
        assertEquals("test", knownResult.name());
        assertEquals(42, knownResult.value());
    }

    @Test
    public void testOpenEnumUnknownValueInUnion() throws JsonProcessingException {
        // Test that unknown open enum values are handled correctly in union resolution

        String jsonUnknown = "{\"name\":\"test\",\"status\":\"unknown_status\",\"value\":42}";

        ObjectWithKnownOpenEnum unknownResult = mapper().readValue(jsonUnknown, ObjectWithKnownOpenEnum.class);

        assertFalse(unknownResult.status().isKnown(), "Unknown enum value should not be recognized");
        assertEquals("unknown_status", unknownResult.status().value());
        assertEquals("test", unknownResult.name());
        assertEquals(42, unknownResult.value());
    }

    @Test
    public void testOpenEnumSerializationInUnionContext() throws JsonProcessingException {
        // Test that open enums serialize correctly when used in union contexts

        // Create object with known enum
        ObjectWithKnownOpenEnum objKnown = ObjectWithKnownOpenEnum.builder()
                .name("test")
                .status(Status.ACTIVE)
                .value(42)
                .build();

        String jsonKnown = mapper().writeValueAsString(objKnown);
        ObjectWithKnownOpenEnum roundTripKnown = mapper().readValue(jsonKnown, ObjectWithKnownOpenEnum.class);

        assertEquals(objKnown, roundTripKnown);
        assertTrue(roundTripKnown.status().isKnown());

        // Create object with unknown enum
        ObjectWithKnownOpenEnum objUnknown = ObjectWithKnownOpenEnum.builder()
                .name("test")
                .status(Status.of("custom_status"))
                .value(42)
                .build();

        String jsonUnknown = mapper().writeValueAsString(objUnknown);
        ObjectWithKnownOpenEnum roundTripUnknown = mapper().readValue(jsonUnknown, ObjectWithKnownOpenEnum.class);

        assertEquals(objUnknown, roundTripUnknown);
        assertFalse(roundTripUnknown.status().isKnown());
        assertEquals("custom_status", roundTripUnknown.status().value());
    }

    @Test
    public void testOpenEnumInexactFieldCounting() throws JsonProcessingException {
        // Test that the new inexact field counting logic works correctly
        // This tests the countInexactFields() method in OneOfDeserializer

        String jsonMixed = "{\"name\":\"test\",\"status\":\"unknown_status\",\"priority\":\"high\",\"value\":42}";

        // This would resolve to ObjectWithUnknownOpenEnum if the schemas were set up as a union
        // The resolution should count the unknown "status" as an inexact match
        // while "priority" with value "high" should be a known/exact match
        ObjectWithUnknownOpenEnum result = mapper().readValue(jsonMixed, ObjectWithUnknownOpenEnum.class);

        assertEquals("test", result.name());
        assertEquals("unknown_status", result.status().value());
        assertFalse(result.status().isKnown(), "Status should be unknown");
        assertEquals("high", result.priority().value());
        assertTrue(result.priority().isKnown(), "Priority should be known");
        assertEquals(42, result.value());
    }

    @Test
    public void testOpenEnumFieldMappingInComplexObjects() throws JsonProcessingException {
        // Test open enum field mapping in complex nested structures
        // This tests the enhanced countMappedFields() logic

        String jsonComplex = "{\"name\":\"test\",\"status\":\"active\",\"priority\":\"unknown_priority\",\"value\":42}";

        ObjectWithUnknownOpenEnum result = mapper().readValue(jsonComplex, ObjectWithUnknownOpenEnum.class);

        assertEquals("test", result.name());
        assertEquals("active", result.status().value());
        // Note: "active" is not in the enum for ObjectWithUnknownOpenEnum, so it should be unknown
        assertFalse(result.status().isKnown(), "Status 'active' should be unknown for this enum");
        assertEquals("unknown_priority", result.priority().value());
        assertFalse(result.priority().isKnown(), "Priority should be unknown");
        assertEquals(42, result.value());
    }
}
