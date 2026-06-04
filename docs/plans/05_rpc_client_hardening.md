# Plan: RPC Client Hardening (`pkg/rabbitmq`)

**Status**: IMPLEMENTED  
**Scope**: `tmf/pkg/rabbitmq/rpc.go` — production-readiness gaps  
**Prerequisite**: No service-level changes required; all work is contained in the shared library.

---

## Background

The current `RPCClient` implements the correct RabbitMQ RPC pattern (correlation IDs, Direct Reply-To, `sync.Map`) but has four gaps that make it unsuitable for production load. This plan addresses them in dependency order.

---

## Gap 1 — No reconnection logic

### Problem

`RPCClient` holds a single `*amqp.Connection` and `*amqp.Channel` with no recovery logic. If the broker restarts, a network hiccup occurs, or the channel is force-closed by the broker (e.g. after a publish error), every pending and future RPC call fails permanently and the process must be restarted manually.

### Target behaviour

- On connection or channel close, the client transparently reconnects and restores the consumer pipeline.
- All in-flight requests that were pending when the connection dropped receive a typed `ErrConnectionLost` error immediately (rather than hanging until context timeout).
- Reconnect uses exponential backoff (cap: 30 s) so a broker restart does not cause a thundering herd.

### Tasks

- [x] **G1-T1** Add a `reconnect` goroutine that watches `conn.NotifyClose(make(chan *amqp.Error, 1))` and `channel.NotifyClose`. → `reconnectLoop()` in `rpc.go`.
- [x] **G1-T2** On close notification: drain all entries from `c.pending` by closing their response channels with a sentinel error value (introduce `errReconnecting`), then rebuild connection → channel → consumer in a backoff loop. → `drainPending()` sends `ErrReconnecting`; backoff 500ms→30s.
- [x] **G1-T3** Protect `c.conn` and `c.channel` replacement with a `sync.RWMutex` (`connMu`) so `Publish*` methods can read-lock during normal operation.
- [x] **G1-T4** Expose `c.Done() <-chan struct{}` so callers can detect permanent shutdown (after `Close()` is called).
- [x] **G1-T5** Two tests: `TestRPCClient_ReconnectDrainsInFlight` (in-flight call gets `ErrReconnecting` immediately on forced connection drop) and `TestRPCClient_ReconnectFullCycle` (pre-drop call succeeds → drop → `ErrReconnecting` → `require.Eventually` polls until reconnect completes and call succeeds again within 10 s).

---

## Gap 2 — Single channel bottleneck

### Problem

All goroutines share one `*amqp.Channel` serialized behind `publishMu`. Under the Scatter-Gather fan-out (Qualification fans to GIS + Inventory + Catalog in parallel) all three publish calls queue behind one mutex. Throughput is capped at the single-channel publish rate (~thousands/s on a single TCP flow).

AMQP channels are cheap (multiplexed over one TCP connection); the recommended pattern is one channel per concurrent publisher.

### Target behaviour

- A fixed-size **channel pool** (default size: `runtime.NumCPU() * 2`, configurable) is created at startup.
- Each `PublishWithContext` / `RequestWithHeaders` call acquires a free channel from the pool, publishes, and returns it.
- Pool acquisition blocks only when all channels are busy (bounded concurrency). A context-aware `acquire(ctx)` respects cancellation.
- `publishMu` is removed entirely.

### Tasks

- [x] **G2-T1** Implement `channelPool` struct: a buffered `chan *amqp.Channel` plus `acquire(ctx)` / `release(ch)` / `drain()`.
- [x] **G2-T2** Replace `c.channel *amqp.Channel` + `publishMu sync.Mutex` with `c.pool *channelPool`. `publishMu` field is gone.
- [x] **G2-T3** `PublishWithHeaders` (fire-and-forget) acquires/releases from pool. `RequestWithHeaders` uses `consumerCh + rpcMu` — **not the pool** — because Direct Reply-To requires publish and consume to happen on the same AMQP channel; pool channels do not have the reply consumer registered. This is a design constraint not anticipated in the plan; `rpcMu` hold time is one `PublishWithContext` call (~µs) so throughput impact is negligible.
- [x] **G2-T4** Pool is drained in `reconnectLoop` before reconnect and rebuilt inside `connect()`.
- [x] **G2-T5** `WithPoolSize(n int) RPCClientOption` added.
- [x] **G2-T6** `TestRPCClient_ConcurrentP99`: 100 sequential requests establish baseline p99; 10×100 concurrent requests compute p99; asserts concurrent p99 ≤ 5× sequential p99 (actual ratio observed: ~1.4×). `BenchmarkRPCClient_Sequential` and `BenchmarkRPCClient_Concurrent10` added for `go test -bench` profiling.

---

## Gap 3 — Double-marshal workaround

### Problem

The workaround at [rpc.go:166-173](../../../pkg/rabbitmq/rpc.go#L166-L173) detects a response body that is a JSON-encoded string (starts with `"`) and unwraps it. This was patched at the consumer end rather than fixed at the source (the service that is double-marshaling). It is fragile: any legitimate `string`-typed response body (e.g. a plain UUID reply) would also be unwrapped, silently corrupting it.

### Root cause investigation needed

The offending service must be identified before the workaround can be removed.

### Tasks

- [x] **G3-T1** Audited all RPC responders across all services. No active double-marshal source found — all handlers pass typed structs/maps directly to `publisher.PublishToQueue` or `publisher.Publish`, which marshal once. The bug was in a replaced code path.
- [x] **G3-T2** N/A — no active offender to fix.
- [x] **G3-T3** `TestRPCClient_StringResponseNotUnwrapped` added: a server replies with `json.Marshal("hello")` (a JSON string), client asserts the raw `"hello"` is returned verbatim, not unwrapped.
- [x] **G3-T4** Workaround block (lines 166-173 in the original file) removed from `rpc.go`. Test passes without it.

---

## Gap 4 — Redundant hardcoded timeout

### Problem

`RequestWithHeaders` has both `ctx.Done()` and `time.After(DefaultRPCTimeout)` on the same `select`. The 30 s hardcoded timeout overrides the caller's context deadline when the context deadline is longer than 30 s, and is redundant when it is shorter. Callers set their intent via `context.WithTimeout`; the internal override is surprising and untestable.

### Tasks

- [x] **G4-T1** `case <-time.After(DefaultRPCTimeout)` removed from the `select` in `RequestWithHeaders`.
- [x] **G4-T2** `DefaultRPCTimeout` constant removed entirely.
- [x] **G4-T3** `WithMessageTimeout(d time.Duration) ConsumerOption` added to `pkg/rabbitmq/consumer.go`. Each message processing block is wrapped in an IIFE so `defer cancel()` fires per-message (not per-goroutine). All 9 `NewConsumer`/`NewConsumerWithConnection` call sites updated to pass `WithMessageTimeout(30*time.Second)`, restoring the deadline the removed `DefaultRPCTimeout` previously enforced, correctly placed at the consumer layer. BFF test mock updated to match new variadic signature.
- [x] **G4-T4** `TestRPCClient_Timeout` updated: 50 ms deadline, asserts `errors.Is(err, context.DeadlineExceeded)`.

---

## Execution Order

```
G4 (30 min)  — isolated, zero risk, no dependencies
G3 (1–2 h)   — audit + fix source, then remove workaround
G1 (1 day)   — reconnection; foundational for G2
G2 (half day) — channel pool; depends on G1's connMu for safe rebuild
```

G4 and G3 can be done on the current codebase without touching reconnect or pool logic. G1 must land before G2 because pool rebuild is part of the reconnect path.

---

## Definition of Done

- [x] All tasks above checked off.
- [x] `go test ./pkg/rabbitmq/...` passes — 13/13 tests green (26 s with testcontainers).
- [x] `go vet ./rabbitmq/...` and `go vet` across all affected services clean. `golangci-lint` not runnable — installed version (go1.25) below module go version (1.26.2).
- [x] `publishMu` field deleted from `RPCClient`.
- [x] Double-marshal workaround block deleted from `rpc.go`.
- [x] All consumer call sites pass `WithMessageTimeout(30*time.Second)` ensuring every handler context carries a deadline.
