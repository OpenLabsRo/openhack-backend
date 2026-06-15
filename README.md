# OpenHack Backend API

The backend service that powers OpenHack — a hackathon platform handling
participant accounts, team formation, project submissions, feature flagging,
staff check-in tooling, and crowd-sourced judging.

It is a single Go binary built on [Fiber v3](https://github.com/gofiber/fiber),
backed by MongoDB (primary store) and Redis (cache + event counters).

This repository is designed to be driven by
[**openhack-hypervisor**](https://github.com/openlabsro/openhack-hypervisor),
which clones, tests, builds, and deploys `openhack-backend` instances
automatically.
The lifecycle scripts and version conventions in this repo (see
[Build & Release](#build--release) and [Hypervisor integration](#hypervisor-integration))
exist to support that workflow, but the service also runs standalone for local
development.

## Architecture

The HTTP app is assembled in `internal/app.go` (`SetupApp`), which wires
configuration, the database/cache connections, the event emitter, and mounts the
route groups. The process entrypoint is `cmd/server/main.go`.

Request handling is organized by domain under `internal/`, each grouped under a
top-level route prefix:

| Prefix        | Package                  | Responsibility |
|---------------|--------------------------|----------------|
| `/meta`       | `internal/meta`          | Service-wide health (`/ping`) and version endpoints |
| `/accounts`   | `internal/accounts`      | Participant registration/login, profile, flags, promotionals, vouchers, finalist voting |
| `/teams`      | `internal/teams`         | Team lifecycle, membership, and project submission metadata |
| `/judge`      | `internal/judge`         | Judge auth (token upgrade) and pairwise judging flow |
| `/superusers` | `internal/superusers`    | Admin/staff tooling: feature flags, flag stages, badges, judging setup, participants, staff check-in |

Supporting packages:

- `internal/models` — MongoDB document models and the auth middleware
  (`AccountMiddleware`, `JudgeMiddleware`, `SuperUserMiddlewareBuilder`,
  `FlagsMiddlewareBuilder`).
- `internal/db` — MongoDB and Redis connection setup and collection handles.
- `internal/env` — environment/`.env` and `VERSION` loading.
- `internal/events` — asynchronous, batched event emitter that records domain
  events to the `events` collection.
- `internal/errmsg` — typed status errors per domain.
- `internal/utils` — shared helpers, including the badge-pile and judging
  (Gavel-style) algorithms.
- `internal/swagger` — embedded Swagger/OpenAPI spec and the `/docs` UI.

### Feature flags

Most participant- and judge-facing routes are gated behind feature flags via
`FlagsMiddlewareBuilder` (e.g. `teams_read`, `teams_write`, `submissions_*`,
`judging`, `voting`). The available flags are listed in `flags_config.json`, and
staged rollout configuration lives in `flagstages_config.json`. Superusers
manage these at runtime through `/superusers/flags` and `/superusers/flagstages`.

### Judging

Judges don't score projects on an absolute scale; they make **pairwise
comparisons** ("team A is better than team B"). Those comparisons are turned into
a ranking by the Crowd Bradley-Terry (Crowd-BT) scorer in
`internal/utils/gavel.go`.

The model tracks, per team, a skill estimate `mu` and its uncertainty
`sigma_sq`, and, per judge, a reliability described by a Beta distribution
(`alpha`, `beta`). It fits these together with an EM-style loop (a fixed 10
iterations) that alternates two passes over all recorded judgments:

1. **Update judges** — each judge's `alpha`/`beta` is nudged based on how often
   their calls agree with the current team estimates, then regularized back
   toward a prior. Consistent judges gain reliability; erratic ones lose it.
2. **Update teams** — each team's `mu`/`sigma_sq` is moved by a Bradley-Terry
   update for every comparison it took part in, **weighted by the reliability of
   the judge who made it**, so unreliable judges move the standings less.

`win_probability` is the standard Bradley-Terry `exp(mu_w) / (exp(mu_w) +
exp(mu_l))`, adjusted for uncertainty and clamped to `[0.01, 0.99]`. Final
standings come from `RankTeams`, sorting by `mu` descending and breaking ties by
lower deviation (higher confidence). `ScoreCrowdBT` is the one-call entry point.

### Badge piles

To spread badge pickup across multiple physical queues, each participant is
deterministically assigned to one of `BADGE_PILES` "piles"
(`internal/utils/badge_pile.go`). The pile is `(fnv32a(accountID) XOR salt) %
BADGE_PILES`, so the same account always lands in the same pile and staff can
pre-stage badges.

The `salt` is chosen to keep the piles evenly sized. `ChooseBestSalt` scans salt
values starting from 0, and for each one bins every account and scores the
result with `BalanceScore` — a chi-squared statistic against an even split, plus
a tiny `max−min` tie-breaker (lower is better, 0 is perfect). The search stops on
a perfect split or when it hits its evaluation/time budget, returning the best
salt found. That salt is persisted in the `settings` collection and reloaded
into `BADGE_PILES_SALT` at startup (`initBadgePileSalt` in `internal/app.go`).

## Deployments & data isolation

The server is started with a **deployment profile** (`dev`, `test`, or `prod`)
that controls which MongoDB database and Redis logical DB it talks to:

- MongoDB: the deployment name is used as the database name; collections are
  shared by name (`accounts`, `teams`, `flags`, `events`, …).
- Redis: `prod` → DB 0, `dev` → DB 1, `test` → DB 2.

Note that cache reads/writes in `internal/db` are no-ops unless the profile is
`prod`, so dev/test always hit MongoDB directly.

## Configuration

Configuration comes from a `.env` file (loaded via `godotenv`) plus a `VERSION`
file at the repo root. Environment keys consumed by `internal/env`:

| Key           | Purpose |
|---------------|---------|
| `MONGO_URI`   | MongoDB connection string |
| `JWT_SECRET`  | Secret used to sign/verify auth tokens |
| `BADGE_PILES` | Number of badge piles to balance into |
| `PREFORK`     | Enables Fiber prefork mode when `true` |
| `NO_HYPER`    | Disables hypervisor-oriented Swagger version stamping when `true` |

Redis is expected at `127.0.0.1:6379`. The listen **port** and **deployment
profile** are passed as CLI flags, not env vars.

## Running locally

Requires Go (see `go.mod` for the version), plus reachable MongoDB and Redis
instances.

```bash
# dev profile on port 9000 (see the RUNDEV.sh script)
./RUNDEV.sh

# or invoke the binary directly
go run ./cmd/server --deployment dev --port 9000
```

`main.go` flags:

```
--deployment <dev|test|prod>   (required)
--port <port>                  (required)
--env-root <dir>               directory containing the .env file (optional)
--app-version <version>        overrides the VERSION file (optional)
```

## Testing

Integration and unit specs live under `test/`, mirroring the package layout,
and run against the `test` profile (Redis DB 2). MongoDB and Redis must be
running.

```bash
./TEST.sh                    # go test ./test/... -v -count=1 -p 1
go test ./test/... -count=1  # equivalent core invocation
```

`BATCH_INITIALIZE.sh` and `cmd/batchinitialize` seed participant data from a
CSV for end-to-end scenarios.

## Build & Release

| Script        | Purpose |
|---------------|---------|
| `./BUILD.sh`    | Builds a stripped, static `linux/amd64` binary into `bin/<VERSION>` |
| `./TEST.sh`     | Runs the Go test suite |
| `./API_SPEC.sh` | Regenerates the Swagger docs via `swag init` |

Versions follow a `YY.MM.DD.B` scheme (build number `B` increments within a
day). `BUILD.sh` names artifacts after the version, so multiple builds coexist in
`bin/`.

To cut a release, a developer runs `./RELEASE.sh`: it bumps `VERSION` (to the next
`YY.MM.DD.B`), stamps the Swagger metadata for that version, then commits, tags
`v<version>`, and pushes. The pushed tag is what the hypervisor picks up to build
and deploy.

## API documentation

Swagger UI is served at `/docs`, with the OpenAPI JSON at `/docs/doc.json`. The
spec is embedded in the binary and generated from annotations in
`cmd/server/main.go` and the `swagger_models.go` files in each domain package.

## Hypervisor integration

This service is intended to run as a managed instance under
[openhack-hypervisor](https://github.com/openlabsro/openhack-hypervisor). A few
conventions in this repo support that:

- **Lifecycle scripts** (`BUILD.sh`, `TEST.sh`) give the hypervisor stable
  entrypoints to test and build a checkout.
- **Version-named artifacts** (`bin/<VERSION>`) and the `VERSION` file let
  multiple versions be built and deployed side by side.
- **Swagger version stamping**: by default the served OpenAPI doc has its
  `basePath` set to `/<VERSION>` so docs line up with the version-prefixed route
  the hypervisor mounts. Setting `NO_HYPER=true` disables this stamping.

For local development without the hypervisor, just use `./RUNDEV.sh` (see
[Running locally](#running-locally)).
