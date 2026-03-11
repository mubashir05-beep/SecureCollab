---
applyTo: "services/**"
description: "Use when implementing backend services, boundaries, and architecture decisions for SecureCollab."
---

# Architecture Rules

- Keep service responsibilities narrow and explicit.
- Preserve zero-knowledge constraints: no plaintext message content on server paths.
- Avoid deep abstraction layers unless duplication proves the need.
- Prefer composition and small modules over large shared utility packages.
- Record meaningful architectural trade-offs in `docs/adr/`.
