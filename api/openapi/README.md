# OpenAPI Source

This directory stores the OpenAPI source contract used as the public endpoint reference.

## Main Files

- `api/openapi/openapi.yaml`: canonical source contract.
- `docs/generated/openapi.yaml`: generated artifact that must remain synchronized.

## Standard Commands

- `make openapi-generate`: validate source spec and synchronize generated artifact.
- `make openapi-check`: validate source+generated spec and check runtime route drift.
- `make openapi-autogen-watch`: watch source/runtime contract files and auto-regenerate artifacts.
- `make openapi-autogen-install-hook`: install pre-commit hook that auto-regenerates artifacts.
