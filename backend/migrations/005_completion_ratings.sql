-- ============================================================
-- Migration 005: Dual-Sided Completion + Mandatory Ratings
-- ============================================================

-- 1. Add completion tracking columns to jobs
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS worker_completed_at TIMESTAMPTZ;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS client_confirmed_at TIMESTAMPTZ;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS worker_rated BOOLEAN DEFAULT FALSE;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS client_rated BOOLEAN DEFAULT FALSE;

-- 2. Mutual ratings table (one rating per rater per job)
CREATE TABLE IF NOT EXISTS mutual_ratings (
    id           TEXT PRIMARY KEY,
    job_id       TEXT NOT NULL,
    rater_id     TEXT NOT NULL,
    ratee_id     TEXT NOT NULL,
    rater_role   TEXT NOT NULL,
    rating_value INT NOT NULL CHECK (rating_value BETWEEN 1 AND 5),
    review_text  TEXT,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

-- Unique constraint: one rating per rater per job
CREATE UNIQUE INDEX IF NOT EXISTS idx_rater_job
    ON mutual_ratings (rater_id, job_id);

-- Indexes for lookups
CREATE INDEX IF NOT EXISTS idx_mutual_ratings_ratee
    ON mutual_ratings (ratee_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_mutual_ratings_job
    ON mutual_ratings (job_id);
