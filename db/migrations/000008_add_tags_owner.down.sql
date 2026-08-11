ALTER TABLE tags DROP CONSTRAINT tags_created_by_name_key;
ALTER TABLE tags ADD CONSTRAINT tags_name_key UNIQUE (name);
ALTER TABLE tags DROP COLUMN created_by;
