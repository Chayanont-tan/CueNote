CREATE TABLE ai_suggested_sentences (
    id            SERIAL PRIMARY KEY,
    flashcard_id  UUID NOT NULL REFERENCES flashcards(id) ON DELETE CASCADE,
    sentence_text TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_suggested_sentences_flashcard_id ON ai_suggested_sentences(flashcard_id);
