import sys
import types
from unittest.mock import patch

import pytest

import openapi.sdk as sdk_module
from openapi import SDK
from openapi.utils import dynamic_imports


def test_subsdk_map_uses_relative_module_paths():
    """Guards against lazy sub-SDK imports regressing to absolute module paths."""
    assert SDK._sub_sdk_map["auth"][0] == ".auth"


def test_subsdk_retry_cleanup_removes_resolved_relative_module():
    """Guards against sub-SDK retry cleanup popping '.module' instead of the resolved key."""
    module_key = "openapi.auth"
    previous_module = sys.modules.get(module_key)
    sys.modules[module_key] = types.ModuleType(module_key)

    def raise_key_error(*args, **kwargs):
        raise KeyError("half-initialized module")

    try:
        with SDK() as sdk:
            with patch.object(sdk_module.importlib, "import_module", raise_key_error):
                with pytest.raises(KeyError, match="Failed to import module '.auth'"):
                    sdk.dynamic_import(".auth", retries=1)

        assert module_key not in sys.modules
    finally:
        if previous_module is not None:
            sys.modules[module_key] = previous_module


def test_package_lazy_retry_cleanup_removes_resolved_relative_module():
    """Guards against package lazy retry cleanup popping '.module' instead of the resolved key."""
    module_key = "openapi.models.shared"
    previous_module = sys.modules.get(module_key)
    sys.modules[module_key] = types.ModuleType(module_key)

    def raise_key_error(*args, **kwargs):
        raise KeyError("half-initialized module")

    try:
        with patch.object(dynamic_imports, "import_module", raise_key_error):
            with pytest.raises(KeyError, match="Failed to import module '.shared'"):
                dynamic_imports.dynamic_import("openapi.models", ".shared", retries=1)

        assert module_key not in sys.modules
    finally:
        if previous_module is not None:
            sys.modules[module_key] = previous_module
