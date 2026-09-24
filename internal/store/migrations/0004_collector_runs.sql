-- One row per collector invocation, whatever the outcome.
--
-- The collector VM powers itself off when it finishes and its serial console is
-- gone with it, so "what happened last night" has to live somewhere durable.
-- fetch_log records attempts against a source; this records the run itself, and
-- so captures the failures that happen before any fetch is attempted.

CREATE TABLE IF NOT EXISTS collector_runs (
  id           uuid PRIMARY KEY,
  command      text NOT NULL,          -- run | watch | migrate
  host         text NOT NULL,
  tier         text NOT NULL,
  version      text,
  started_at   timestamptz NOT NULL DEFAULT now(),
  finished_at  timestamptz,            -- null while running, or if it never returned
  ok           boolean,
  error        text,
  detail       jsonb NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS collector_runs_started_idx ON collector_runs (started_at DESC);
