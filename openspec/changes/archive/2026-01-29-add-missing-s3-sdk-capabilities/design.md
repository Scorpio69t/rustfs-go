## Context
The current SDK lacks several S3 capabilities that existed in the legacy example set. These gaps impact compatibility for existing users and make it harder to provide complete S3 workflows.

## Goals / Non-Goals
- Goals:
  - Provide presigned HEAD URL generation.
  - Support response header overrides for GetObject (response-content-type, response-content-disposition, etc.).
  - Support listing multipart uploads and upload parts.
  - Add checksum mode support for uploads (and copy when applicable).
  - Support S3 Accelerate endpoints for object operations.
  - Provide client-side encryption (CSE) helpers and usage patterns.
  - Ship examples and documentation for all added features.
- Non-Goals:
  - Changing existing public API behavior (unless required for correctness).
  - Supporting non-S3 storage backends that do not expose these features.

## Decisions
- Decision: Keep new features opt-in via options to avoid breaking existing users.
- Decision: Place CSE helpers under a dedicated package (e.g., pkg/cse) to keep server-side encryption (SSE) APIs clean.
- Decision: Use typed option structs for new request query parameters instead of raw map usage.

## Alternatives considered
- Add raw query/header overrides in every call (rejected: too flexible and unsafe).
- Implement CSE directly in object service with no helper package (rejected: harder to test and reuse).

## Risks / Trade-offs
- CSE requires careful key management and increases CPU usage. Mitigate with clear documentation and test vectors.
- Accelerate requires DNS/endpoint differences and only applies to compatible endpoints.
- Multipart listing APIs add pagination logic and more edge cases.

## Migration Plan
1) Add new options and APIs.
2) Add unit tests and minimal integration examples.
3) Update README and examples index.
4) Validate with lint and go test.

## Open Questions
- Should Accelerate support be limited to upload/download or expanded to presign?
- Should checksum mode be exposed for multipart uploads as well?
