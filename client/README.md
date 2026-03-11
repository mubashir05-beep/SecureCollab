# Client Phase 2 Bootstrap

This folder is the Phase 2 starting point for the Tauri desktop application.

## Goal
Create an initial Tauri shell with authentication screen, workspace sidebar, and channel view skeleton.

## Implemented
- Rust core crate scaffold at `client/src-tauri`.
- Identity keypair generation module (X25519) in `client/src-tauri/src/lib.rs`.
- Unit tests for key generation behavior (`cargo test`).
- Svelte + Tailwind UI shell scaffold in `client/ui`.
- Reusable UI components (`Button`, `Panel`) for consistency.
- Frontend smoke test via Vitest + Testing Library.

## UI Stack Decision
- Svelte + Tailwind selected for clean, fast UI iteration.
- Ant Design is React-focused, so we use a Svelte-native component system and design tokens for consistency.

## Planned Bootstrap Commands
```bash
cd client
cargo tauri init
cargo tauri dev
```

## Run Current Client Tests
```bash
task test:client
task ui:test
```

## Current Status
- Folder initialized for Phase 2 kickoff.
- Phase 2 crypto foundation started with tested identity key generation.
- Backend services from Phase 1 are available for integration.
