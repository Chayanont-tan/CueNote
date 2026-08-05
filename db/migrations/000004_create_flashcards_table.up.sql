CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE flashcards (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vocabulary_id        BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,
    image_url            TEXT NOT NULL,
    user_sentence        TEXT,
    selected_ai_sentence TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_flashcards_user_id ON flashcards(user_id);
CREATE INDEX idx_flashcards_vocabulary_id ON flashcards(vocabulary_id);
