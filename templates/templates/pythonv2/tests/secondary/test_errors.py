import pytest
from openapi import SDK
from openapi.models import errors, operations
from .common_helpers import HTTPBIN_URL


def test_discriminated_union_of_errors_post():
    # Tests `utils.get_discriminator` for `enumFormat: enum`

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    req1 = operations.TaggedError1RequestBody(tag="tag1", error="Error1")
    with pytest.raises(
        errors.ErrorUnionDiscriminatedPostResponseBody,
        match='{"error":"Error1","tag":"tag1"}',
    ) as error1:
        s.errors.error_union_discriminated_post(request=req1)
    assert error1.value.data.error == "Error1"

    req2 = operations.TaggedError2RequestBody(
        tag="tag2",
        tagged_error2_message=operations.TaggedError2Message(message="Error2"),
    )
    with pytest.raises(
        errors.ErrorUnionDiscriminatedPostResponseBody,
        match='{"error":{"message":"Error2"},"tag":"tag2"}',
    ) as error2:
        s.errors.error_union_discriminated_post(request=req2)
    assert isinstance(error2.value.data.error, errors.SchemasTaggedError2Error)
    assert error2.value.data.error.message == "Error2"


def test_base_and_default_error_name_resolution():
    """ Tests that `fixBuiltInErrorNameConflicts()` properly renamed the Base and Default errors
    since they have both been set to an already existing error name 'Error' (see `gen.yaml`).
    """
    assert issubclass(errors.DefaultError, errors.DefaultBaseError)
