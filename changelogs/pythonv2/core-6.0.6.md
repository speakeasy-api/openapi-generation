## core: 6.0.6 - 2026-02-14
### :bug: Bug Fixes
- resolve namespace import collisions by using root package import strategy, gated behind `fixFlags.conflictResistantModelImportsFeb2026` in gen.yaml *(commit by [@mfbx9da4](https://github.com/mfbx9da4))*

### :recycle: Refactors
- extract duplicated `dynamic_import`, `__getattr__`, and `__dir__` helpers from `__init__.py` templates into a shared `utils/dynamic_imports.py` module, reducing code duplication across generated package init files
