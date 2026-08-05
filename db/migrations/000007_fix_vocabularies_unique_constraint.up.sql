-- SaveVocabulariesForTag ทำ ON CONFLICT (word, part_of_speech) แต่ constraint เดิมมีแค่
-- UNIQUE(word) เฉยๆ ทำให้ insert คำที่มีอยู่แล้วพัง (SQLSTATE 42P10) และห้ามคำเดียวกัน
-- มีได้หลาย part_of_speech (เช่น "book" เป็นได้ทั้ง noun/verb) — เปลี่ยนเป็น composite unique
ALTER TABLE vocabularies DROP CONSTRAINT vocabularies_word_key;
ALTER TABLE vocabularies ADD CONSTRAINT vocabularies_word_pos_key UNIQUE (word, part_of_speech);
