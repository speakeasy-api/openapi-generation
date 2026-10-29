import json
from typing import List

import httpx
import pydantic
import pytest
from pydantic import TypeAdapter

from openapi import SDK, utils
from openapi.models import shared
from .common_helpers import HTTPBIN_URL


UNKNOWN_CHOICE = {"kind": "future", "weight": 9000, "detail": {"enabled": False, "tags": ["a", None]}}
PAYLOAD = {"label": "root", "choice": UNKNOWN_CHOICE, "choices": [{"kind": "text", "text": "t"}, UNKNOWN_CHOICE]}


def receive(payload):
    transport = httpx.MockTransport(lambda request: httpx.Response(200, json={"json": payload}))
    with httpx.Client(transport=transport) as client:
        with SDK(server_url=HTTPBIN_URL, client=client) as sdk:
            return sdk.echo_unknown_fields(label="root").object.json_


@pytest.mark.parametrize("build", [
    lambda received: shared.FieldEnvelopeInput(label=received.label, choice=received.choice, choices=received.choices),
    lambda received: shared.FieldEnvelopeInput.model_validate({"label": received.label, "choice": received.choice, "choices": received.choices}),
])
def test_received_unknown_variants_are_marshalled_as_raw_payload(build):
    received = receive(PAYLOAD)
    assert received.choice.is_unknown and received.choices[1].is_unknown
    request = build(received)
    assert json.loads(utils.marshal_json(request, shared.FieldEnvelopeInput)) == PAYLOAD
    assert request.to_dict() == PAYLOAD
    assert json.loads(request.to_json()) == PAYLOAD


def test_received_unknown_variants_are_sent_as_raw_payload():
    received = receive(PAYLOAD)
    sent = []
    with httpx.Client(event_hooks={"request": [sent.append]}) as client:
        with SDK(server_url=HTTPBIN_URL, client=client) as sdk:
            response = sdk.echo_unknown_fields(label=received.label, choice=received.choice, choices=received.choices)
    assert json.loads(sent[-1].content) == PAYLOAD
    echoed = response.object.json_
    assert isinstance(echoed.choice, shared.UnknownChoice)
    assert echoed.to_dict() == PAYLOAD
    assert echoed.model_dump(mode="json", by_alias=True, warnings="error") == PAYLOAD


def test_top_level_unknown_variant_is_marshalled_as_raw_payload():
    unknown = receive(PAYLOAD).choice
    assert json.loads(utils.marshal_json(unknown, shared.Choice)) == UNKNOWN_CHOICE
    assert json.loads(utils.marshal_json([unknown], List[shared.Choice])) == [UNKNOWN_CHOICE]


def test_unknown_variant_dumps_as_raw_payload():
    received = receive(PAYLOAD)
    unknown = received.choice
    assert isinstance(unknown, shared.UnknownChoice)
    assert unknown.raw == UNKNOWN_CHOICE
    assert unknown.model_dump(warnings="error") == UNKNOWN_CHOICE
    assert unknown.model_dump(mode="json", warnings="error") == UNKNOWN_CHOICE
    assert unknown.model_dump(by_alias=True, warnings="error") == UNKNOWN_CHOICE
    assert json.loads(unknown.model_dump_json(warnings="error")) == UNKNOWN_CHOICE
    assert unknown.to_dict() == UNKNOWN_CHOICE
    assert json.loads(unknown.to_json()) == UNKNOWN_CHOICE
    assert received.model_dump(by_alias=True, warnings="error")["choice"] == UNKNOWN_CHOICE
    assert received.model_dump(by_alias=True, warnings="error")["choices"] == PAYLOAD["choices"]
    assert json.loads(received.model_dump_json(by_alias=True, warnings="error")) == PAYLOAD


@pytest.mark.parametrize("build", [
    lambda: shared.FieldEnvelopeInput.model_validate(PAYLOAD),
    lambda: shared.FieldEnvelopeInput(label="root", choice=UNKNOWN_CHOICE),
    lambda: shared.FieldEnvelopeInput.model_validate(receive(PAYLOAD).to_dict()),
])
def test_unknown_discriminator_in_request_dicts_still_raises(build):
    with pytest.raises(pydantic.ValidationError, match="unrecognized kind value 'future'"):
        build()


def test_serialization_schema_describes_raw_payload():
    validation = TypeAdapter(shared.Choice).json_schema(mode="validation")
    serialization = TypeAdapter(shared.Choice).json_schema(mode="serialization")
    assert set(validation["$defs"]["UnknownChoice"]["properties"]) == {"kind", "raw", "is_unknown"}
    assert serialization["$defs"]["UnknownChoice"] == {}
