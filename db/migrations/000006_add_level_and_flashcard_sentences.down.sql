ALTER TABLE flashcards ADD COLUMN user_sentence TEXT;
ALTER TABLE flashcards ADD COLUMN selected_ai_sentence TEXT;

DROP TABLE IF EXISTS flashcard_sentences;

ALTER TABLE vocabularies DROP COLUMN IF EXISTS level;
