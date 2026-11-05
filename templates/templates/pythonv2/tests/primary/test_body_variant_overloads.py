import inspect
import json
import subprocess
import sys
from pathlib import Path

import httpx
import pytest

import openapi
from openapi import SDK

from .common_helpers import record_test

PRELUDE = """\
from typing import Any

from openapi import SDK
from openapi.models.shared import (
    BodyVariantModelTypedDict,
    DisjointFieldTypesIntegerTypedDict,
    DisjointFieldTypesStringTypedDict,
    StreamingBodyVariantModelTypedDict,
    SubsetVariantBaseTypedDict,
)

s = SDK()
"""

INVALID_CALLS = {
    "unknown_flattened_kwarg": """
s.unions.body_variant_overloads_discriminated(
    "id", kind="model", model="m", input="i", unknown_field=1
)
""",
    "nested_typed_dict_typo": """
s.unions.body_variant_overloads_discriminated(
    "id", kind="model", model="m", input="i", generation_config={"temprature": 0.5}
)
""",
    "request_typed_dict_typo": """
s.unions.body_variant_overloads_discriminated(
    "id",
    request={
        "variant_id": "id",
        "discriminated_variant_body": {
            "kind": "model",
            "model": "m",
            "input": "i",
            "generation_config": {"temprature": 0.5},
        },
    },
)
""",
    "mixed_variant_fields": """
s.unions.body_variant_overloads_discriminated(
    "id", kind="model", model="m", agent="a", input="i"
)
""",
    "unknown_flattened_kwarg_with_stream": """
s.unions.body_variant_overloads_sse_with_path(
    "id", kind="model", model="m", input="i", stream=True, unknown_field=1
)
""",
    "nested_typed_dict_typo_with_bool_stream": """
def call(stream: bool) -> None:
    s.unions.body_variant_overloads_sse_with_path(
        "id",
        kind="model",
        model="m",
        input="i",
        stream=stream,
        generation_config={"temprature": 0.5},
    )
""",
}

INVALID_CALL_ERRORS = {
    "unknown_flattened_kwarg": 'No overloads for "body_variant_overloads_discriminated"',
    "nested_typed_dict_typo": '"temprature" is an undefined item',
    "request_typed_dict_typo": 'parameter "request"',
    "mixed_variant_fields": 'No overloads for "body_variant_overloads_discriminated"',
    "unknown_flattened_kwarg_with_stream": 'No overloads for "body_variant_overloads_sse_with_path"',
    "nested_typed_dict_typo_with_bool_stream": '"temprature" is an undefined item',
}

VALID_CALLS = {
    "flattened_kwargs": """
s.unions.body_variant_overloads_discriminated(
    "id", kind="model", model="m", input="i", generation_config={"temperature": 0.5}
)
""",
    "stream_true": """
s.unions.body_variant_overloads_sse_with_path(
    "id", kind="model", model="m", input="i", stream=True
)
""",
    "bool_stream": """
def call(stream: bool) -> None:
    s.unions.body_variant_overloads_sse_with_path(
        "id", kind="model", model="m", input="i", stream=stream
    )
""",
    "typed_dict_spread": """
params: BodyVariantModelTypedDict = {"kind": "model", "model": "m", "input": "i"}
s.unions.body_variant_overloads_discriminated("id", **params)
""",
    "typed_dict_spread_with_stream": """
def call(stream: bool) -> None:
    params: StreamingBodyVariantModelTypedDict = {
        "kind": "model",
        "model": "m",
        "input": "i",
        "stream": stream,
    }
    s.unions.body_variant_overloads_sse_with_path("id", **params)
""",
    "none_for_optional_fields": """
s.unions.body_variant_overloads_discriminated(
    "id", kind="model", model="m", input="i", background=None, generation_config=None
)
""",
    "subset_variant_typed_dict_spread": """
params: SubsetVariantBaseTypedDict = {"name": "n", "type": "t"}
s.unions.body_variant_overloads_subset_variants(**params)
""",
    "untyped_dict_spread": """
body: dict[str, Any] = {"kind": "model", "model": "m", "input": "i"}
s.unions.body_variant_overloads_discriminated("id", **body)
""",
    "request_with_non_body_kwarg": """
s.unions.body_variant_overloads_discriminated(
    "id",
    request={
        "variant_id": "id",
        "discriminated_variant_body": {"kind": "model", "model": "m", "input": "i"},
    },
    x_trace_id="trace",
)
""",
    "request_with_nullable_non_body_kwarg": """
s.unions.body_variant_overloads_discriminated(
    "id",
    request={
        "variant_id": "id",
        "discriminated_variant_body": {"kind": "model", "model": "m", "input": "i"},
    },
    cursor=None,
)
""",
    "disjoint_field_types_typed_dict_spread": """
text: DisjointFieldTypesStringTypedDict = {"name": "n", "type": "t", "value": "v"}
number: DisjointFieldTypesIntegerTypedDict = {"name": "n", "type": "t", "value": 1}
s.unions.body_variant_overloads_disjoint_field_types(**text)
s.unions.body_variant_overloads_disjoint_field_types(**number)
""",
    "single_value_enum_discriminator": """
s.unions.body_variant_overloads_single_value_enum_discriminator(
    kind="a", alpha="x"
)
s.unions.body_variant_overloads_single_value_enum_discriminator(
    kind="b", beta="y"
)
""",
    "request_satisfies_required_query_param": """
s.unions.body_variant_overloads_renamed_params(
    "id",
    request={
        "item_id": "id",
        "count": 1,
        "discriminated_variant_body": {"kind": "model", "model": "m", "input": "i"},
    },
)
""",
}


def run_pyright(tmp_path: Path, cases: dict[str, str]) -> dict[str, list[str]]:
    for name, call in cases.items():
        (tmp_path / f"{name}.py").write_text(PRELUDE + call)

    src_dir = Path(openapi.__file__).resolve().parent.parent
    (tmp_path / "pyrightconfig.json").write_text(
        json.dumps({"extraPaths": [str(src_dir)]})
    )

    result = subprocess.run(
        [
            sys.executable,
            "-m",
            "pyright",
            "--outputjson",
            "--pythonpath",
            sys.executable,
            "--project",
            str(tmp_path),
        ],
        cwd=tmp_path,
        capture_output=True,
        text=True,
        check=False,
    )
    try:
        report = json.loads(result.stdout)
    except json.JSONDecodeError:
        pytest.fail(f"pyright did not report JSON:\n{result.stdout}\n{result.stderr}")

    errors: dict[str, list[str]] = {name: [] for name in cases}
    for diagnostic in report["generalDiagnostics"]:
        if diagnostic["severity"] == "error":
            errors.setdefault(Path(diagnostic["file"]).stem, []).append(
                diagnostic["message"]
            )
    return errors


def test_body_variant_overloads_reject_invalid_calls(tmp_path: Path):
    errors = run_pyright(tmp_path, INVALID_CALLS)

    assert {
        name: messages
        for name, messages in errors.items()
        if not any(INVALID_CALL_ERRORS[name] in message for message in messages)
    } == {}


def test_body_variant_overloads_accept_valid_calls(tmp_path: Path):
    errors = run_pyright(tmp_path, VALID_CALLS)

    assert {name: messages for name, messages in errors.items() if messages} == {}


def overload_headers(method_name: str, sub_sdk: object) -> list[str]:
    lines = inspect.getsource(type(sub_sdk)).splitlines()
    starts = (f"def {method_name}(", f"def {method_name}_async(")
    headers = []
    for index, line in enumerate(lines):
        if not any(start in line for start in starts):
            continue
        end = index
        while not lines[end].lstrip().startswith(")"):
            end += 1
        headers.append("\n".join(lines[index : end + 1]))
    return headers


def unpack_overload_ignores(method_name: str) -> dict[str, list[bool]]:
    headers = overload_headers(method_name, SDK().unions)
    return {
        name: [
            "# pyright: ignore[reportOverlappingOverload]" in header.splitlines()[0]
            for header in headers
            if f"def {name}(" in header and "Unpack[" in header
        ]
        for name in (method_name, f"{method_name}_async")
    }


@pytest.mark.parametrize(
    "method_name",
    [
        "body_variant_overloads_discriminated",
        "body_variant_overloads_subset_variants",
        "body_variant_overloads_disjoint_field_types",
    ],
)
def test_body_variant_overloads_ignore_overlap_after_first_unpack(method_name: str):
    ignores = unpack_overload_ignores(method_name)

    assert ignores == {name: [False, True] for name in ignores}


def test_body_variant_overloads_skip_unpack_for_single_value_enum_discriminators():
    headers = overload_headers(
        "body_variant_overloads_single_value_enum_discriminator", SDK().unions
    )

    assert headers
    assert not [header for header in headers if "Unpack[" in header]


def mock_sdk(sent: list[httpx.Request]) -> SDK:
    def transport(request: httpx.Request) -> httpx.Response:
        sent.append(request)
        return httpx.Response(200, json={"ok": True})

    return SDK(
        server_url="http://localhost",
        client=httpx.Client(transport=httpx.MockTransport(transport)),
    )


def send_discriminated_with_request(**kwargs) -> httpx.Request:
    sent: list[httpx.Request] = []
    mock_sdk(sent).unions.body_variant_overloads_discriminated(
        "from-argument",
        request={
            "variant_id": "from-request",
            "x_trace_id": "from-request",
            "discriminated_variant_body": {"kind": "model", "model": "m", "input": "i"},
        },
        **kwargs,
    )
    return sent[0]


def test_body_variant_overloads_apply_non_body_arguments_to_request():
    record_test("unions-body-variant-overloads-non-body-arguments")
    sent = send_discriminated_with_request(
        x_trace_id="from-argument", include_meta=True, tags=("a", "b")
    )

    assert sent.url.path.endswith("/discriminated/from-argument")
    assert sent.headers["X-Trace-ID"] == "from-argument"
    assert sent.url.params["include_meta"] == "true"
    assert sent.url.params.get_list("tags") == ["a", "b"]


def test_body_variant_overloads_reject_body_fields_with_request():
    with pytest.raises(ValueError, match="'request' and body fields"):
        send_discriminated_with_request(background=True)


def test_body_variant_overloads_keep_request_values_for_none_arguments():
    sent = send_discriminated_with_request(x_trace_id=None)

    assert sent.headers["X-Trace-ID"] == "from-request"


def test_body_variant_overloads_apply_renamed_arguments_to_request():
    sent: list[httpx.Request] = []
    mock_sdk(sent).unions.body_variant_overloads_renamed_params(
        "from-argument",
        request={
            "item_id": "from-request",
            "count": 1,
            "search": "from-request",
            "request_ref": "from-request",
            "discriminated_variant_body": {"kind": "model", "model": "m", "input": "i"},
        },
        count=2,
        search="from-argument",
        request_ref="from-argument",
    )

    assert sent[0].url.path.endswith("/renamed-params/from-argument")
    assert sent[0].url.params["count"] == "2"
    assert sent[0].url.params["q"] == "from-argument"
    assert sent[0].headers["X-Request-Ref"] == "from-argument"


def test_body_variant_overloads_send_renamed_flattened_arguments():
    sent: list[httpx.Request] = []
    mock_sdk(sent).unions.body_variant_overloads_renamed_params(
        "id",
        count=1,
        search="term",
        request_ref="ref",
        kind="model",
        model="m",
        input="i",
    )

    assert sent[0].url.params["q"] == "term"
    assert sent[0].headers["X-Request-Ref"] == "ref"


def test_body_variant_overloads_send_each_discriminated_variant():
    record_test("unions-body-variant-overloads-discriminated")
    sent: list[httpx.Request] = []
    sdk = mock_sdk(sent)

    sdk.unions.body_variant_overloads_discriminated(
        "id", kind="model", model="m", input="i"
    )
    sdk.unions.body_variant_overloads_discriminated(
        "id", kind="agent", agent="a", input="i"
    )

    assert [json.loads(request.content) for request in sent] == [
        {"kind": "model", "model": "m", "input": "i"},
        {"kind": "agent", "agent": "a", "input": "i"},
    ]


def test_body_variant_overloads_send_each_non_discriminated_variant():
    record_test("unions-body-variant-overloads-non-discriminated")
    sent: list[httpx.Request] = []
    sdk = mock_sdk(sent)

    sdk.unions.body_variant_overloads_non_discriminated(alpha="a", shared=1)
    sdk.unions.body_variant_overloads_non_discriminated(gamma="g", c_only=True)

    assert [json.loads(request.content) for request in sent] == [
        {"alpha": "a", "shared": 1},
        {"gamma": "g", "c_only": True},
    ]


def test_body_variant_overloads_match_request_body():
    record_test("unions-body-variant-overloads-request-body")
    sent: list[httpx.Request] = []
    sdk = mock_sdk(sent)

    sdk.unions.body_variant_overloads_discriminated(
        "id", kind="model", model="m", input="i", type_="t", background=True
    )
    sdk.unions.body_variant_overloads_discriminated(
        "id",
        request={
            "variant_id": "id",
            "discriminated_variant_body": {
                "kind": "model",
                "model": "m",
                "input": "i",
                "type": "t",
                "background": True,
            },
        },
    )

    assert json.loads(sent[0].content) == json.loads(sent[1].content)
    assert json.loads(sent[0].content)["type"] == "t"


def test_body_variant_overloads_send_renamed_body_arguments():
    record_test("unions-body-variant-overloads-renamed-argument")
    sent: list[httpx.Request] = []
    sdk = mock_sdk(sent)

    sdk.unions.body_variant_overloads_reserved_field_names(
        alpha="a", retries_=3, timeout_ms_=5
    )
    sdk.unions.body_variant_overloads_reserved_field_names(beta="b", server_url_="u")

    assert [json.loads(request.content) for request in sent] == [
        {"alpha": "a", "retries": 3, "timeout_ms": 5},
        {"beta": "b", "server_url": "u"},
    ]


def test_body_variant_overloads_send_body_field_named_request():
    sent: list[httpx.Request] = []
    mock_sdk(sent).unions.body_variant_overloads_reserved_field_names(
        beta="b", request_="r"
    )

    assert json.loads(sent[0].content) == {"beta": "b", "request": "r"}


def test_body_variant_overloads_document_renamed_body_field():
    headers = overload_headers(
        "body_variant_overloads_reserved_field_names", SDK().unions
    )
    docs = inspect.getdoc(SDK().unions.body_variant_overloads_reserved_field_names)

    assert any("request_: Optional[str] = ..." in header for header in headers)
    assert ":param request_:" in docs
    assert "The request object to send." not in docs
