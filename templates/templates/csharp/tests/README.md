# C# test templates

Tests here run for both `csharp` (template version v1) and `csharpv2` (template version v2).
`csharpv2` reads this directory through its template overlay.

- `Common/`: helpers and project file copied into every test group.
- `<group>/`: tests for one test group (`primary`, `quaternary`, ...).

## Scope

- **v1 + v2 (default)**: `<Name>.cs`. Write new tests here. Where v1 and v2 differ
  (serializer API, exception type, wire format), call a `Compat.*` member instead of
  branching on the template version. v1 members live in `<group>/Compat.V1.cs.stmpl`,
  v2 members in `../../csharpv2/tests/Common/Compat.cs.stmpl`; both expose the same signatures.
- **v1 only**: `<Name>.V1.cs.stmpl`, wrapped in a whole-file gate with a one-line reason:

  ```
  {{- /* v1-only: <reason> */ -}}
  {{- if not isCsharpV2 -}}
  ...
  {{ end -}}
  ```

  Use only when v2 cannot express the test (v1-only option or internal API). To add
  v1-only tests to an existing test class, declare it `partial` in both files.
- **v2 only**: `../../csharpv2/tests/<group>/<Name>.cs`. Use a name that does not exist
  here; a file with the same path replaces the shared one for v2.
