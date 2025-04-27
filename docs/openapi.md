# OpenAPI Client & Orval Codegen Guide

This document explains how to regenerate the TypeScript API client for the Inbox Whisperer frontend using [Orval](https://orval.dev/) and your OpenAPI spec.

---

## Prerequisites
- Ensure your OpenAPI spec is up to date at `api/openapi.yaml`.
- Orval should be installed as a dev dependency (`npm install --save-dev orval@7.8.0`).
- The Orval config file exists at `orval.config.ts` in the project root.

---

## Regenerating the API Client

Whenever you update your OpenAPI spec or want to refresh the generated client, run:

```sh
npx orval --config ./orval.config.ts
```

This will:
- Generate/refresh the TypeScript client and hooks in `web/src/api/generated/`
- Update schemas in `web/src/api/generated/inboxWhispererAPI.schemas.ts`

---

## Troubleshooting
- If you see errors about the mutator, ensure `web/src/api/useApi.ts` exports a `useApi` function (even if it's a dummy).
- If you see errors about a file being a directory, delete `web/src/api/generated/inboxWhispererAPI.schemas.ts` and try again.
- If you add new endpoints to your OpenAPI spec, re-run the command above to keep your client in sync.

---

## References
- [Orval Documentation](https://orval.dev/docs/)
- [OpenAPI Specification](https://swagger.io/specification/)

---

_Last updated: 2025-04-27_
