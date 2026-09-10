package snapshots

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

// TestSnapPyModelValidatorInjection covers the
// .speakeasy/addons/model_validator/<modelPath>/<file>.py splice path.
//
// Imports are parsed into (from, symbol) pairs and routed through
// “addImport“ so they dedup against generator-emitted imports. The
// body is spliced into the owning class body at 4-space indent.
// Bare “import X as Y“ is rejected (see model-validator.ts).
//
//   - UserProfile:       happy path, single “from X import Y“.
//   - Lifecycle:         classmethod helper + chained @model_validator.
//   - OrderRecord:       bare “import x“ + decorator using dotted name.
//   - DottedImport:      bare “import x.y“ (dotted module path).
//   - MultiLineImport:   paren-spanning “from X import (a, b)“.
//   - BackslashImport:   trailing-backslash continuation imports.
//   - AliasImporter:     “from X import Y as Z“ with body using Z.
//   - RelativeImporter:  relative “from . import x“ + inline comments.
//   - BareAsRejected:    “import X as Y“ warns and skips splice.
//   - NoCompanion:       control — no validator file present.
func TestSnapPyModelValidatorInjection(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Model Validator Cases
  version: 1.0.0
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserProfile'
  /lifecycle:
    get:
      operationId: getLifecycle
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Lifecycle'
  /orders:
    get:
      operationId: getOrder
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OrderRecord'
  /dotted:
    get:
      operationId: getDotted
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DottedImport'
  /multiline:
    get:
      operationId: getMultiline
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MultiLineImport'
  /backslash:
    get:
      operationId: getBackslash
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/BackslashImport'
  /alias:
    get:
      operationId: getAlias
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AliasImporter'
  /relative:
    get:
      operationId: getRelative
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/RelativeImporter'
  /bareas:
    get:
      operationId: getBareAs
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/BareAsRejected'
  /control:
    get:
      operationId: getControl
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/NoCompanion'
components:
  schemas:
    UserProfile:
      type: object
      required:
        - firstName
        - lastName
      properties:
        firstName:
          type: string
        lastName:
          type: string
    Lifecycle:
      type: object
      required:
        - id
        - status
      properties:
        id:
          type: string
        status:
          type: string
    OrderRecord:
      type: object
      required:
        - order_id
      properties:
        order_id:
          type: string
    DottedImport:
      type: object
      required:
        - path
      properties:
        path:
          type: string
    MultiLineImport:
      type: object
      required:
        - value
      properties:
        value:
          type: string
    BackslashImport:
      type: object
      required:
        - token
      properties:
        token:
          type: string
    AliasImporter:
      type: object
      required:
        - name
      properties:
        name:
          type: string
    RelativeImporter:
      type: object
      required:
        - label
      properties:
        label:
          type: string
    BareAsRejected:
      type: object
      required:
        - marker
      properties:
        marker:
          type: string
    NoCompanion:
      type: object
      required:
        - tag
      properties:
        tag:
          type: string`

	genYaml := `python:
  packageName: userapi
`

	happyValidator := `from pydantic import model_validator

@model_validator(mode="after")
def check_names(self):
    if self.first_name == self.last_name:
        raise ValueError("first_name and last_name must differ")
    return self
`

	chainValidator := `from typing import Any
from pydantic import model_validator

@classmethod
def _normalize_status(cls, data: Any) -> Any:
    if isinstance(data, dict) and isinstance(data.get("status"), str):
        data["status"] = data["status"].lower()
    return data

@model_validator(mode="before")
@classmethod
def _coerce_status(cls, data: Any) -> Any:
    return cls._normalize_status(data)

@model_validator(mode="after")
def _trim_id(self):
    self.id = self.id.strip()
    return self
`

	// Bare ``import x`` form. Decorator references it via dotted name.
	bareImportValidator := `import pydantic

@pydantic.model_validator(mode="after")
def reject_blank(self):
    if not self.order_id.strip():
        raise ValueError("order_id must not be blank")
    return self
`

	// Bare ``import x.y`` dotted module path. Use a Pylint-compatible standard
	// library module so this parser fixture remains valid under Python 3.14.
	dottedImportValidator := `import os.path
from pydantic import model_validator

@model_validator(mode="after")
def reject_non_string(self):
    if not os.path.basename(self.path):
        raise ValueError("path must not be empty")
    return self
`

	// Paren-spanning ``from X import (a, b)``.
	multiLineValidator := `from typing import (
    Any,
    Optional,
)
from pydantic import model_validator

@model_validator(mode="after")
def normalize_value(self):
    _: Optional[Any] = None
    self.value = self.value.strip()
    return self
`

	// Trailing-backslash continuation across an import.
	backslashValidator := `from typing import \
    Any
from pydantic import model_validator, \
    ValidationError

@model_validator(mode="after")
def reject_empty(self):
    if not self.token:
        raise ValidationError.from_exception_data("empty token", [])
    _: Any = self.token
    return self
`

	// ``from X import Y as Z`` alias.
	aliasValidator := `from pydantic import model_validator as mv

@mv(mode="after")
def lowercase_name(self):
    self.name = self.name.lower()
    return self
`

	// Bare ``import X as Y`` is rejected by the parser — splice is
	// skipped and the file generates without addon content.
	bareAsRejectedValidator := `import pydantic as pd

@pd.model_validator(mode="after")
def will_not_be_spliced(self):
    return self
`

	// Relative import + trailing comments on the import lines.
	relativeValidator := `from . import nocompanion  # type: ignore
from pydantic import model_validator  # noqa: F401

@model_validator(mode="after")
def normalize_label(self):
    _ = nocompanion  # keep import live for the splice
    self.label = self.label.strip().lower()
    return self
`

	expectedSnapshotFiles := []string{
		"src/userapi/models/userprofile.py",
		"src/userapi/models/lifecycle.py",
		"src/userapi/models/orderrecord.py",
		"src/userapi/models/dottedimport.py",
		"src/userapi/models/multilineimport.py",
		"src/userapi/models/backslashimport.py",
		"src/userapi/models/aliasimporter.py",
		"src/userapi/models/relativeimporter.py",
		"src/userapi/models/bareasrejected.py",
		"src/userapi/models/nocompanion.py",
	}

	expectedSnapshot := `--- src/userapi/models/aliasimporter.py ---
"""Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT."""
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

from __future__ import annotations
from pydantic import model_validator as mv
from typing_extensions import TypedDict
from userapi.types import BaseModel


class AliasImporterTypedDict(TypedDict):
    name: str


class AliasImporter(BaseModel):
    name: str

    @mv(mode="after")
    def lowercase_name(self):
        self.name = self.name.lower()
        return self


--- src/userapi/models/backslashimport.py ---
"""Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT."""
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

from __future__ import annotations
from pydantic import ValidationError, model_validator
from typing import Any
from typing_extensions import TypedDict
from userapi.types import BaseModel


class BackslashImportTypedDict(TypedDict):
    token: str


class BackslashImport(BaseModel):
    token: str

    @model_validator(mode="after")
    def reject_empty(self):
        if not self.token:
            raise ValidationError.from_exception_data("empty token", [])
        _: Any = self.token
        return self


--- src/userapi/models/bareasrejected.py ---
"""Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT."""
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

from __future__ import annotations
from typing_extensions import TypedDict
from userapi.types import BaseModel


class BareAsRejectedTypedDict(TypedDict):
    marker: str


class BareAsRejected(BaseModel):
    marker: str


--- src/userapi/models/dottedimport.py ---
"""Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT."""
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

from __future__ import annotations
import os.path
from pydantic import model_validator
from typing_extensions import TypedDict
from userapi.types import BaseModel


class DottedImportTypedDict(TypedDict):
    path: str


class DottedImport(BaseModel):
    path: str

    @model_validator(mode="after")
    def reject_non_string(self):
        if not os.path.basename(self.path):
            raise ValueError("path must not be empty")
        return self


--- src/userapi/models/lifecycle.py ---
"""Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT."""
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

from __future__ import annotations
from pydantic import model_validator
from typing import Any
from typing_extensions import TypedDict
from userapi.types import BaseModel


class LifecycleTypedDict(TypedDict):
    id: str
    status: str


class Lifecycle(BaseModel):
    id: str

    status: str

    @classmethod
    def _normalize_status(cls, data: Any) -> Any:
        if isinstance(data, dict) and isinstance(data.get("status"), str):
            data["status"] = data["status"].lower()
        return data

    @model_validator(mode="before")
    @classmethod
    def _coerce_status(cls, data: Any) -> Any:
        return cls._normalize_status(data)

    @model_validator(mode="after")
    def _trim_id(self):
        self.id = self.id.strip()
        return self


--- src/userapi/models/multilineimport.py ---
"""Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT."""
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

from __future__ import annotations
from pydantic import model_validator
from typing import Any, Optional
from typing_extensions import TypedDict
from userapi.types import BaseModel


class MultiLineImportTypedDict(TypedDict):
    value: str


class MultiLineImport(BaseModel):
    value: str

    @model_validator(mode="after")
    def normalize_value(self):
        _: Optional[Any] = None
        self.value = self.value.strip()
        return self


--- src/userapi/models/nocompanion.py ---
"""Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT."""
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

from __future__ import annotations
from typing_extensions import TypedDict
from userapi.types import BaseModel


class NoCompanionTypedDict(TypedDict):
    tag: str


class NoCompanion(BaseModel):
    tag: str


--- src/userapi/models/orderrecord.py ---
"""Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT."""
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

from __future__ import annotations
import pydantic
from typing_extensions import TypedDict
from userapi.types import BaseModel


class OrderRecordTypedDict(TypedDict):
    order_id: str


class OrderRecord(BaseModel):
    order_id: str

    @pydantic.model_validator(mode="after")
    def reject_blank(self):
        if not self.order_id.strip():
            raise ValueError("order_id must not be blank")
        return self


--- src/userapi/models/relativeimporter.py ---
"""Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT."""
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

from __future__ import annotations
from . import nocompanion
from pydantic import model_validator
from typing_extensions import TypedDict
from userapi.types import BaseModel


class RelativeImporterTypedDict(TypedDict):
    label: str


class RelativeImporter(BaseModel):
    label: str

    @model_validator(mode="after")
    def normalize_label(self):
        _ = nocompanion  # keep import live for the splice
        self.label = self.label.strip().lower()
        return self


--- src/userapi/models/userprofile.py ---
"""Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT."""
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

from __future__ import annotations
import pydantic
from pydantic import model_validator
from typing_extensions import Annotated, TypedDict
from userapi.types import BaseModel


class UserProfileTypedDict(TypedDict):
    first_name: str
    last_name: str


class UserProfile(BaseModel):
    first_name: Annotated[str, pydantic.Field(alias="firstName")]

    last_name: Annotated[str, pydantic.Field(alias="lastName")]

    @model_validator(mode="after")
    def check_names(self):
        if self.first_name == self.last_name:
            raise ValueError("first_name and last_name must differ")
        return self


try:
    UserProfile.model_rebuild()
except NameError:
    pass


` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
		AfterGenerate: func(t *testing.T, tempDir string) {
			t.Helper()
			writeCompanion := func(name, body string) {
				path := filepath.Join(tempDir, ".speakeasy", "addons", "model_validator", "models", name)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
					t.Fatalf("write %s: %v", name, err)
				}
			}
			writeCompanion("userprofile.py", happyValidator)
			writeCompanion("lifecycle.py", chainValidator)
			writeCompanion("orderrecord.py", bareImportValidator)
			writeCompanion("dottedimport.py", dottedImportValidator)
			writeCompanion("multilineimport.py", multiLineValidator)
			writeCompanion("backslashimport.py", backslashValidator)
			writeCompanion("aliasimporter.py", aliasValidator)
			writeCompanion("relativeimporter.py", relativeValidator)
			writeCompanion("bareasrejected.py", bareAsRejectedValidator)
			// Intentionally omit nocompanion.py — control case.
		},
	})
}
