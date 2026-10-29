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
            return sdk.echo_unknown_fields(request={"label": "root"}).json_


def echo(request):
    sent = []
    with httpx.Client(event_hooks={"request": [sent.append]}) as client:
        with SDK(server_url=HTTPBIN_URL, client=client) as sdk:
            response = sdk.echo_unknown_fields(request=request)
    return response.json_, json.loads(sent[-1].content)


@pytest.mark.parametrize("build", [
    lambda received: shared.FieldEnvelope(label=received.label, choice=received.choice, choices=received.choices),
    lambda received: shared.FieldEnvelope.model_validate({"label": received.label, "choice": received.choice, "choices": received.choices}),
])
def test_received_unknown_variants_are_sent_as_raw_payload(build):
    received = receive(PAYLOAD)
    assert received.choice.is_unknown and received.choices[1].is_unknown
    request = build(received)
    assert json.loads(utils.marshal_json(request, shared.FieldEnvelope)) == PAYLOAD
    echoed, sent = echo(request)
    assert sent == PAYLOAD
    assert isinstance(echoed.choice, shared.UnknownChoice)
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
    assert received.model_dump(by_alias=True, warnings="error")["choice"] == UNKNOWN_CHOICE
    assert received.model_dump(by_alias=True, warnings="error")["choices"] == PAYLOAD["choices"]
    assert json.loads(received.model_dump_json(by_alias=True, warnings="error")) == PAYLOAD


@pytest.mark.parametrize("build", [
    lambda: shared.FieldEnvelope.model_validate(PAYLOAD),
    lambda: shared.FieldEnvelope(**PAYLOAD),
    lambda: shared.FieldEnvelope.model_validate(receive(PAYLOAD).model_dump(mode="json", by_alias=True, warnings="error")),
])
def test_request_dicts_with_unknown_discriminator_are_sent_as_raw_payload(build):
    request = build()
    assert isinstance(request.choice, shared.UnknownChoice) and isinstance(request.choices[1], shared.UnknownChoice)
    echoed, sent = echo(request)
    assert sent == PAYLOAD
    assert echoed.model_dump(mode="json", by_alias=True, warnings="error") == PAYLOAD


def test_request_dict_with_unknown_discriminator_is_sent_through_sdk():
    _, sent = echo(PAYLOAD)
    assert sent == PAYLOAD


@pytest.mark.parametrize("build", [
    lambda: shared.FieldEnvelope.model_validate(INVALID_PAYLOAD),
    lambda: shared.FieldEnvelope(**INVALID_PAYLOAD),
    lambda: echo(INVALID_PAYLOAD),
])
def test_known_variant_with_invalid_payload_in_request_raises(build):
    with pytest.raises(pydantic.ValidationError, match="Input should be a valid string"):
        build()


def test_known_variant_with_invalid_payload_is_marshalled_as_raw_payload():
    invalid = {"kind": "text", "text": 3}
    received = receive({"label": "root", "choice": invalid}).choice
    assert isinstance(received, shared.UnknownChoice)
    assert received.model_dump(warnings="error") == invalid
    assert json.loads(utils.marshal_json(received, shared.Choice)) == invalid


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
    request = shared.FieldEnvelope.model_validate({"label": "root", "history": HISTORY})
    assert isinstance(request.history[0], shared.NoteStep)
    assert isinstance(request.history[0].content[1], shared.UnknownChoice)
    assert isinstance(request.history[1], shared.UnknownHistoryStep)
    echoed, sent = echo(request)
    assert sent == {"label": "root", "history": HISTORY}
    assert echoed.model_dump(mode="json", by_alias=True, warnings="error")["history"] == HISTORY


def test_received_history_is_sent_back_as_raw_payload():
    received = receive({"label": "root", "history": HISTORY})
    assert isinstance(received.history[1], shared.UnknownHistoryStep)
    _, sent = echo(shared.FieldEnvelope(label="root", history=received.history))
    assert sent == {"label": "root", "history": HISTORY}


@pytest.mark.parametrize("history", INVALID_HISTORIES)
def test_request_history_with_invalid_known_variant_raises(history):
    with pytest.raises(pydantic.ValidationError):
        shared.FieldEnvelope.model_validate({"label": "root", "history": history})
    with pytest.raises(pydantic.ValidationError):
        echo({"label": "root", "history": history})


@pytest.mark.parametrize("history", UNKNOWN_HISTORIES)
def test_request_history_with_unknown_variants_is_sent_as_raw_payload(history):
    _, sent = echo({"label": "root", "history": history})
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
    ("stepHistory", [{"kind": "text", "text": 3}, {"kind": "pause", "seconds": 1}]),
    ("stepHistory", [{"kind": "pause", "seconds": 1}, {"kind": "text", "text": "t"}]),
]
UNKNOWN_FIELDS = [
    ("nullableHistory", [{"kind": "future_step"}, None]),
    ("nestedHistory", [[{"kind": "future_step"}]]),
    ("historyByName", {"a": {"kind": "future_step"}}),
    ("freeformHistory", {"kind": "text", "text": 3}),
    ("stepHistory", [{"kind": "future_step"}, {"kind": "pause", "seconds": 1}]),
]


@pytest.mark.parametrize("payload", STORED_RESPONSES)
def test_stored_response_with_variants_of_unrelated_unions_is_sent_back(payload):
    stored = receive(payload).model_dump(mode="json", by_alias=True, warnings="error")
    assert shared.FieldEnvelope.model_validate(stored).model_dump(mode="json", by_alias=True) == stored
    _, sent = echo(stored)
    assert sent == payload


@pytest.mark.parametrize("field, value", SIBLING_FIELDS)
def test_variant_of_another_open_union_member_raises(field, value):
    with pytest.raises(pydantic.ValidationError, match="belongs to another member of the enclosing union"):
        shared.FieldEnvelope.model_validate({"label": "root", field: value})
    with pytest.raises(pydantic.ValidationError):
        echo({"label": "root", field: value})


@pytest.mark.parametrize("field, value", UNKNOWN_FIELDS)
def test_value_no_open_union_member_recognizes_is_sent_as_given(field, value):
    _, sent = echo({"label": "root", field: value})
    assert sent == {"label": "root", field: value}
