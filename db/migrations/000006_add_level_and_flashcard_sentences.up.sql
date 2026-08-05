-- CEFR level ต่อคำศัพท์ (A1-C2) สำหรับ level filter chips ในหน้า Vocab List
ALTER TABLE vocabularies ADD COLUMN level VARCHAR(2) NOT NULL DEFAULT 'A1';

-- แต่งได้หลายประโยคต่อ flashcard (พิมพ์เอง หรือกดเลือกจาก AI suggestion)
-- แทนที่ flashcards.user_sentence / selected_ai_sentence แบบเดี่ยวเดิม
CREATE TABLE flashcard_sentences (
    id            SERIAL PRIMARY KEY,
    flashcard_id  UUID NOT NULL REFERENCES flashcards(id) ON DELETE CASCADE,
    sentence_text TEXT NOT NULL,
    source        VARCHAR(10) NOT NULL DEFAULT 'user',  -- 'user' หรือ 'ai_picked'
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_flashcard_sentences_flashcard_id ON flashcard_sentences(flashcard_id);

ALTER TABLE flashcards DROP COLUMN IF EXISTS user_sentence;
ALTER TABLE flashcards DROP COLUMN IF EXISTS selected_ai_sentence;
