-- tags เดิมเป็น global catalog (unique แค่ name) แต่ที่จริงต้องแยกเป็นของแต่ละ user
-- ข้อมูล tag เก่า (สร้างตอนยังไม่มี owner) ไม่มีทางรู้ว่าเป็นของใคร เลยล้างทิ้งก่อน
-- (cascade ลบ vocabulary_tags ที่ผูกอยู่ไปด้วย — flashcards/vocabularies ของ user ไม่โดนกระทบ)
TRUNCATE TABLE tags CASCADE;

ALTER TABLE tags ADD COLUMN created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE tags DROP CONSTRAINT tags_name_key;
ALTER TABLE tags ADD CONSTRAINT tags_created_by_name_key UNIQUE (created_by, name);
