# SecureCollab Architecture

This document summarizes the implementation architecture used by this repository.

## Principles
- Zero-knowledge by design: server stores ciphertext and metadata only.
- Simplicity-first implementation.
- Strong observability from day one.
- Test-backed changes only.

## Repository Boundaries
- `docs/`: documentation and ADRs.
- `services/`: Go backend services.
- `client/`: Tauri client.
- `pipeline/`: CDC and analytics pipeline assets.
- `deploy/`: Docker Compose and Helm deployment assets.
- `infra/`: Terraform and provisioning assets.
- `tests/`: integration and load tests.
- `db/`: schema migrations.
- `proto/`: protobuf contracts.
