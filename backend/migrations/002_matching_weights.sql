-- ============================================================
-- Migration 002: Matching Weights Configuration
-- Dynamic weight storage for Hungarian matching pre-processing
-- ============================================================

-- 1. Create matching_weights config table
CREATE TABLE IF NOT EXISTS matching_weights (
    id                  SERIAL PRIMARY KEY,
    distance_weight     DOUBLE PRECISION NOT NULL DEFAULT 0.4,
    reputation_weight   DOUBLE PRECISION NOT NULL DEFAULT 0.3,
    skill_weight        DOUBLE PRECISION NOT NULL DEFAULT 0.2,
    wait_penalty_weight DOUBLE PRECISION NOT NULL DEFAULT 0.1,
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Seed with initial config row (idempotent)
INSERT INTO matching_weights (distance_weight, reputation_weight, skill_weight, wait_penalty_weight)
SELECT 0.4, 0.3, 0.2, 0.1
WHERE NOT EXISTS (SELECT 1 FROM matching_weights LIMIT 1);

-- 3. Index on worker_profiles.skills for JSONB overlap queries
-- Used by candidate filtering: skills::jsonb ?| ARRAY[...]
CREATE INDEX IF NOT EXISTS idx_worker_profiles_skills_gin
    ON worker_profiles USING GIN (skills);
