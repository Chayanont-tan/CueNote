CREATE TABLE shadowing_attempts (
    id                    BIGSERIAL PRIMARY KEY,
    user_id               BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    flashcard_sentence_id BIGINT NOT NULL REFERENCES flashcard_sentences(id) ON DELETE CASCADE,
    transcript            TEXT NOT NULL,
    score                 INT NOT NULL,
    correct_words         JSONB NOT NULL DEFAULT '[]',
    mispronounced_words   JSONB NOT NULL DEFAULT '[]',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shadowing_attempts_user_id ON shadowing_attempts(user_id);
CREATE INDEX idx_shadowing_attempts_sentence_id ON shadowing_attempts(flashcard_sentence_id);
