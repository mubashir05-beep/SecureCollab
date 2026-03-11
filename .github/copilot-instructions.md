# SecureCollab Copilot Instructions

Apply these rules across the entire repository.

## Core Engineering Rules
- Keep documentation in `docs/`.
- Keep implementation code in domain folders (`services/`, `client/`, `pipeline/`, `deploy/`, `infra/`, `tests/`, `db/`, `proto/`, `collections/`).
- Prefer simple, explicit designs over complex abstractions.
- Every code change must include tests.
- Keep code easy to read: clear naming, small functions, straightforward control flow.

## Delivery Rules
- Do not mark work complete unless lint and tests pass.
- Bug fixes must include a regression test.
- For non-obvious decisions, add a short ADR under `docs/adr/`.
