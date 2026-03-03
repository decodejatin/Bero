-- ============================================================
-- Migration 001: Location & Map Infrastructure for Bero
-- Requires: PostGIS extension (use postgis/postgis Docker image)
-- ============================================================

-- 1. Enable PostGIS
CREATE EXTENSION IF NOT EXISTS postgis;

-- 2. Add geospatial columns to worker_profiles
ALTER TABLE worker_profiles
    ADD COLUMN IF NOT EXISTS latitude    DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS longitude   DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS location    geometry(Point, 4326),
    ADD COLUMN IF NOT EXISTS is_available BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS h3_index    TEXT;

-- 3. Add geospatial columns to jobs
ALTER TABLE jobs
    ADD COLUMN IF NOT EXISTS location geometry(Point, 4326),
    ADD COLUMN IF NOT EXISTS h3_index TEXT;

-- 4. Populate location geometry from existing lat/lon data (jobs)
UPDATE jobs
SET location = ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)
WHERE latitude IS NOT NULL
  AND longitude IS NOT NULL
  AND location IS NULL;

-- 5. Create GiST spatial indexes for <100ms query performance
CREATE INDEX IF NOT EXISTS idx_worker_profiles_location_gist
    ON worker_profiles USING GIST (location);

CREATE INDEX IF NOT EXISTS idx_jobs_location_gist
    ON jobs USING GIST (location);

-- 6. B-tree indexes on H3 for kRing lookups
CREATE INDEX IF NOT EXISTS idx_worker_profiles_h3_index
    ON worker_profiles (h3_index);

CREATE INDEX IF NOT EXISTS idx_jobs_h3_index
    ON jobs (h3_index);

-- 7. Composite index for nearby worker query (availability + location)
CREATE INDEX IF NOT EXISTS idx_worker_profiles_available_location
    ON worker_profiles USING GIST (location)
    WHERE is_available = TRUE;

-- 8. Trigger to auto-update location geometry when lat/lon changes (workers)
CREATE OR REPLACE FUNCTION update_worker_location_geom()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.latitude IS NOT NULL AND NEW.longitude IS NOT NULL THEN
        NEW.location := ST_SetSRID(ST_MakePoint(NEW.longitude, NEW.latitude), 4326);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_worker_location_geom ON worker_profiles;
CREATE TRIGGER trg_worker_location_geom
    BEFORE INSERT OR UPDATE OF latitude, longitude ON worker_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_worker_location_geom();

-- 9. Trigger to auto-update location geometry when lat/lon changes (jobs)
CREATE OR REPLACE FUNCTION update_job_location_geom()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.latitude IS NOT NULL AND NEW.longitude IS NOT NULL THEN
        NEW.location := ST_SetSRID(ST_MakePoint(NEW.longitude, NEW.latitude), 4326);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_job_location_geom ON jobs;
CREATE TRIGGER trg_job_location_geom
    BEFORE INSERT OR UPDATE OF latitude, longitude ON jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_job_location_geom();
