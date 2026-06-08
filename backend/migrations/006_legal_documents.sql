-- ============================================================
-- Migration 006: Legal Documents & Acceptance Tracking
-- Compliance: Indian DPDP Act + Consumer Protection Rules
-- ============================================================

-- 1. Master table for versioned legal documents
CREATE TABLE IF NOT EXISTS legal_documents (
    id             TEXT PRIMARY KEY,
    slug           TEXT NOT NULL UNIQUE,
    title          TEXT NOT NULL,
    version        TEXT NOT NULL DEFAULT 'v1.0',
    effective_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    pdf_hash       TEXT NOT NULL DEFAULT '',
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    worker_only    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

-- 2. User legal acceptance records (audit trail)
CREATE TABLE IF NOT EXISTS user_legal_acceptance (
    id                TEXT PRIMARY KEY,
    user_id           TEXT NOT NULL,
    document_id       TEXT NOT NULL REFERENCES legal_documents(id),
    document_slug     TEXT NOT NULL,
    accepted_version  TEXT NOT NULL,
    accepted_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address        TEXT NOT NULL DEFAULT '',
    device_info       TEXT NOT NULL DEFAULT '',
    pdf_hash          TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_user_legal_acceptance_user
    ON user_legal_acceptance (user_id, document_slug);

CREATE INDEX IF NOT EXISTS idx_user_legal_acceptance_doc
    ON user_legal_acceptance (document_id, accepted_at DESC);

-- 3. Separate worker policy acceptance table
CREATE TABLE IF NOT EXISTS worker_policy_acceptance (
    id                     TEXT PRIMARY KEY,
    user_id                TEXT NOT NULL UNIQUE,
    worker_policy_version  TEXT NOT NULL,
    accepted_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address             TEXT NOT NULL DEFAULT '',
    device_info            TEXT NOT NULL DEFAULT '',
    pdf_hash               TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_worker_policy_user
    ON worker_policy_acceptance (user_id);

-- ============================================================
-- Seed the 5 legal documents
-- ============================================================
INSERT INTO legal_documents (id, slug, title, version, effective_date, pdf_hash, is_active, worker_only)
VALUES
    ('ld-terms-conditions',    'terms-conditions',    'Terms & Conditions',         'v1.0', NOW(), '', TRUE, FALSE),
    ('ld-privacy-policy',      'privacy-policy',      'Privacy Policy',             'v1.0', NOW(), '', TRUE, FALSE),
    ('ld-liability-disclaimer','liability-disclaimer','Liability Disclaimer',       'v1.0', NOW(), '', TRUE, FALSE),
    ('ld-dispute-resolution',  'dispute-resolution',  'Dispute Resolution Policy',  'v1.0', NOW(), '', TRUE, FALSE),
    ('ld-worker-responsibility','worker-responsibility','Worker Responsibility Policy','v1.0', NOW(), '', TRUE, TRUE)
ON CONFLICT (slug) DO NOTHING;
