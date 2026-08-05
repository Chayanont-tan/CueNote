CREATE TABLE IF NOT EXISTS tags (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS vocabulary_tags (
    vocabulary_id BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,
    tag_id        BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (vocabulary_id, tag_id)
);

DROP INDEX IF EXISTS idx_vocabularies_tag;
ALTER TABLE vocabularies DROP COLUMN IF EXISTS tag;

-- ตาราง sentences เดิมผูกประโยคกับ vocabulary_id โดยตรง ถูกแทนที่ด้วย
-- ai_suggested_sentences ที่ผูกกับ flashcard_id แทน (ดู 000005)
DROP TABLE IF EXISTS sentences;
