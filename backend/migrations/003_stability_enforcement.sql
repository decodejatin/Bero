-- ============================================================
-- Migration 003: Stability Enforcement
-- Analytics table + config for switching costs & cancellation limits
-- ============================================================

-- 1. Stability config (runtime-updatable parameters)
CREATE TABLE IF NOT EXISTS stability_config (
    id                      SERIAL PRIMARY KEY,
    switch_cost_fixed       DOUBLE PRECISION NOT NULL DEFAULT 0.15,
    earnings_penalty_percent DOUBLE PRECISION NOT NULL DEFAULT 5.0,
    max_cancels_per_hour    INT NOT NULL DEFAULT 3,
    escalation_threshold    INT NOT NULL DEFAULT 5,
    cooldown_minutes        INT NOT NULL DEFAULT 30,
    decay_lambda            DOUBLE PRECISION NOT NULL DEFAULT 0.01,
    travel_cost_per_km      DOUBLE PRECISION NOT NULL DEFAULT 5.0,
    visibility_penalty      DOUBLE PRECISION NOT NULL DEFAULT 0.2,
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- Seed with initial config (idempotent)
INSERT INTO stability_config (switch_cost_fixed, earnings_penalty_percent, max_cancels_per_hour)
SELECT 0.15, 5.0, 3
WHERE NOT EXISTS (SELECT 1 FROM stability_config LIMIT 1);

-- 2. Stability events analytics table
CREATE TABLE IF NOT EXISTS stability_events (
    id          TEXT PRIMARY KEY,
    event_type  TEXT NOT NULL,
    actor_id    TEXT NOT NULL,
    actor_role  TEXT NOT NULL,
    job_id      TEXT NOT NULL,
    details     TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Composite indexes for analytics queries
CREATE INDEX IF NOT EXISTS idx_stability_events_actor_type
    ON stability_events (actor_id, event_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_stability_events_type_time
    ON stability_events (event_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_stability_events_job
    ON stability_events (job_id, created_at DESC);
