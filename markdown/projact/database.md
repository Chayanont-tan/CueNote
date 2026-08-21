# Database Architecture & Schema Specification
## Target Database: PostgreSQL (ปัจจุบันรันบน Docker container `cuenote-postgres`, DB name `mission_note`)

> **สถานะ:** อัปเดต 2026-08-21 — backend ครอบคลุมฟีเจอร์ **Flashcard**
> (tags + vocabulary catalog + flashcards + sentences) และตอนนี้ **Shadowing**
> (Design A: pronunciation score, migration `000009` เพิ่มตาราง `shadowing_attempts`
> — ดูหัวข้อ "Shadowing (Design A)" ด้านล่าง) โค้ด Flashcard รวมอยู่ใน **Go module
> เดียว** `internal/app/flashcard` (โมดูล `vocabulary` เดิมถูกลบทิ้งแล้ว ยุบรวมเข้ามาเป็น
> โค้ดภายในของ `flashcard`) — ดู `structure.md` ส่วน Workspace (Photo Flip
> Card) ถูกตัดออกจาก scope แล้วและลบโมดูลทิ้งไปแล้ว

---

## 🛠️ ตารางที่มีอยู่จริงตอนนี้ (7 ตาราง)

Migration ที่ apply แล้ว: `db/migrations/000001` ถึง `000006`

```sql
-- =============================================================================
-- 1. users — ระบบจัดการบัญชีผู้ใช้งาน
-- =============================================================================
CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name  VARCHAR(100),          -- เพิ่มเข้ามาแล้ว แต่ auth/entity.go ยังไม่ใช้ (ดู Known Gaps #1)
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ
);

-- =============================================================================
-- 2. vocabularies — คลังคำศัพท์หลัก
-- =============================================================================
CREATE TABLE vocabularies (
    id             BIGSERIAL PRIMARY KEY,
    word           VARCHAR(255) NOT NULL,
    part_of_speech VARCHAR(50) NOT NULL,
    meaning_th     TEXT NOT NULL,
    level          VARCHAR(2) NOT NULL DEFAULT 'A1',  -- CEFR: A1-C2 (migration 000006)
    created_at     TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (word, part_of_speech)          -- คำเดียวกันแต่คนละ part_of_speech ถือเป็นคนละคำ
);

-- =============================================================================
-- 3. tags — ชื่อหมวดหมู่/ภารกิจ (เช่น ร้านกาแฟ, food, directions)
-- =============================================================================
CREATE TABLE tags (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- 4. vocabulary_tags — Junction Table (Many-to-Many: 1 คำมีได้หลาย Tag)
-- =============================================================================
CREATE TABLE vocabulary_tags (
    vocabulary_id BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,
    tag_id        BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (vocabulary_id, tag_id)
);

-- =============================================================================
-- 5. flashcards — การ์ดที่ User สร้างขึ้นทีละใบ (1 คำศัพท์ = 1 การ์ด ต่อครั้งที่กด generate)
-- =============================================================================
CREATE EXTENSION IF NOT EXISTS pgcrypto;   -- จำเป็นสำหรับ gen_random_uuid()

CREATE TABLE flashcards (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vocabulary_id BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,
    image_url     TEXT NOT NULL,     -- รูปที่ AI เจนให้ (mock อยู่ตอนนี้ผ่าน AI_MOCK_IMAGES=true)
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
    -- user_sentence / selected_ai_sentence (แบบเดี่ยว) ถูกลบทิ้งแล้วใน migration
    -- 000006 แทนที่ด้วยตาราง flashcard_sentences ด้านล่าง เพื่อรองรับหลายประโยคต่อการ์ด
);

CREATE INDEX idx_flashcards_user_id ON flashcards(user_id);
CREATE INDEX idx_flashcards_vocabulary_id ON flashcards(vocabulary_id);

-- =============================================================================
-- 6. ai_suggested_sentences — ประโยค "ตัวเลือก" ที่ AI เจนให้ ผูกกับ flashcard ใบนั้นๆ
--    (ยังไม่ถูกเลือก/ยืนยัน — แค่โชว์ให้ user กดเลือกหรือไม่ก็ได้)
-- =============================================================================
CREATE TABLE ai_suggested_sentences (
    id            SERIAL PRIMARY KEY,
    flashcard_id  UUID NOT NULL REFERENCES flashcards(id) ON DELETE CASCADE,
    sentence_text TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_suggested_sentences_flashcard_id ON ai_suggested_sentences(flashcard_id);

-- =============================================================================
-- 7. flashcard_sentences — ประโยคที่ "ยืนยันแล้ว" บน flashcard (migration 000006)
--    พิมพ์เอง (source='user') หรือกดเลือกจาก ai_suggested_sentences (source='ai_picked')
--    รองรับหลายประโยคต่อการ์ด ลบทีละประโยคได้ (DELETE /sentences/{id})
-- =============================================================================
CREATE TABLE flashcard_sentences (
    id            SERIAL PRIMARY KEY,
    flashcard_id  UUID NOT NULL REFERENCES flashcards(id) ON DELETE CASCADE,
    sentence_text TEXT NOT NULL,
    source        VARCHAR(10) NOT NULL DEFAULT 'user',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_flashcard_sentences_flashcard_id ON flashcard_sentences(flashcard_id);
```

### ทำไมประโยคผูกกับ `flashcard_id` ไม่ใช่ `vocabulary_id`
คำศัพท์คำเดียวกัน (เช่น "coffee") อาจถูกสร้างเป็น flashcard ซ้ำได้หลายใบ (คนละ user, หรือ user เดิมสร้างซ้ำ) แต่ละใบควรได้ภาพและประโยคตัวอย่างชุดใหม่ของตัวเอง ไม่ใช้ร่วมกันข้ามใบ

### ทำไมแยก `ai_suggested_sentences` กับ `flashcard_sentences` เป็นสองตาราง
`ai_suggested_sentences` คือ "ตัวเลือกที่ AI เสนอ" (แสดงไว้เฉยๆ ไม่มีผลอะไรจนกว่าจะถูกเลือก) ส่วน `flashcard_sentences` คือ "ประโยคที่ยืนยันแล้วจริงบนการ์ด" (พิมพ์เองหรือกดเลือกจากตัวเลือกของ AI ก็ตกลงมาอยู่ในตารางนี้เหมือนกัน) แยกกันเพื่อให้ลบ/แก้ฝั่ง "ยืนยันแล้ว" ได้โดยไม่กระทบประวัติตัวเลือกที่ AI เคยเสนอไว้

### ER Summary
```
users 1───* flashcards *───1 vocabularies *───* tags
                │
                ├──1───* ai_suggested_sentences
                └──1───* flashcard_sentences
```

---

## 📡 API Endpoints (flashcard module)

ทุก endpoint ต้องมี JWT bearer token (ยกเว้น `/auth/*`) และ scope ตาม `user_id` ของเจ้าของเสมอ

| Method | Path | หน้าที่ |
|---|---|---|
| POST | `/api/v1/tags` | สร้าง tag ใหม่ (ยังไม่เจนคำศัพท์) |
| GET | `/api/v1/tags?limit=10` | หน้าแรก: tag ที่เคยสร้าง flashcard ไว้ (จำกัดจำนวน) |
| GET | `/api/v1/tags` | หน้า All Tags: ทั้งหมดไม่จำกัด |
| GET | `/api/v1/tags/{tag_id}` | รายละเอียด tag เดียว |
| POST | `/api/v1/tags/{tag_id}/flashcards/generate` | สั่ง AI เจน flashcard ชุดใหม่ในนี้ (`{limit}`) |
| GET | `/api/v1/tags/{tag_id}/flashcards?level=A1` | รายการ flashcard ใน tag (กรอง CEFR level ได้) |
| GET | `/api/v1/flashcards/{card_id}` | รายละเอียด flashcard ใบเดียว (รวม sentences + ai_suggested_sentences) |
| POST | `/api/v1/flashcards/{card_id}/sentences` | เพิ่มประโยค (พิมพ์เอง หรือกดเลือกจาก AI suggestion) |
| POST | `/api/v1/flashcards/{card_id}/sentences/generate` | ให้ AI เจนประโยคตัวอย่างใหม่ 3 ประโยค (ไม่ auto-add ต้องกด add เอง) |
| DELETE | `/api/v1/sentences/{sentence_id}` | ลบประโยคที่เคยเพิ่มไว้ |

**ยังไม่ implement**: `PUT /flashcards/{card_id}` (แก้ไข flashcard) — ยังไม่ชัดว่าจะแก้ field ไหนบ้างหลังจากที่ `user_sentence`/`selected_ai_sentence` ย้ายไปอยู่ใน `flashcard_sentences` แล้ว รอ requirement ชัดเจนกว่านี้ก่อนค่อย implement

---

## ⚠️ ช่องว่าง/ความไม่ตรงกันที่ควรแก้ต่อ (Known Gaps)

1. **`users.display_name`/`updated_at` มีคอลัมน์แล้วแต่ `auth/entity.go` ยังไม่ใช้** — ไม่มี endpoint ไหนตั้งค่า/อ่านค่า 2 ฟิลด์นี้เลย (ไม่มี register field, ไม่มีหน้า profile) รอ requirement ก่อนค่อยผูกเข้าโค้ด
2. **`db/migrations/000002` ไฟล์ไม่ตรงกับของจริงในบางส่วน** — ไฟล์ migration ที่ track ไว้ระบุ `word VARCHAR(100) UNIQUE NOT NULL` (unique เดี่ยว) แต่ DB จริงเป็น `VARCHAR(255)` และ unique แบบ composite `(word, part_of_speech)` — ยังไม่ reconcile
3. **`flashcards.user_id` เป็น `BIGINT` ไม่ใช่ `UUID`** เพราะ `users.id` จริงเป็น `BIGSERIAL` — คงไว้แบบนี้โดยตั้งใจ (การ migrate `users` ทั้งระบบไปเป็น uuid เป็นงานใหญ่ที่กระทบ `auth` ทั้งโมดูล)
4. **AI ยังไม่รับประกันความแม่นยำของ `level` (CEFR)** — prompt ขอให้ AI ประเมินระดับคำเอง (`internal/pkg/openai/client.go`) ไม่มีการันตี/ตรวจสอบความถูกต้อง คำที่มาจาก seed data เก่าก่อน migration 000006 จะได้ level default เป็น `A1` ทั้งหมด (ไม่ได้จัดระดับจริง)

---

## ✅ Shadowing (Design A) — Implemented (migration `000009`)

เลือกแนวทาง (A) audio pronunciation scoring แล้ว (ดู `product.md` Feature 3) แนวทาง (B) scenario-based Read-Along chat ยังไม่ implement และไม่มีตาราง DB (out of scope งานนี้)

Shadowing **ไม่มีตาราง sentence แยกของตัวเอง** — ใช้ `flashcard_sentences` ที่มีอยู่แล้วเป็นเนื้อหาให้ user อ่าน (ownership เช็คผ่าน join `flashcard_sentences.flashcard_id -> flashcards.user_id`) มีแค่ตารางใหม่ตารางเดียว:

```sql
-- =============================================================================
-- shadowing_attempts — ผลการอ่านตามของ user แต่ละครั้ง (ไม่เก็บไฟล์เสียง)
-- =============================================================================
CREATE TABLE shadowing_attempts (
    id                    BIGSERIAL PRIMARY KEY,
    user_id               BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    flashcard_sentence_id BIGINT NOT NULL REFERENCES flashcard_sentences(id) ON DELETE CASCADE,
    transcript            TEXT NOT NULL,          -- ข้อความที่ได้จาก Whisper
    score                 INT NOT NULL,           -- 0-100, จาก word-level LCS diff
    correct_words         JSONB NOT NULL DEFAULT '[]',
    mispronounced_words   JSONB NOT NULL DEFAULT '[]',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

ไฟล์เสียงที่ user อัปโหลดจะถูกอ่านเข้า memory ส่งไป transcribe แล้วทิ้งทันที — ไม่มีคอลัมน์ `audio_url` และไม่ใช้ `internal/infra/storage` (ยังเป็น stub เหมือนเดิม ไม่มีอะไรอ้างอิงถึงแล้ว)

(Workspace เคยอยู่ในหัวข้อนี้ด้วย แต่ถูกตัดออกจาก scope และลบโมดูลทิ้งแล้ว — ดู `product.md` Feature 2)
