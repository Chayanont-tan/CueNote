INSERT INTO vocabularies (word, part_of_speech, meaning_th) VALUES
    ('wake up', 'verb', 'ตื่นนอน'),
    ('brush', 'verb', 'แปรงฟัน'),
    ('shower', 'noun', 'การอาบน้ำ'),
    ('tired', 'adjective', 'เหนื่อย'),
    ('relax', 'verb', 'พักผ่อน'),
    ('delicious', 'adjective', 'อร่อย'),
    ('hungry', 'adjective', 'หิว'),
    ('menu', 'noun', 'เมนูอาหาร'),
    ('spicy', 'adjective', 'เผ็ด'),
    ('portion', 'noun', 'ปริมาณอาหารต่อจาน'),
    ('turn', 'verb', 'เลี้ยว'),
    ('straight', 'adverb', 'ตรงไป'),
    ('left', 'adjective', 'ซ้าย'),
    ('corner', 'noun', 'หัวมุมถนน'),
    ('nearby', 'adjective', 'ใกล้เคียง')
ON CONFLICT (word) DO NOTHING;

INSERT INTO tags (name) VALUES
    ('daily_life'),
    ('food'),
    ('directions')
ON CONFLICT (name) DO NOTHING;

INSERT INTO vocabulary_tags (vocabulary_id, tag_id)
SELECT v.id, t.id
FROM vocabularies v
JOIN tags t ON (
    (t.name = 'daily_life' AND v.word IN ('wake up', 'brush', 'shower', 'tired', 'relax')) OR
    (t.name = 'food' AND v.word IN ('delicious', 'hungry', 'menu', 'spicy', 'portion')) OR
    (t.name = 'directions' AND v.word IN ('turn', 'straight', 'left', 'corner', 'nearby'))
)
ON CONFLICT DO NOTHING;
