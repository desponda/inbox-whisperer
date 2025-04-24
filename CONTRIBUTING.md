# Contributing to Inbox Whisperer

## Database Integration/E2E Tests

**All tests that interact with the database _must_ use the `SetupTestDB` helper from `internal/data/testutil.go`.**

- `SetupTestDB` spins up a dedicated Postgres Testcontainer (not your dev DB) and applies all migrations automatically.
- Do **not** use a static Postgres connection string, `localhost`, or a shared dev/test DB for integration/E2E tests.
- Do **not** run `make dev-up` or similar for test setup. The testcontainer will handle all DB setup/teardown.
- This guarantees isolated, reproducible, and production-like DB environments for every test run.

**Example:**
```go
import "github.com/desponda/inbox-whisperer/internal/data"

func TestSomething(t *testing.T) {
    db, cleanup := data.SetupTestDB(t)
    defer cleanup()
    // ... your test code using db.Pool ...
}
```

## General Test Guidelines
- Prefer table-driven tests and clear test names.
- Clean up any resources you create.
- If you need to test DB connection errors, you may use a static URL, but _never_ for real DB operations.

---

For questions, contact the maintainers or see `internal/data/testutil.go` for details on the test DB setup.
