-- 001_init.sql: bootstrap schema for visa-tracker

CREATE TABLE IF NOT EXISTS visa_routes (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    slug          TEXT    NOT NULL UNIQUE,
    name          TEXT    NOT NULL,
    description   TEXT    NOT NULL DEFAULT '',
    requires_sponsor   INTEGER NOT NULL DEFAULT 0,
    requires_endorsement INTEGER NOT NULL DEFAULT 0,
    salary_threshold   INTEGER NOT NULL DEFAULT 0,  -- pence, 0 = no threshold
    duration_years     REAL    NOT NULL DEFAULT 0,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS salary_thresholds (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    visa_route_id INTEGER NOT NULL REFERENCES visa_routes(id),
    soc_code      TEXT    NOT NULL DEFAULT '',       -- empty = general threshold
    amount_pence  INTEGER NOT NULL,
    effective_date DATE   NOT NULL,
    notes         TEXT    NOT NULL DEFAULT '',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_thresholds_route ON salary_thresholds(visa_route_id);
CREATE INDEX idx_thresholds_soc   ON salary_thresholds(soc_code);

CREATE TABLE IF NOT EXISTS soc_codes (
    code          TEXT PRIMARY KEY,
    title         TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    going_rate_pence INTEGER NOT NULL DEFAULT 0,
    on_immigration_salary_list INTEGER NOT NULL DEFAULT 0,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sponsors (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT    NOT NULL,
    city          TEXT    NOT NULL DEFAULT '',
    county        TEXT    NOT NULL DEFAULT '',
    route         TEXT    NOT NULL DEFAULT '',       -- e.g. "Skilled Worker", "Intra-Company"
    rating        TEXT    NOT NULL DEFAULT '',       -- A-rated, B-rated
    sub_rating    TEXT    NOT NULL DEFAULT '',
    ingested_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_sponsors_name ON sponsors(name);
CREATE INDEX idx_sponsors_city ON sponsors(city);
CREATE INDEX idx_sponsors_route ON sponsors(route);

CREATE VIRTUAL TABLE IF NOT EXISTS sponsors_fts USING fts5(
    name, city, county, route,
    content='sponsors',
    content_rowid='id'
);

CREATE TRIGGER sponsors_ai AFTER INSERT ON sponsors BEGIN
    INSERT INTO sponsors_fts(rowid, name, city, county, route)
    VALUES (new.id, new.name, new.city, new.county, new.route);
END;

CREATE TRIGGER sponsors_ad AFTER DELETE ON sponsors BEGIN
    INSERT INTO sponsors_fts(sponsors_fts, rowid, name, city, county, route)
    VALUES ('delete', old.id, old.name, old.city, old.county, old.route);
END;

CREATE TABLE IF NOT EXISTS processing_times (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    visa_route_id INTEGER NOT NULL REFERENCES visa_routes(id),
    median_days   INTEGER NOT NULL DEFAULT 0,
    p90_days      INTEGER NOT NULL DEFAULT 0,
    snapshot_date DATE    NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_processing_route ON processing_times(visa_route_id);
