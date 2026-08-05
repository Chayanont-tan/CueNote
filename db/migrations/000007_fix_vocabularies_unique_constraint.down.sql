ALTER TABLE vocabularies DROP CONSTRAINT vocabularies_word_pos_key;
ALTER TABLE vocabularies ADD CONSTRAINT vocabularies_word_key UNIQUE (word);
