import inspect
from pathlib import Path

from openapi.parameters import Parameters
from openapi.requestbodies import RequestBodies


def test_required_path_parameters_are_positional():
    signature = inspect.signature(Parameters.mixed_parameters_camel_case)

    assert (
        signature.parameters["path_param"].kind
        is inspect.Parameter.POSITIONAL_OR_KEYWORD
    )
    assert signature.parameters["header_param"].kind is inspect.Parameter.KEYWORD_ONLY
    assert (
        signature.parameters["query_string_param"].kind
        is inspect.Parameter.KEYWORD_ONLY
    )


def test_required_path_parameters_are_positional_for_async_methods():
    signature = inspect.signature(Parameters.mixed_parameters_camel_case_async)

    assert (
        signature.parameters["path_param"].kind
        is inspect.Parameter.POSITIONAL_OR_KEYWORD
    )
    assert signature.parameters["header_param"].kind is inspect.Parameter.KEYWORD_ONLY
    assert (
        signature.parameters["query_string_param"].kind
        is inspect.Parameter.KEYWORD_ONLY
    )


def test_non_path_parameters_remain_keyword_only():
    signature = inspect.signature(Parameters.allow_empty_value_query_params)

    assert signature.parameters["str_param"].kind is inspect.Parameter.KEYWORD_ONLY


def test_body_variant_overload_implementation_keeps_path_parameters_positional():
    signature = inspect.signature(RequestBodies.body_variant_overloads_sse_with_path)

    assert signature.parameters["id"].kind is inspect.Parameter.POSITIONAL_OR_KEYWORD
    assert signature.parameters["request"].kind is inspect.Parameter.KEYWORD_ONLY
    assert signature.parameters["body_kwargs"].kind is inspect.Parameter.VAR_KEYWORD


def test_body_variant_overloads_compose_with_sse_and_positional_path_params():
    source = Path(inspect.getsourcefile(RequestBodies)).read_text()

    assert "def body_variant_overloads_sse_with_path(\n        self,\n        id: str,\n        *," in source
    assert "stream: Union[Literal[False], None] = None" in source
    assert "stream: Literal[True]" in source
    assert ") -> models.operations.BodyVariantOverloadsSSEWithPathResponse:" in source
    assert '_request_kwargs["StreamingVariantBody"] = _body_kwargs' in source
    assert "request.streaming_variant_body" in source
    assert 'getattr(\n                getattr(request, "streaming_variant_body", None), "stream", False\n            )\n            is True' in source
