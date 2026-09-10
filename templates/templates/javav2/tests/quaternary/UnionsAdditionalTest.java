package org.openapis.quaternary.openapi;

import com.fasterxml.jackson.databind.exc.InvalidTypeIdException;
import org.junit.jupiter.api.Test;
import org.openapis.quaternary.openapi.models.shared.Car;
import org.openapis.quaternary.openapi.models.shared.Vehicle;
import org.openapis.quaternary.openapi.utils.JSON;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertThrows;

/**
 * Quaternary sets forwardCompatibleUnionsByDefault: false, so the Vehicle
 * discriminated union is generated CLOSED: no UnknownVehicle fallback class.
 *
 * Open-behavior counterparts live in primary's OpenUnionAdditionalTest.
 */
public class UnionsAdditionalTest {

    // Primary counterpart (open behavior): open-union-known-variant
    @Test
    void testClosedUnionKnownVariantParses() throws Exception {
        Vehicle car = JSON.getMapper().readValue(
                "{\"vehicleType\":\"car\",\"wheelsType\":\"four\"}",
                Vehicle.class);
        assertInstanceOf(Car.class, car);
        assertEquals("car", car.vehicleType());
    }

    // Primary counterpart (open behavior): open-union-unknown-discriminator
    @Test
    void testClosedUnionUnknownDiscriminatorThrows() {
        assertThrows(InvalidTypeIdException.class, () -> JSON.getMapper().readValue(
                "{\"vehicleType\":\"spaceship\",\"thrust\":9000}",
                Vehicle.class));
    }

    // Primary counterpart (open behavior): open-union-embedded (artifact presence)
    @Test
    void testOpenUnionArtifactsAbsent() {
        assertThrows(ClassNotFoundException.class, () ->
                Class.forName("org.openapis.quaternary.openapi.models.shared.UnknownVehicle"));
        assertThrows(ClassNotFoundException.class, () ->
                Class.forName("org.openapis.quaternary.openapi.utils.UnknownType"));
    }
}
