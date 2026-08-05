CREATE TABLE vocabularies (
    id             SERIAL PRIMARY KEY,
    word           VARCHAR(100) UNIQUE NOT NULL,
    part_of_speech VARCHAR(50) NOT NULL,
    meaning_th     VARCHAR(255) NOT NULL,
    tag            VARCHAR(50) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_vocabularies_tag ON vocabularies(tag);
