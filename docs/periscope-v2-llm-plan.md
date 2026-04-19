# Periscope V2 — LLM-Augmented Guidance Plan

## Document Status

- Status: Draft for implementation
- Date: 2026-04-19
- Depends on:
  [`periscope-spec.md`](./periscope-spec.md),
  [`periscope-v1-plan.md`](./periscope-v1-plan.md)
- Follows: `82ba482` (heuristic compact signal) and `66c1340` (heuristic
  rewind signal), which implement FR4/FR5 at the rule-based layer only.

## Objective

Layer LLM-derived semantic context on top of the existing heuristic V2
signals so Periscope can satisfy the parts of the spec that pure
heuristics cannot reach:

- **FR7**: operational guidance text (suggested rewind reprompt, compact
  focus, fork brief, subagent task wording, copyable command).
- **FR5**: branch-point *reasons* and *confidence* grounded in turn
  content, not just token shape.
- **FR6**: a recommendation engine that can compare candidate actions
  using semantic similarity, not just retry-signature overlap.
- **FR10**: an optional guidance agent that runs on a fresh context
  window and diagnoses the primary session from outside.

The heuristic signals stay authoritative for *whether* a signal fires.
The LLM layer only supplies *what text to show the user*.

## Architectural decisions (locked)

1. **LLM transport**: direct Anthropic SDK calls behind a thin internal
   package. Not the CLI subprocess path used by `internal/insight`.
   Rationale: simpler, faster, cacheable, and the existing CLI
   subprocess path is tuned for long single-shot insight generation.
   Per-turn summarisation is many short calls with prompt caching, which
   the SDK path serves better.
2. **Execution model**: async. Summaries are produced by a background
   worker, persisted, and delivered to the UI via existing SSE.
   Synchronous fallbacks are allowed for Phase B when a summary happens
   to be ready at read time.
3. **Caching strategy**: content-addressed per turn
   (`session_id + end_ordinal + content_hash`). New turns invalidate
   nothing — only the new turn's summary is missing. Compacting a turn
   never changes its content hash so summaries survive re-reads.
4. **Evidence trail**: every LLM output carries explicit turn-ordinal
   refs so the UI can keep UX5 ("measured vs model-generated") honest.

## Scoping gate — starred sessions only

LLM summarisation is gated on the user *starring* a session. This keeps
cost predictable and avoids running generations on the entire local
archive. The trigger lifecycle:

- When a user stars a session (`PUT /sessions/{id}/star`) that does not
  have summaries yet, a background job enqueues per-turn summarisation.
- Unstarring does not delete existing summaries (they are cheap to keep;
  re-starring is instant).
- A user may still open an un-starred session's Context page. No LLM
  text appears; heuristic signals still show.

## Phases

Each phase is testable end-to-end before commit.

- **Phase A — turn-summary infrastructure.** Schema, Anthropic client,
  worker, star-trigger. No UI surface beyond a small badge showing
  summary coverage. Ship behind an `ANTHROPIC_API_KEY` check.
- **Phase B — operational text on existing banners.** Rewind banner
  gains `rewind_reprompt_text` and a tangent label. Compact banner
  gains `compact_focus_text`. Uses Phase-A summaries.
- **Phase C — guidance agent (FR10).** New `analyze` endpoint, new
  cache table, new `ContextGuidancePanel` component. Runs in an
  isolated subagent for live sessions.
- **Phase D — branch-point + recommendation engine (FR5/FR6).** Pivot
  detection via topic-vector drift, candidate-action ranking across all
  five actions (continue/rewind/compact/fork/subagent).

---

## Phase A — Turn-Summary Infrastructure

### A.1 Schema

One new table in `internal/db/schema.sql`:

```sql
CREATE TABLE IF NOT EXISTS context_turn_summaries (
    id              INTEGER PRIMARY KEY,
    session_id      TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    turn_index      INTEGER NOT NULL,
    start_ordinal   INTEGER NOT NULL,
    end_ordinal     INTEGER NOT NULL,
    content_hash    TEXT NOT NULL,
    summary         TEXT NOT NULL,      -- short prose summary
    intent          TEXT NOT NULL,      -- "implement"|"debug"|"read"|...
    outcome         TEXT NOT NULL,      -- "progress"|"stuck"|"failed"|"off-topic"
    topic           TEXT NOT NULL,      -- short topic label
    files_touched   TEXT NOT NULL,      -- JSON array
    tags            TEXT NOT NULL,      -- JSON array
    model           TEXT NOT NULL,
    prompt_version  INTEGER NOT NULL,
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    UNIQUE(session_id, turn_index, content_hash)
);

CREATE INDEX IF NOT EXISTS idx_context_turn_summaries_session
    ON context_turn_summaries(session_id, turn_index);
```

Migration is column-only on an existing DB (SQLite `CREATE TABLE IF
NOT EXISTS`). Bumping `dataVersion` is not required — the table is
additive.

`content_hash` is a stable hash over the turn's user message, assistant
message, and ordered tool-call signatures. This lets us skip
regenerating a summary when the turn's content has not changed.

### A.2 Anthropic client package

New `internal/llm/` package. Minimal interface:

```go
type Client interface {
    Summarize(ctx context.Context, req SummarizeRequest) (SummarizeResult, error)
    Generate(ctx context.Context, req GenerateRequest) (GenerateResult, error)
}
```

Implementation details:

- Direct HTTPS `POST /v1/messages` to `api.anthropic.com`; no third-party
  SDK dependency (project convention "prefer stdlib").
- Reads `ANTHROPIC_API_KEY` from env. If unset, the worker no-ops and
  the UI shows heuristic-only output.
- Uses prompt caching on the system prompt so many short per-turn calls
  reuse the same cache entry.
- Default model: `claude-haiku-4-5-20251001` for per-turn summaries
  (cheap, fast); `claude-sonnet-4-6` for guidance-agent calls in Phase
  C. Both overridable by env.
- Hard input-size cap per turn (truncate tool output previews); hard
  output cap via `max_tokens`.
- Retries with exponential backoff on 429/5xx.

### A.3 Worker

New `internal/summarize/worker.go`.

- Singleton goroutine owned by `server.Server`. Job queue is an
  unbuffered channel; enqueue = "summarise session X". Jobs dedupe on
  session ID.
- Per session, the worker:
  1. Reads messages via existing `db.GetAllMessages`.
  2. Reconstructs turns via the same `buildTimelineTurns` logic already
     used in `internal/server/context.go`. This is extracted to a
     shared helper so the worker doesn't reimplement it.
  3. For each turn missing a summary (by `content_hash`), calls
     `llm.Client.Summarize` and writes the row.
  4. Broadcasts an SSE update so open Context tabs refresh.
- Bounded concurrency: one in-flight generation at a time per process
  (tunable). API key absence short-circuits the loop.
- Graceful shutdown via `context.Context` wired from server startup.

### A.4 Star trigger

Hook into `handleStarSession`: after a successful star, the server
calls `worker.EnqueueSession(sessionID)`. Existing rows mean the
worker is a near-no-op; fresh stars kick off a generation pass.

Also add a startup reconciliation pass: on boot, enumerate starred
sessions and enqueue any whose turn-summary coverage is incomplete.
Prevents losing work across restarts.

### A.5 API surface

Augment the existing `GET /sessions/{id}/context` response with
`summary_coverage`:

```json
"summary_coverage": {
  "total_turns": 23,
  "summarised_turns": 17,
  "status": "pending" | "complete" | "disabled" | "idle",
  "last_updated_at": "..."
}
```

`status = disabled` when no API key; `idle` when session is not
starred and thus not scheduled; `pending` during generation;
`complete` when all turns have summaries.

No new endpoint in Phase A — the payload rides along with the
existing context response.

### A.6 UI (Phase A only)

Small pill near the Context Guidance section header: "Summaries
17/23" → "Summaries ready". Clicking a starred-but-idle session
shows a hint: "Summaries generate when a session is starred." No
banner text changes yet — that's Phase B.

### A.7 Tests

- `internal/llm`: fake HTTP server, asserts request shape (model,
  cache-control on system prompt, max_tokens, retry on 429).
- `internal/summarize`: in-memory worker + stub client; asserts that
  only missing turns are generated, content-hash stability, and that
  star triggers the right enqueue.
- `internal/db`: schema migration + round-trip for the new table.

### A.8 Phase A end-to-end test

1. Ensure `ANTHROPIC_API_KEY` is set.
2. Star a session that has ≥ 5 turns.
3. Open its Context page. Pill shows "Summaries 0/N" →
   "Summaries N/N" within ~30s.
4. `SELECT summary, intent, topic FROM context_turn_summaries
   WHERE session_id = ?` shows meaningful per-turn rows.
5. Unset the key and restart: pill shows "Summaries disabled".

---

## Phase B — Operational Text on Existing Banners

### B.1 Data flow

When `handleGetSessionContext` computes the rewind or compact signal,
it additionally loads the per-turn summaries for the affected turn
range and calls `llm.Client.Generate` with a targeted prompt. Results
are cached in-memory per signal+content-hash so repeated reads do not
regenerate.

- **Rewind prompt input**: summaries of turns `BadStretchFrom..BadStretchTo`
  + the last clean turn's summary + overall task topic.
- **Rewind prompt output**: `{ tangent_label, rewind_reprompt_text }`.
- **Compact prompt input**: older-turn summaries + recent-turn summaries
  + dominant low-value categories.
- **Compact prompt output**: `{ keep_items: [...], drop_items: [...],
  compact_focus_text }`.

If the session has no summaries yet, the banners render exactly as
today (heuristic-only). No blocking.

### B.2 Response additions

```json
"rewind_signal": {
  ...existing fields...,
  "tangent_label": "debugging the deployment warning branch",
  "rewind_reprompt_text": "Ignore the deployment warning branch...",
  "reprompt_provenance": "model-generated",
  "evidence_turns": [85, 91, 103]
}

"compact_signal": {
  ...existing fields...,
  "keep_items": ["auth refactor plan", "validated constraint B"],
  "drop_items": ["exploratory test debugging branch"],
  "compact_focus_text": "Preserve auth refactor...",
  "focus_provenance": "model-generated",
  "evidence_turns": [4, 7, 9, 12]
}
```

### B.3 UI additions

Both banners gain a collapsed "Suggested prompt" block below their
reasons list:

- Header: "Suggested rewind prompt" / "Suggested compact focus"
- Body: the generated text, monospace, with a copy button.
- Footer micro-label: "model-generated · claude-haiku-4-5" so the
  user can tell it apart from heuristic facts (UX5).
- If the session has no summaries yet, this block is replaced with a
  small "Star to enable guidance text" hint.

### B.4 Tests

- Backend: stub `llm.Client`, assert that banners include generated
  text when summaries cover the signal's turn range, and fall back
  gracefully when they don't.
- Frontend: component test renders copy button, handles missing text,
  renders the "star to enable" hint when `summary_coverage.status ==
  idle`.

### B.5 Phase B end-to-end test

1. Star a session with a rewind heuristic hit.
2. Wait for summaries.
3. Refresh: rewind banner shows "Suggested rewind prompt" with a
   copy-to-clipboard button, plus a tangent label in the banner
   headline ("Rewind to turn X — drop the deployment-warning branch").
4. Same for a session that triggers the compact heuristic: compact
   banner shows "Preserve X, drop Y".
5. Open an un-starred session with the same heuristic hits: banners
   render as they do today, with the "Star to enable guidance text"
   hint.

---

## Phase C — Guidance Agent (FR10)

### C.1 New endpoint

```
POST /api/v1/sessions/{id}/context/analyze
GET  /api/v1/sessions/{id}/context/guidance
```

`POST` enqueues an analysis run. `GET` returns the most recent cached
result. Results are persisted in a new `context_guidance` table keyed on
`session_id + input_fingerprint`.

### C.2 Input bundle

Per spec §Guidance-Agent > Inputs. Never the raw transcript; always the
structured bundle built from:

- session metadata,
- occupancy + capacity,
- composition breakdown,
- timeline summary (per-turn summaries from Phase A),
- detected branch points and compactions,
- top growth spikes (from heuristics),
- recent turns' full text,
- task summary (first user message + dominant topics).

### C.3 Output

One JSON document containing diagnosis, primary recommendation,
alternatives, rationale, copyable command, suggested prompt/handoff,
confidence, and `evidence_refs` citing turn ordinals.

### C.4 Live-session isolation

The analysis call runs as a Claude Agent SDK subagent with its own
context window. Implementation choice: a separate HTTP call using a
fresh system prompt and no shared state — the spec line 885–886
constraint is satisfied by construction because we never thread the
live session's context into the call.

### C.5 UI

New `ContextGuidancePanel.svelte` component under "Context Guidance",
shown below the existing banners. States: no-run-yet (button "Run
guidance agent"), running (spinner), ready (card with diagnosis +
copyable action text + "why" evidence list).

### C.6 Phase C end-to-end test

Starred session with high-occupancy + tangent → "Run guidance agent"
→ panel populates with diagnosis + exact rewind-or-compact command
text + cited turns.

---

## Phase D — Branch-Point + Recommendation Engine (FR5/FR6)

### D.1 Pivot detection

Using Phase-A topic labels, compute adjacent-window cosine distance
(or coarse topic-cluster hop count since topics are short strings).
Fork recommendation fires when the trailing window's topic set has
~zero overlap with the leading window's.

### D.2 Candidate scoring

For each candidate action (continue / rewind / compact / fork / fresh
/ subagent), compute a score from heuristic inputs + LLM-derived
inputs (pivot strength, tangent cohesion, etc.). Pick the primary
action; return the top 2 alternatives.

### D.3 Branch-point surface

Existing `ContextBranchPointList` component (per V1 spec) is added
here. Each row: turn index, reason, recommended action, estimated
benefit, confidence, link to evidence.

### D.4 Phase D end-to-end test

A session with a genuine task pivot (e.g. implementation → docs)
surfaces a "Fork recommended" branch point with a handoff brief.
Sessions without a pivot never surface this card.

---

## Implementation order and commits

1. Phase A schema + client + worker + star trigger + UI pill + tests.
   → test end-to-end → commit.
2. Phase B banner augmentations + prompt package + UI + tests. →
   test end-to-end → commit.
3. Phase C endpoint + guidance panel + cache table + tests. → test
   → commit.
4. Phase D pivot detection + scoring + branch-point UI + tests. →
   test → commit.

Each commit is self-contained and leaves the app fully functional
without the next phase.

## Risks and open questions

- **Rate limits**: per-turn summarisation on a 100-turn session issues
  100 calls. Haiku + prompt cache keeps this cheap but we should
  throttle.
- **API key provisioning**: the spec says "local by default". We
  require the user to set `ANTHROPIC_API_KEY` themselves; no key in
  the binary.
- **Content-hash stability**: changes to the hash function invalidate
  all existing summaries. Keep the function committed and versioned.
- **Prompt drift**: bumping `prompt_version` invalidates old rows on
  the next scan, which is the desired behaviour.
- **Embedded-tab UX**: the Context page is also embedded inside the
  session detail view; the new pill and banner text must render in
  both contexts.
