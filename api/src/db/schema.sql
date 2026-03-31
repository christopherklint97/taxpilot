CREATE TABLE IF NOT EXISTS tax_returns (
    id          TEXT PRIMARY KEY,
    tax_year    INTEGER NOT NULL,
    state_code  TEXT NOT NULL DEFAULT 'CA',
    filing_status TEXT,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS field_values (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    return_id   TEXT NOT NULL REFERENCES tax_returns(id) ON DELETE CASCADE,
    field_key   TEXT NOT NULL,
    value_num   REAL,
    value_str   TEXT,
    source      TEXT NOT NULL DEFAULT 'user_input',
    updated_at  TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(return_id, field_key)
);

CREATE TABLE IF NOT EXISTS pdf_documents (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    return_id   TEXT NOT NULL REFERENCES tax_returns(id) ON DELETE CASCADE,
    form_id     TEXT,
    tax_year    INTEGER,
    doc_type    TEXT NOT NULL,
    file_path   TEXT NOT NULL,
    file_name   TEXT NOT NULL,
    uploaded_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS prior_year_values (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    return_id   TEXT NOT NULL REFERENCES tax_returns(id) ON DELETE CASCADE,
    source_year INTEGER NOT NULL,
    field_key   TEXT NOT NULL,
    value_num   REAL,
    value_str   TEXT,
    UNIQUE(return_id, source_year, field_key)
);

CREATE INDEX IF NOT EXISTS idx_field_values_return ON field_values(return_id);
CREATE INDEX IF NOT EXISTS idx_field_values_key ON field_values(return_id, field_key);
CREATE INDEX IF NOT EXISTS idx_pdf_documents_return ON pdf_documents(return_id);
CREATE INDEX IF NOT EXISTS idx_prior_year_return ON prior_year_values(return_id, source_year);
