# Repository Guidelines

## Project Structure & Module Organization
This Go service exposes the OpenHack API via Fiber. The entrypoint lives in `cmd/server/main.go`, while domain logic resides under `internal/` (e.g., `accounts`, `teams`, `events`, `env`, `db`). Shared helpers stay in `internal/utils` and configuration defaults in `internal/env`. Integration and unit specs are collected under `test/...` mirroring the package layout. API collateral such as `contract.yaml` and `models.json` plus helper scripts (`run_*.sh`) sit at repo root.

## Build, Test, and Development Commands
Run the API locally with `./run_dev.sh` (dev profile) or point to production-style settings using `./run_prod.sh`. Use `./run_test.sh` when exercising the server against the test profile. Build a binary with `go build ./cmd/server` and run the full suite with `go test ./test/... -count=1`. For event counter verification and Redis priming, use `./runtests.sh`, which seeds DB 2 before executing tests.

## Environment & Services
Create a `.env` alongside `VERSION`; required keys include `PORT`, `MONGO_URI`, `JWT_SECRET`, `BADGE_PILES`, and optional `PREFORK`. The app expects reachable MongoDB and Redis instances; Redis database is auto-selected per profile (prod=0, dev=1, test=2). Tests run faster when Mongo has fixture data for accounts, teams, flags, and events collections.

## Coding Style & Naming Conventions
Format Go files with `gofmt` (tabs for indentation) and keep imports `goimports`-sorted. Packages and directories are lower_snake to align with module names, while exported identifiers use PascalCase and internal helpers stay camelCase. Fiber routes follow resource-oriented paths (`/accounts`, `/teams`); keep handler filenames plural and align tests in `test/<package>`.

## Testing Guidelines
Write `_test.go` companions under `test/` using Go's `testing` plus `testify/assert`. Name functions `Test<Behavior>` and favor table-driven cases for data variations. Ensure Redis and Mongo are running before tests; `runtests.sh` resets the Redis event counter and prints totals for verification. Keep new tests idempotent and avoid relying on production datasets.

## Commit & Pull Request Guidelines
Adopt the existing `type: summary` format (`feat:`, `fix:`, `docs:`) shown in `git log`. Commits should be scoped to a single concern, run `gofmt`/tests beforehand, and mention configuration changes. PRs need a short problem statement, linked issue or ticket, testing evidence (`go test` output, screenshots for new endpoints), and call out any data migrations or new environment variables.
