from openapi.models import shared


def _has(name: str) -> bool:
    return hasattr(shared, name)


def test_request_body_schema_uses_param_suffix():
    # AuthServiceRequestBody is the body of an authenticated request operation
    # so its TypedDict companion must use the configured request body suffix.
    assert _has("AuthServiceRequestBody")
    assert _has("AuthServiceRequestBodyParam")
    assert not _has("AuthServiceRequestBodyTypedDict")


def test_nested_ref_field_inherits_param_suffix():
    # BasicAuth is reached transitively via AuthServiceRequestBody.basic_auth ($ref).
    assert _has("BasicAuthParam")
    assert not _has("BasicAuthTypedDict")


def test_array_item_ref_inherits_param_suffix():
    # HeaderAuth is reached via AuthServiceRequestBody.header_auth: List[HeaderAuth].
    assert _has("HeaderAuthParam")
    assert not _has("HeaderAuthTypedDict")


def test_map_value_ref_inherits_param_suffix():
    # SimpleObject is reached via DeepObject.map: Dict[str, SimpleObject], and
    # DeepObject is itself reachable from a request body.
    assert _has("SimpleObjectParam")
    assert not _has("SimpleObjectTypedDict")


def test_union_member_inherits_param_suffix():
    # AnyOfMultiMatchMember1/2 are union members of AnyOfMultiMatch, which is
    # used in a request body.
    assert _has("AnyOfMultiMatchMember1Param")
    assert _has("AnyOfMultiMatchMember2Param")
    assert not _has("AnyOfMultiMatchMember1TypedDict")
    assert not _has("AnyOfMultiMatchMember2TypedDict")


def test_response_only_schema_keeps_typeddict_suffix():
    # AllOfToAllOf appears only in response payloads so its TypedDict
    # companion must keep the default TypedDict suffix.
    assert _has("AllOfToAllOf")
    assert _has("AllOfToAllOfTypedDict")
    assert not _has("AllOfToAllOfParam")


def test_param_companion_is_typeddict_subclass():
    cls = shared.AuthServiceRequestBodyParam
    # TypedDict classes carry required/optional key metadata.
    assert hasattr(cls, "__required_keys__") or hasattr(cls, "__optional_keys__")


def test_cascade_collision_falls_back_to_param_model():
    # CollisionBase is body-reachable; its renamed companion would be
    # CollisionBaseParam, but that name is already taken by a sibling
    # body-reachable schema. The cascade fallback picks
    # CollisionBaseParamModel.
    assert _has("CollisionBase")
    assert _has("CollisionBaseParam")  # sibling BaseModel schema
    assert _has("CollisionBaseParamModel")  # CollisionBase's renamed companion
    assert _has("CollisionBaseParamParam")  # CollisionBaseParam's own companion
    assert not _has("CollisionBaseTypedDict")
    assert not _has("CollisionBaseParamTypedDict")


def test_numeric_counter_fallback_when_cascade_exhausted():
    # Every cascade suffix (Param, ParamModel, ParamCompanion,
    # ParamCompanionModel) is occupied by a sibling body-reachable
    # schema. CounterBase's companion must fall through to the numeric
    # counter and emit CounterBaseParam2.
    assert _has("CounterBaseParam2")
    assert not _has("CounterBaseTypedDict")


def test_x_speakeasy_name_override_wins_then_suffix_applies():
    # OriginalNameOverride carries x-speakeasy-name-override:
    # RenamedOverride. The override renames the schema before the
    # input suffix is appended, yielding RenamedOverrideParam (not
    # OriginalNameOverrideParam).
    assert _has("RenamedOverride")
    assert _has("RenamedOverrideParam")
    assert not _has("OriginalNameOverride")
    assert not _has("OriginalNameOverrideParam")
    assert not _has("RenamedOverrideTypedDict")
