package org.openapis.openapi;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.exc.ValueInstantiationException;
import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.shared.Bike;
import org.openapis.openapi.models.shared.Car;
import org.openapis.openapi.models.shared.DiscriminatedOpenEnumUnion;
import org.openapis.openapi.models.shared.ObjectWithOpenEnumStatus1;
import org.openapis.openapi.models.shared.UnknownDiscriminatedOpenEnumUnion;
import org.openapis.openapi.models.shared.UnknownVehicle;
import org.openapis.openapi.models.shared.Vehicle;
import org.openapis.openapi.utils.JSON;

import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

/**
 * Tests for open discriminated union (forward-compatible tagged union) pattern.
 *
 * Uses the generated Vehicle discriminated union:
 *   Car: vehicleType="car", wheelsType="four"
 *   Bike: vehicleType="bike", wheelsType="two", colour=string (required)
 *   UnknownVehicle: captures unknown discriminator values with raw JSON
 * Discriminator field: vehicleType
 */
public class OpenUnionAdditionalTest {

    @Test
    void testKnownVariant() throws Exception {
        CommonHelpers.recordTest("open-union-known-variant");

        // Known "car" discriminator parses to Car
        Vehicle car = JSON.getMapper().readValue(
                "{\"vehicleType\":\"car\",\"wheelsType\":\"four\"}",
                Vehicle.class);
        assertInstanceOf(Car.class, car);
        assertEquals("car", car.vehicleType());
        assertEquals("four", ((Car) car).wheelsType());

        // Known "bike" discriminator parses to Bike
        Vehicle bike = JSON.getMapper().readValue(
                "{\"vehicleType\":\"bike\",\"wheelsType\":\"two\",\"colour\":\"red\"}",
                Vehicle.class);
        assertInstanceOf(Bike.class, bike);
        assertEquals("bike", bike.vehicleType());
        assertEquals("red", ((Bike) bike).colour());

        // Serialization round-trip
        String carJson = JSON.getMapper().writeValueAsString(car);
        Vehicle restored = JSON.getMapper().readValue(carJson, Vehicle.class);
        assertInstanceOf(Car.class, restored);
        assertEquals("car", restored.vehicleType());
    }

    @Test
    void testUnknownDiscriminator() throws Exception {
        CommonHelpers.recordTest("open-union-unknown-discriminator");

        // Unknown discriminator -> UnknownVehicle with raw payload
        Vehicle result = JSON.getMapper().readValue(
                "{\"vehicleType\":\"spaceship\",\"thrust\":9000}",
                Vehicle.class);
        assertInstanceOf(UnknownVehicle.class, result);
        assertEquals("spaceship", result.vehicleType());

        UnknownVehicle unknown = (UnknownVehicle) result;
        assertEquals(9000, unknown.asJson().get("thrust").asInt());

        // Preserves full payload including nested objects
        Vehicle nested = JSON.getMapper().readValue(
                "{\"vehicleType\":\"hovercraft\",\"specs\":{\"weight\":500}}",
                Vehicle.class);
        assertInstanceOf(UnknownVehicle.class, nested);
        assertEquals(500, ((UnknownVehicle) nested).asJson().get("specs").get("weight").asInt());

        // Also test with DiscriminatedOpenEnumUnion (status discriminator)
        DiscriminatedOpenEnumUnion statusResult = JSON.getMapper().readValue(
                "{\"status\":\"suspended\",\"userId\":\"user-789\",\"activeAt\":\"2024-01-17T14:20:00Z\"}",
                DiscriminatedOpenEnumUnion.class);
        assertInstanceOf(UnknownDiscriminatedOpenEnumUnion.class, statusResult);
        assertEquals("suspended", statusResult.status());
        assertEquals("user-789", ((UnknownDiscriminatedOpenEnumUnion) statusResult).asJson().get("userId").asText());
    }

    @Test
    void testMissingDiscriminator() throws Exception {
        CommonHelpers.recordTest("open-union-missing-discriminator");

        // Missing discriminator field -> UnknownVehicle (via defaultImpl)
        Vehicle result = JSON.getMapper().readValue(
                "{\"wheelsType\":\"four\"}",
                Vehicle.class);
        assertInstanceOf(UnknownVehicle.class, result);
        assertEquals("UNKNOWN", result.vehicleType());
    }

    @Test
    void testInvalidPayload() throws Exception {
        CommonHelpers.recordTest("open-union-invalid-payload");

        // In Java, non-object payloads fall back to UnknownVehicle via defaultImpl
        // (Jackson wraps the raw JsonNode regardless of type)

        // String payload -> UnknownVehicle
        Vehicle strResult = JSON.getMapper().readValue("\"hello\"", Vehicle.class);
        assertInstanceOf(UnknownVehicle.class, strResult);
        assertEquals("UNKNOWN", strResult.vehicleType());

        // Number payload -> UnknownVehicle
        Vehicle numResult = JSON.getMapper().readValue("42", Vehicle.class);
        assertInstanceOf(UnknownVehicle.class, numResult);
        assertEquals("UNKNOWN", numResult.vehicleType());

        // Null payload -> null
        Vehicle nullResult = JSON.getMapper().readValue("null", Vehicle.class);
        assertEquals(null, nullResult);
    }

    @Test
    void testKnownDiscInvalidSchema() throws Exception {
        CommonHelpers.recordTest("open-union-known-disc-invalid-schema");

        // Known "bike" discriminator but missing required "colour" field
        assertThrows(ValueInstantiationException.class, () ->
                JSON.getMapper().readValue(
                        "{\"vehicleType\":\"bike\",\"wheelsType\":\"two\"}",
                        Vehicle.class));
    }

    @Test
    void testEmbedded() throws Exception {
        CommonHelpers.recordTest("open-union-embedded");

        // List of mixed known and unknown vehicles
        List<Vehicle> vehicles = JSON.getMapper().readValue(
                "[{\"vehicleType\":\"car\",\"wheelsType\":\"four\"},"
                        + "{\"vehicleType\":\"spaceship\",\"thrust\":9000},"
                        + "{\"vehicleType\":\"bike\",\"wheelsType\":\"two\",\"colour\":\"green\"}]",
                new TypeReference<List<Vehicle>>() {});

        assertEquals(3, vehicles.size());
        assertInstanceOf(Car.class, vehicles.get(0));
        assertInstanceOf(UnknownVehicle.class, vehicles.get(1));
        assertInstanceOf(Bike.class, vehicles.get(2));
        assertEquals("spaceship", vehicles.get(1).vehicleType());

        // Nested in a JSON object (parse parent, extract vehicle)
        JsonNode wrapper = JSON.getMapper().readTree(
                "{\"label\":\"mystery\",\"vehicle\":{\"vehicleType\":\"spaceship\",\"thrust\":9000}}");
        Vehicle embedded = JSON.getMapper().treeToValue(wrapper.get("vehicle"), Vehicle.class);
        assertInstanceOf(UnknownVehicle.class, embedded);
        assertEquals("spaceship", embedded.vehicleType());
    }
}
