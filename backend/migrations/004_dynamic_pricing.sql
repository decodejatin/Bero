-- ============================================================
-- Migration 004: Dynamic Pricing Engine
-- Per-H3-hexagon surge pricing with sigmoid market clearing
-- ============================================================

-- 1. Pricing config (runtime-updatable sigmoid parameters)
CREATE TABLE IF NOT EXISTS pricing_config (
    id                      SERIAL PRIMARY KEY,
    max_surge_multiplier    DOUBLE PRECISION NOT NULL DEFAULT 3.0,
    equilibrium_theta       DOUBLE PRECISION NOT NULL DEFAULT 1.5,
    elasticity_sensitivity  DOUBLE PRECISION NOT NULL DEFAULT 3.0,
    emergency_surge_cap     DOUBLE PRECISION NOT NULL DEFAULT 2.5,
    max_surge_change_rate   DOUBLE PRECISION NOT NULL DEFAULT 0.3,
    min_surge_multiplier    DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    supply_kring_radius     INT NOT NULL DEFAULT 1,
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- Seed defaults (idempotent)
INSERT INTO pricing_config (max_surge_multiplier, equilibrium_theta, elasticity_sensitivity)
SELECT 3.0, 1.5, 3.0
WHERE NOT EXISTS (SELECT 1 FROM pricing_config LIMIT 1);

-- 2. Surge history analytics table
CREATE TABLE IF NOT EXISTS surge_history (
    id               TEXT PRIMARY KEY,
    h3_index         TEXT NOT NULL,
    demand           INT NOT NULL,
    supply           INT NOT NULL,
    theta            DOUBLE PRECISION NOT NULL,
    surge_multiplier DOUBLE PRECISION NOT NULL,
    is_emergency     BOOLEAN DEFAULT FALSE,
    created_at       TIMESTAMPTZ DEFAULT NOW()
);

-- Composite indexes for per-hex time-series queries
CREATE INDEX IF NOT EXISTS idx_surge_history_hex_time
    ON surge_history (h3_index, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_surge_history_time
    ON surge_history (created_at DESC);

-- 3. Add surge columns to jobs table (store snapshot at booking time)
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS surge_multiplier DOUBLE PRECISION DEFAULT 1.0;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS surge_price DOUBLE PRECISION;
