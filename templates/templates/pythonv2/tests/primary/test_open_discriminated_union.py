"""
Test suite for open discriminated union (forward-compatible tagged union) pattern.

Tests use the generated Vehicle discriminated union which has:
- Car: vehicleType="car", wheelsType="four"
- Bike: vehicleType="bike", wheelsType="two", colour=str
- UnknownVehicle: vehicle_type=Literal["UNKNOWN"], raw=Any, is_unknown=True, frozen=True
Discriminator field: vehicleType (JSON alias) / vehicle_type (Python)

Forward-compatibility contract (mirrors TypeScript, where only the inbound
schema is open):
- The Unknown fallback applies only when validating with the
  ALLOW_UNKNOWN_UNION_VARIANTS context flag, which the SDK sets on response
  deserialization. There:
  - Unknown discriminator value -> Unknown fallback with raw payload
  - Known discriminator whose payload fails variant validation -> Unknown fallback
- Without the flag (user-built request payloads), both cases raise
  ValidationError so mistakes surface locally instead of going on the wire.
- Missing discriminator / non-dict payload -> ValidationError in both modes
  (lets pydantic try sibling union branches, e.g. None in Optional[Vehicle])
"""

import pytest
from pydantic import BaseModel, TypeAdapter, ValidationError

from openapi.models.shared.vehicle import Vehicle, UnknownVehicle
from openapi.models.shared.car import Car
from openapi.models.shared.bike import Bike
from openapi.utils import ALLOW_UNKNOWN_UNION_VARIANTS
from .common_helpers import record_test

RESPONSE_CONTEXT = {ALLOW_UNKNOWN_UNION_VARIANTS: True}


class TestOpenUnionKnownVariant:
    """Known discriminator values parse, serialize, and validate correctly."""

    def test_known_variant(self):
        record_test("open-union-known-variant")

        adapter = TypeAdapter(Vehicle)

        # Known "car" discriminator parses to Car
        car = adapter.validate_python({"vehicleType": "car", "wheelsType": "four"})
        assert isinstance(car, Car)
        assert car.vehicle_type == "car"
        assert car.wheels_type == "four"

        # Known "bike" discriminator parses to Bike
        bike = adapter.validate_python(
            {"vehicleType": "bike", "wheelsType": "two", "colour": "red"}
        )
        assert isinstance(bike, Bike)
        assert bike.vehicle_type == "bike"
        assert bike.colour == "red"

        # Serialization round-trip
        data = car.model_dump(by_alias=True)
        assert data == {"vehicleType": "car", "wheelsType": "four"}
        restored = adapter.validate_python(data)
        assert isinstance(restored, Car)

        # JSON round-trip
        json_str = car.model_dump_json(by_alias=True)
        restored_json = Car.model_validate_json(json_str)
        assert isinstance(restored_json, Car)
        assert restored_json.wheels_type == "four"


class TestOpenUnionUnknownDiscriminator:
    """Unknown discriminator values produce Unknown fallback with raw payload
    when validating with the response context."""

    def test_unknown_discriminator(self):
        record_test("open-union-unknown-discriminator")

        adapter = TypeAdapter(Vehicle)

        # Unknown discriminator -> UnknownVehicle with raw payload
        payload = {"vehicleType": "spaceship", "thrust": 9000}
        result = adapter.validate_python(payload, context=RESPONSE_CONTEXT)
        assert isinstance(result, UnknownVehicle)
        assert result.vehicle_type == "UNKNOWN"
        assert result.raw == payload
        assert result.raw["vehicleType"] == "spaceship"
        assert result.is_unknown is True

        # Preserves full payload including nested objects
        nested = {"vehicleType": "hovercraft", "specs": {"weight": 500}}
        result2 = adapter.validate_python(nested, context=RESPONSE_CONTEXT)
        assert isinstance(result2, UnknownVehicle)
        assert result2.vehicle_type == "UNKNOWN"
        assert result2.raw["specs"]["weight"] == 500

        # Unknown variant can be model_dump'd
        data = result.model_dump()
        assert data["vehicle_type"] == "UNKNOWN"
        assert data["raw"] == payload
        assert data["is_unknown"] is True

        # Unknown variant is frozen (immutable)
        with pytest.raises(ValidationError):
            result.vehicle_type = "other"


class TestOpenUnionStrictWithoutResponseContext:
    """Without the response context flag — the request-building path — invalid
    payloads raise locally instead of degrading to the Unknown fallback and
    being sent to the server."""

    def test_unknown_discriminator_raises(self):
        record_test("open-union-strict-request-validation")

        adapter = TypeAdapter(Vehicle)

        with pytest.raises(ValidationError):
            adapter.validate_python({"vehicleType": "spaceship", "thrust": 9000})

    def test_known_disc_invalid_payload_raises(self):
        record_test("open-union-strict-request-validation")

        adapter = TypeAdapter(Vehicle)

        # Known "bike" discriminator with a misspelled required field must
        # surface the variant's error, not silently become Unknown
        with pytest.raises(ValidationError) as exc_info:
            adapter.validate_python({"vehicleType": "bike", "wheelsType": "two"})
        assert "colour" in str(exc_info.value)

    def test_known_disc_invalid_embedded_in_parent_raises(self):
        record_test("open-union-strict-request-validation")

        class Garage(BaseModel):
            name: str
            vehicle: Vehicle

        with pytest.raises(ValidationError):
            Garage.model_validate({
                "name": "my garage",
                "vehicle": {"vehicleType": "bike", "wheelsType": "two"},
            })

    def test_received_unknown_instance_passes_through(self):
        record_test("open-union-strict-request-validation")

        adapter = TypeAdapter(Vehicle)

        # An Unknown instance received from a response can be echoed back
        # into a request without re-validation rejecting it
        received = adapter.validate_python(
            {"vehicleType": "spaceship", "thrust": 9000}, context=RESPONSE_CONTEXT
        )
        assert adapter.validate_python(received) is received


class TestOpenUnionMissingDiscriminator:
    """Missing discriminator field raises validation error."""

    def test_missing_discriminator(self):
        record_test("open-union-missing-discriminator")

        adapter = TypeAdapter(Vehicle)

        with pytest.raises(ValidationError):
            adapter.validate_python({"wheelsType": "four"})

        with pytest.raises(ValidationError):
            adapter.validate_python({"wheelsType": "four"}, context=RESPONSE_CONTEXT)


class TestOpenUnionInvalidPayload:
    """Non-object payloads (null, primitives) raise validation error."""

    def test_invalid_payload(self):
        record_test("open-union-invalid-payload")

        adapter = TypeAdapter(Vehicle)

        with pytest.raises(ValidationError):
            adapter.validate_python(None, context=RESPONSE_CONTEXT)

        with pytest.raises(ValidationError):
            adapter.validate_python("not an object", context=RESPONSE_CONTEXT)

        with pytest.raises(ValidationError):
            adapter.validate_python(42, context=RESPONSE_CONTEXT)


class TestOpenUnionKnownDiscInvalidSchema:
    """Known discriminator whose payload fails variant validation falls back
    to the Unknown variant when validating with the response context,
    preserving the raw payload (matches TypeScript discriminatedUnion
    behavior). Servers may stream partial variants (for example a
    `step.start` event carrying a known step type without its required
    fields yet); the SDK must degrade gracefully, not crash."""

    def test_known_disc_missing_required_field(self):
        record_test("open-union-known-disc-invalid-schema")

        adapter = TypeAdapter(Vehicle)

        # "bike" is known but missing required "colour" field
        payload = {"vehicleType": "bike", "wheelsType": "two"}
        result = adapter.validate_python(payload, context=RESPONSE_CONTEXT)
        assert isinstance(result, UnknownVehicle)
        assert result.vehicle_type == "UNKNOWN"
        assert result.is_unknown is True
        assert result.raw == payload

    def test_known_disc_wrong_field_type(self):
        record_test("open-union-known-disc-invalid-schema")

        adapter = TypeAdapter(Vehicle)

        # "bike" is known but "colour" has the wrong type
        payload = {"vehicleType": "bike", "wheelsType": "two", "colour": 123}
        result = adapter.validate_python(payload, context=RESPONSE_CONTEXT)
        assert isinstance(result, UnknownVehicle)
        assert result.vehicle_type == "UNKNOWN"
        assert result.raw == payload

    def test_known_disc_wrong_const_value(self):
        record_test("open-union-known-disc-invalid-schema")

        adapter = TypeAdapter(Vehicle)

        # "car" is known but "wheelsType" violates its const
        payload = {"vehicleType": "car", "wheelsType": "three"}
        result = adapter.validate_python(payload, context=RESPONSE_CONTEXT)
        assert isinstance(result, UnknownVehicle)
        assert result.vehicle_type == "UNKNOWN"
        assert result.raw == payload

    def test_known_disc_fallback_preserves_raw_exactly(self):
        record_test("open-union-known-disc-invalid-schema")

        adapter = TypeAdapter(Vehicle)

        # Extra/nested data in an invalid known-variant payload survives intact
        payload = {
            "vehicleType": "bike",
            "wheelsType": "two",
            "extras": {"nested": [1, 2, 3]},
        }
        result = adapter.validate_python(payload, context=RESPONSE_CONTEXT)
        assert isinstance(result, UnknownVehicle)
        assert result.raw == payload
        assert result.raw["extras"]["nested"] == [1, 2, 3]

        # Fallback still serializes and stays frozen
        data = result.model_dump()
        assert data["vehicle_type"] == "UNKNOWN"
        assert data["raw"] == payload
        with pytest.raises(ValidationError):
            result.vehicle_type = "other"

    def test_known_disc_invalid_embedded_in_parent(self):
        record_test("open-union-known-disc-invalid-schema")

        class Garage(BaseModel):
            name: str
            vehicle: Vehicle

        # Parent model still validates; invalid known variant degrades inline
        garage = Garage.model_validate({
            "name": "my garage",
            "vehicle": {"vehicleType": "bike", "wheelsType": "two"},
        }, context=RESPONSE_CONTEXT)
        assert isinstance(garage.vehicle, UnknownVehicle)
        assert garage.vehicle.raw == {"vehicleType": "bike", "wheelsType": "two"}

    def test_known_disc_invalid_in_mixed_list(self):
        record_test("open-union-known-disc-invalid-schema")

        class Fleet(BaseModel):
            vehicles: list[Vehicle]

        fleet = Fleet.model_validate({
            "vehicles": [
                {"vehicleType": "car", "wheelsType": "four"},
                {"vehicleType": "bike", "wheelsType": "two"},  # invalid known
                {"vehicleType": "spaceship", "thrust": 9000},  # unknown disc
            ],
        }, context=RESPONSE_CONTEXT)
        assert isinstance(fleet.vehicles[0], Car)
        assert isinstance(fleet.vehicles[1], UnknownVehicle)
        assert isinstance(fleet.vehicles[2], UnknownVehicle)

    def test_known_disc_invalid_non_model_variant_falls_back(self):
        record_test("open-union-known-disc-invalid-schema")

        from openapi.models.shared.nesteddiscunion import (
            NestedDiscUnion,
            UnknownNestedDiscUnion,
        )
        from openapi.models.shared.typea1 import TypeA1

        adapter = TypeAdapter(NestedDiscUnion)

        # NestedDiscUnion maps "a" -> TypeA, itself a Union[TypeA1, TypeA2]
        # (validated via TypeAdapter, not BaseModel.model_validate).
        valid = adapter.validate_python({"type": "a", "value": "hello"})
        assert isinstance(valid, TypeA1)

        # Known "a" discriminator but matches neither TypeA1 nor TypeA2
        payload = {"type": "a"}
        result = adapter.validate_python(payload, context=RESPONSE_CONTEXT)
        assert isinstance(result, UnknownNestedDiscUnion)
        assert result.type == "UNKNOWN"
        assert result.raw == payload

    def test_optional_union_none_is_not_swallowed(self):
        record_test("open-union-known-disc-invalid-schema")

        from typing import Optional

        class Slot(BaseModel):
            vehicle: Optional[Vehicle] = None

        # None must stay None, not become an Unknown fallback
        slot = Slot.model_validate({"vehicle": None}, context=RESPONSE_CONTEXT)
        assert slot.vehicle is None

        slot_default = Slot.model_validate({}, context=RESPONSE_CONTEXT)
        assert slot_default.vehicle is None


class TestOpenUnionEmbedded:
    """Open union works when composed in parent models and lists."""

    def test_embedded(self):
        record_test("open-union-embedded")

        class Garage(BaseModel):
            name: str
            vehicle: Vehicle

        # Known variant embedded in parent model
        garage_known = Garage.model_validate({
            "name": "my garage",
            "vehicle": {"vehicleType": "bike", "wheelsType": "two", "colour": "blue"},
        })
        assert isinstance(garage_known.vehicle, Bike)
        assert garage_known.vehicle.colour == "blue"

        # Unknown variant embedded in parent model
        garage_unknown = Garage.model_validate({
            "name": "my garage",
            "vehicle": {"vehicleType": "spaceship", "thrust": 9000},
        }, context=RESPONSE_CONTEXT)
        assert isinstance(garage_unknown.vehicle, UnknownVehicle)
        assert garage_unknown.vehicle.vehicle_type == "UNKNOWN"
        assert garage_unknown.vehicle.raw == {"vehicleType": "spaceship", "thrust": 9000}

        # List of mixed known and unknown vehicles
        class Fleet(BaseModel):
            vehicles: list[Vehicle]

        fleet = Fleet.model_validate({
            "vehicles": [
                {"vehicleType": "car", "wheelsType": "four"},
                {"vehicleType": "spaceship", "thrust": 9000},
                {"vehicleType": "bike", "wheelsType": "two", "colour": "green"},
            ],
        }, context=RESPONSE_CONTEXT)
        assert len(fleet.vehicles) == 3
        assert isinstance(fleet.vehicles[0], Car)
        assert isinstance(fleet.vehicles[1], UnknownVehicle)
        assert isinstance(fleet.vehicles[2], Bike)
