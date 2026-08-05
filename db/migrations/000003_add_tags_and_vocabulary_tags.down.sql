CREATE TABLE IF NOT EXISTS sentences (
    id            BIGSERIAL PRIMARY KEY,
    vocabulary_id BIGINT REFERENCES vocabularies(id) ON DELETE CASCADE,
    text          TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE vocabularies ADD COLUMN tag VARCHAR(50) NOT NULL DEFAULT '';
CREATE INDEX idx_vocabularies_tag ON vocabularies(tag);

DROP TABLE IF EXISTS vocabulary_tags;
DROP TABLE IF EXISTS tags;
