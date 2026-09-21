# Timestamp tools, and the product name

Status: Done
Depends on: `docs/plans/done/INLINE_URL_SOURCES_PLAN.md`

## Goal

Give an agent the two facts it otherwise guesses: what time it is now, and how long ago something happened. Both come
from this service rather than from the model's sense of the clock, so they are as authoritative as the container's
timezone.

Second, name the product correctly. The initialism LFU still stands, but it expands to Librechat Function Utilities;
this MCP is not only about files, and it is about to grow tools that have nothing to do with them.

## Non-goals

- Scheduling, timers, reminders, or any tool that acts on a time rather than reporting it. The two tools are pure
  functions of "now".
- Clock skew detection, NTP, or trusting a client-supplied clock. Time comes from this process only.
- Formatting or parsing dates for domains other than a timestamp and a duration: no cron, no business calendars, no
  user-facing date arithmetic beyond "how long since".
- Negotiating a per-request or per-user timezone. One timezone per deployment, the container's.

## Design

### Timezone: the container's

The reported timezone is the one the process runs in, which is what `TZ` sets and what `/etc/localtime` would
otherwise provide. No new envar: `TZ` is the conventional answer, and `time.Local` already resolves it.

The runtime image is distroless, which ships no zoneinfo, so `cmd/server` imports `time/tzdata` to embed the database.
Without it, a named zone such as `Australia/Sydney` would silently degrade to UTC, which would quietly reintroduce the
guessing these tools exist to remove. An unrecognised `TZ` value still lands on UTC, since neither `time.Local` nor
`LoadLocation` can be made to fail at startup without parsing `TZ` ourselves; UTC is at least an honest label, and the
`zone` and `offset` fields make a wrong one visible.

Both tools report in that zone and always include the numeric offset, so an answer is never ambiguous about what it is
relative to.

### `time`

Takes no arguments. Reports the instant in the forms an agent reasons with:

| Field      | Example                                   |
| ---------- | ----------------------------------------- |
| `iso`      | `2026-09-21T22:49:03+10:00`               |
| `unix`     | `1790004543`                              |
| `zone`     | `AEST`                                    |
| `offset`   | `+10:00`                                  |
| `readable` | `Sunday, 21 September 2026 at 22:49 AEST` |

The text block repeats the essentials on one line so a model that ignores structured output still gets them.

### `since`

Takes `time`, an ISO 8601 timestamp, and answers how long ago it was:

| Field     | Example                     |
| --------- | --------------------------- |
| `human`   | `3 hours ago`               |
| `seconds` | `10800`                     |
| `at`      | `2026-09-21T19:49:03+10:00` |
| `now`     | `2026-09-21T22:49:03+10:00` |

`seconds` is negative when the input is in the future, and `human` then reads `in 3 hours`, so a timestamp from a
different clock is not silently treated as the past.

Accepted inputs, in order: RFC 3339 with or without fractional seconds; the same with a space instead of the `T` that
RFC 3339 requires, which is a common way for a model to write one; then a bare `2006-01-02`, `2006-01-02 15:04:05`, and
the `T`-separated variants, which carry no zone and are therefore read in the deployment's zone. Anything else is a
tool error that names the offending input.

The ladder is deliberately coarse, because a model does not need a second-level duration and a person would not say
one: seconds up to a minute, minutes, hours, days, weeks, months (30 days), years (365 days). Under five seconds is
`just now`.

### Shape

The duration ladder and the parsing rules are pure, so they live in a new leaf package, `internal/timing`, alongside
`internal/utils` and `internal/sourcemap` in the dependency graph. It imports nothing of this module, and `internal/mcp`
is its only consumer. The handlers gain the zone and a clock, so a test can pin both.

Both tools are registered unconditionally, unlike `inline`, which needs a resolver.

## Implementation units

### Unit 1: the product name

Deliverables: `README.md`.

Acceptance criteria:

- [x] Every expansion of LFU in a live document reads "Librechat Function Utilities". `docs/plans/done` is untouched.

### Unit 2: the timing package

Deliverables: `internal/timing/timing.go`, `internal/timing/timing_test.go`, `cmd/server/main.go`.

Acceptance criteria:

- [x] `Stamp` reports an instant as RFC 3339, Unix seconds, zone name, numeric offset, and a spelled-out form.
- [x] `Stamp.ISO` round-trips through `time.Parse(time.RFC3339)`, and `Stamp.Offset` matches the instant's own offset
      for `+`/`-` and half-hour zones.
- [x] `HumanSince` returns `just now` under five seconds, counts up through each unit, singularises a count of one, and
      returns `in <duration>` for a future input.
- [x] `HumanSince` with equal inputs returns `just now`, and never returns an empty string.
- [x] `Parse` accepts RFC 3339, a space separator, fractional seconds, a date alone, and a zone-less date-time; the
      zone-less forms are read in the supplied location, and a zoned form ignores it.
- [x] `Parse` rejects an empty string and free text, naming the input in the error.
- [x] `cmd/server` imports `time/tzdata` so a named `TZ` resolves in the distroless image.
- [x] `go build ./...`, `go vet ./...`, and `go test ./... -race -count=1` pass, and `internal/timing` imports no other
      package of this module.

### Unit 3: the `time` tool

Deliverables: `internal/mcp/time.go`, `internal/mcp/time_test.go`, `internal/mcp/schema.go`, `internal/mcp/handlers.go`,
`internal/mcp/server.go`, `internal/mcp/inline.go`, `internal/mcp/server_test.go`.

Acceptance criteria:

- [x] `time` takes an empty object, and its output is the `Stamp` fields above.
- [x] The text block names the time and the zone, so the answer survives a client that drops structured output.
- [x] `tools/list` contains `time` with and without a resolver configured, and `inline`'s behaviour and schema are
      unchanged, its existing tests passing untouched.
- [x] A handler built with a fixed clock and a fixed zone returns exactly that instant in that zone, proven without
      waiting on the real clock.
- [x] The schema for each tool still reports `type: object`, so `TestToolInputSchemaIsAnObject` covers `time`.
- [x] `go build ./...`, `go vet ./...`, and `go test ./... -race -count=1` pass.

### Unit 4: the `since` tool

Deliverables: `internal/mcp/since.go`, `internal/mcp/since_test.go`, `internal/mcp/server_test.go`.

Acceptance criteria:

- [x] `since` with an RFC 3339 input three hours before a fixed clock returns `3 hours ago` and `10800` seconds.
- [x] A future input returns an `in ...` phrase and negative seconds.
- [x] A zone-less input is read in the deployment's zone, and a zoned input is honoured whatever the handler's zone is.
- [x] An input that is not a timestamp is a tool error naming the input, and the call does not panic.
- [x] `tools/list` contains `since` with and without a resolver configured.
- [x] `go build ./...`, `go vet ./...`, and `go test ./... -race -count=1` pass.

### Unit 5: documentation

Deliverables: `README.md`, `AGENTS.md`.

Acceptance criteria:

- [x] `README.md` drops "one tool", documents `inline`, `time`, and `since`, and shows `TZ` in the Docker example.
- [x] `AGENTS.md`'s diagram adds `internal/timing` with `mcp --> timing`, checked against `go list -deps`.
- [x] Verified by agent over HTTP against a running server with `TZ=Australia/Sydney`: `time` reports `AEST` and
      `+10:00`, and `since` on a timestamp from that same request's `Date` header reads `5 seconds ago`.
- [ ] Human check: in a real LibreChat client, `since` on a timestamp from the conversation reads as a sensible
      number of hours.

## Verification

Per unit, `go build ./...`, `go vet ./...`, and `go test ./... -race -count=1`. Both tools are pure functions of an
injected clock and zone, so every branch of the ladder, the parser, and the rendering is covered by a table test with
no deployment. What a test cannot settle is whether a real client renders the timestamp usefully and whether a real
container resolves its `TZ`; that is the Unit 5 human check.
