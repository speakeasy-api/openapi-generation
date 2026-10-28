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
INVALID_PAYLOAD = {"label": "root", "choice": {"kind": "text", "text": 3}}


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


def send(**kwargs):
    sent = []
    with httpx.Client(event_hooks={"request": [sent.append]}) as client:
        with SDK(server_url=HTTPBIN_URL, client=client) as sdk:
            response = sdk.echo_unknown_fields(**kwargs)
    return response.object.json_, json.loads(sent[-1].content)


@pytest.mark.parametrize("build", [
    lambda: shared.FieldEnvelopeInput.model_validate(PAYLOAD),
    lambda: shared.FieldEnvelopeInput(**PAYLOAD),
    lambda: shared.FieldEnvelopeInput.model_validate(receive(PAYLOAD).to_dict()),
    lambda: shared.FieldEnvelopeInput.model_validate(json.loads(receive(PAYLOAD).to_json())),
])
def test_request_dicts_with_unknown_discriminator_are_marshalled_as_raw_payload(build):
    request = build()
    assert request.choice.is_unknown and request.choices[1].is_unknown
    assert json.loads(utils.marshal_json(request, shared.FieldEnvelopeInput)) == PAYLOAD
    assert request.to_dict() == PAYLOAD


def test_stored_response_dicts_are_sent_as_raw_payload():
    stored = receive(PAYLOAD).to_dict()
    echoed, sent = send(label=stored["label"], choice=stored["choice"], choices=stored["choices"])
    assert sent == PAYLOAD
    assert echoed.to_dict() == PAYLOAD


@pytest.mark.parametrize("build", [
    lambda: shared.FieldEnvelopeInput.model_validate(INVALID_PAYLOAD),
    lambda: shared.FieldEnvelopeInput(**INVALID_PAYLOAD),
    lambda: send(**INVALID_PAYLOAD),
])
def test_known_variant_with_invalid_payload_in_request_raises(build):
    with pytest.raises(pydantic.ValidationError, match="Input should be a valid string"):
        build()


def test_serialization_schema_describes_raw_payload():
    validation = TypeAdapter(shared.Choice).json_schema(mode="validation")
    serialization = TypeAdapter(shared.Choice).json_schema(mode="serialization")
    assert set(validation["$defs"]["UnknownChoice"]["properties"]) == {"kind", "raw", "is_unknown"}
    assert serialization["$defs"]["UnknownChoice"] == {}


HISTORY = [
    {"kind": "note", "content": [{"kind": "text", "text": "t"}, UNKNOWN_CHOICE]},
    {"kind": "future_step", "minutes": 5},
]
NOTE = {"kind": "note", "content": []}
INVALID_HISTORIES = [
    [{"kind": "note", "content": [{"kind": "text", "text": 7}]}],
    [{"kind": "note", "content": [{"kind": "text", "label": "no text"}]}],
    [{"kind": "text", "text": 3}],
    [{"kind": "pause", "seconds": "long"}],
    [{"kind": "note", "content": [{"kind": "text", "text": 7}]}, {"kind": "text", "text": "t"}],
    [NOTE, {"kind": "text", "text": 7}],
    [NOTE, {"kind": "text", "text": "t"}],
    [{"kind": "text", "text": "t"}, {"kind": "pause", "seconds": 5}],
    NOTE,
]
UNKNOWN_HISTORIES = [
    [{"kind": "future_step"}],
    [{"kind": "nots", "content": [{"kind": "text", "text": "t"}]}],
    [{"kind": "note", "content": [{"kind": "future", "nested": [{"kind": "later"}]}]}],
]


def test_request_history_with_nested_unknown_variants_is_sent_as_raw_payload():
    request = shared.FieldEnvelopeInput.model_validate({"label": "root", "history": HISTORY})
    assert isinstance(request.history[0], shared.NoteStep)
    assert isinstance(request.history[0].content[1], shared.UnknownChoice)
    assert isinstance(request.history[1], shared.UnknownHistoryStep)
    assert request.to_dict()["history"] == HISTORY
    echoed, sent = send(label="root", history=request.history)
    assert sent == {"label": "root", "history": HISTORY}
    assert echoed.to_dict()["history"] == HISTORY


def test_stored_history_is_sent_back_as_raw_payload():
    stored = receive({"label": "root", "history": HISTORY}).to_dict()
    _, sent = send(label=stored["label"], history=stored["history"])
    assert sent == {"label": "root", "history": HISTORY}


@pytest.mark.parametrize("history", INVALID_HISTORIES)
def test_request_history_with_invalid_known_variant_raises(history):
    with pytest.raises(pydantic.ValidationError):
        shared.FieldEnvelopeInput.model_validate({"label": "root", "history": history})
    with pytest.raises(pydantic.ValidationError):
        send(label="root", history=history)


@pytest.mark.parametrize("history", UNKNOWN_HISTORIES)
def test_request_history_with_unknown_variants_is_sent_as_raw_payload(history):
    _, sent = send(label="root", history=history)
    assert sent == {"label": "root", "history": history}


STORED_RESPONSES = [
    {"label": "root", "choice": NOTE},
    {"label": "root", "choices": [NOTE, {"kind": "pause", "seconds": 1}]},
    {"label": "root", "history": [{"kind": "note", "content": [{"kind": "pause", "seconds": 1}]}]},
]
SIBLING_FIELDS = [
    ("nullableHistory", [{"kind": "text", "text": 3}]),
    ("nullableHistory", [NOTE, {"kind": "text", "text": "t"}]),
    ("nullableHistory", [{"kind": "text", "text": "t"}, NOTE]),
    ("nestedHistory", [[{"kind": "text", "text": 3}]]),
    ("nestedHistory", [[NOTE, {"kind": "text", "text": "t"}]]),
    ("historyByName", {"a": NOTE, "b": {"kind": "text", "text": "t"}}),
    ("historyByName", {"a": {"kind": "text", "text": "t"}, "b": NOTE}),
]
UNKNOWN_FIELDS = [
    ("nullableHistory", [{"kind": "future_step"}, None]),
    ("nestedHistory", [[{"kind": "future_step"}]]),
    ("historyByName", {"a": {"kind": "future_step"}}),
    ("freeformHistory", {"kind": "text", "text": 3}),
]

ARGUMENT_NAMES = {
    "nullableHistory": "nullable_history",
    "nestedHistory": "nested_history",
    "historyByName": "history_by_name",
    "freeformHistory": "freeform_history",
}


@pytest.mark.parametrize("payload", STORED_RESPONSES)
def test_stored_response_with_variants_of_unrelated_unions_is_sent_back(payload):
    stored = receive(payload).to_dict()
    assert shared.FieldEnvelopeInput.model_validate(stored).to_dict() == stored
    _, sent = send(**stored)
    assert sent == payload


@pytest.mark.parametrize("field, value", SIBLING_FIELDS)
def test_variant_of_another_open_union_member_raises(field, value):
    with pytest.raises(pydantic.ValidationError, match="belongs to another member of the enclosing union"):
        shared.FieldEnvelopeInput.model_validate({"label": "root", field: value})
    with pytest.raises(pydantic.ValidationError):
        send(label="root", **{ARGUMENT_NAMES[field]: value})


@pytest.mark.parametrize("field, value", UNKNOWN_FIELDS)
def test_value_no_open_union_member_recognizes_is_sent_as_given(field, value):
    _, sent = send(label="root", **{ARGUMENT_NAMES[field]: value})
    assert sent == {"label": "root", field: value}
