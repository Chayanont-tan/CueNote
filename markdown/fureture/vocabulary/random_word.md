# 💡 Feature Spec: MissionNote - Smart Tag-Based Vocabulary & Flashcard Generator

> **สถานะ: ✅ Implemented** (อัปเดต 2026-07-27) — flow และ DBML ด้านล่างตรงกับ
> backend ที่ implement จริงแล้วเกือบทั้งหมด แต่ endpoint จริงเปลี่ยนเป็น
> multi-step แล้ว (`POST /tags` → `POST /tags/{tag_id}/flashcards/generate`
> — ดู `database.md` หัวข้อ "API Endpoints") โค้ดทั้งหมดรวมอยู่ใน module
> เดียว `internal/app/flashcard` แล้ว (ไม่มี module `vocabulary` แยกอีกต่อไป)
> ยกเว้น: รูปภาพเป็น mock อยู่ (`AI_MOCK_IMAGES=true`, ยังไม่เรียก DALL-E จริง),
> `users.id` และ `flashcards.user_id` เป็น `BIGINT` ไม่ใช่ `uuid` ตาม DBML นี้
> (เพราะ `auth` module มีอยู่ก่อนแล้วด้วย bigserial), Local SQLite caching
> และการอ้างอิง NGSL/WordNet ยังไม่ได้ทำ (ยังใช้ AI เจนสดทุกคำใหม่) — ดู schema
> จริงที่ใช้งานอยู่ตอนนี้ใน `../../projact/database.md`

## 📌 Overview
ระบบจัดการคลังคำศัพท์และสร้าง Flashcard การเรียนรู้แบบโต้ตอบ โดยทำงานร่วมกันระหว่าง **PostgreSQL Database** (คลังหลักเพื่อความเร็ว ประหยัด และทำ Local Caching) กับ **OpenAI (`gpt-4o-mini`)** (โรงงานผลิตศัพท์ ประโยคตัวอย่าง และ Tag ใหม่สดๆ เมื่อค้นใน DB ไม่พบ)

---

## 🛠️ Key Concepts & Flow Chart

### 1. Hybrid Fetching & Caching Strategy
1. **Search DB First:** เมื่อผู้ใช้ขอคำศัพท์จาก Tag ใดๆ ระบบจะไปค้นหาใน PostgreSQL (และ SQLite ในเครื่อง) ก่อนเสมอ
2. **AI Fallback:** หากไม่พบ Tag นั้นใน DB (เป็น Tag ใหม่สดๆ เช่น "ขายของงานอีเว้น") ระบบจะยิงหา OpenAI API เพื่อสร้างคำศัพท์ + ประโยคตัวอย่างใหม่
3. **Auto Save (Cloud & Local):** คำศัพท์และ Tag ใหม่จาก AI จะถูกบันทึกลง Database ทันที เพื่อให้ผู้ใช้คนถัดไปเรียกใช้ได้ฟรีและเร็วระดับ Millisecond รวมถึงแอบเซฟลง SQLite ในเครื่องสำหรับใช้งานแบบ Offline

```text
[User Request (Tag, Limit)]
         │
         ▼
 ┌───────────────┐
 │ Search in DB  │
 └───────┬───────┘
         │
    ┌────┴────────────┐
    ▼                 ▼
[Found]          [Not Found]
    │                 │
    │                 ▼
    │         ┌───────────────┐
    │         │ Call OpenAI   │ (gpt-4o-mini)
    │         └───────┬───────┘
    │                 │
    │                 ▼
    │         ┌───────────────┐
    │         │  Save to DB   │ (Auto Caching)
    │         └───────┬───────┘
    │                 │
    └────────┬────────┘
             ▼
   [Return JSON Output]
```

---

## 🔄 End-to-End User & System Flow (ขั้นตอนการทำงานอย่างละเอียด)

### Step 1: เลือก หรือ สร้าง Tag ใหม่ (Tag Selection / Generation)
**UX Action:** ผู้ใช้เลือก Tag ที่มีอยู่แล้ว (เช่น ร้านกาแฟ, น้ำดื่ม) หรือ พิมพ์สร้าง Tag ใหม่ขึ้นมาเอง (เช่น ขายของงานอีเว้น)

**System Logic:**
- Backend ทำ String Normalization (TrimSpace, ToLower) เพื่อป้องกัน Tag ซ้ำจากเว้นวรรค
- เช็กใน DB ว่ามี Tag นี้และคำศัพท์ผูกอยู่หรือไม่
- ถ้าเจอ: สุ่มคำศัพท์ตามจำนวนที่กำหนด (เช่น 4 คำ) ส่งกลับทันที
- ถ้าไม่เจอ: ยิงหา OpenAI (gpt-4o-mini) บังคับคาย JSON สเปก: Tag Name + Vocabularies + AI Suggested Sentences

### Step 2: สุ่มคำศัพท์ และ แสดงผล (Vocab Randomization)
**UX Action:** ระบบแสดงคำศัพท์ภาษาอังกฤษ 4 คำพร้อมคำแปลไทยตาม Tag ที่เลือก

**System Logic:**
- ใช้ SQL JOIN ระหว่าง vocabularies, vocabulary_tags, และ tags สุ่มคำศัพท์แบบ ORDER BY RANDOM() LIMIT 4

### Step 3: ปรับแต่งการ์ดทีละใบ (Custom Flashcard Workspace)
**UX Action:** ผู้ใช้กดแต่งการ์ดทีละใบจากคำศัพท์ที่ AI สุ่มมาให้
- ระบบแสดงภาพถ่าย/ภาพประกอบประจำการ์ด
- แสดงประโยคตัวอย่าง 3-4 ประโยคที่ AI เจนให้ (ช่วยกรณีผู้ใช้คิดไม่ออก)
- ผู้ใช้สามารถเลือกพิมพ์ประโยคเอง, กดพูดผ่านไมค์ (Voice-to-Text), หรือจิ้มเลือกประโยคตัวอย่างจาก AI ก็ได้

**System Logic:**
- สร้าง Record ใหม่ในตาราง flashcards ผูก user_id, vocabulary_id, image_url และข้อความที่ผู้ใช้เลือก/แต่ง
- บันทึกประโยคตัวอย่าง 3-4 ประโยคลงตาราง ai_suggested_sentences เพื่อให้ดึงกลับมาดูย้อนหลังได้เสมอ

---

## 🗄️ Database Architecture & DBML Specification
ดีไซน์ตารางแบบ Normalized (Many-to-Many) รองรับกรณี 1 คำศัพท์ไปอยู่ในหลาย Tag ได้ (เช่น คำว่า coffee อยู่ทั้งหมวด ร้านกาแฟ และ น้ำดื่ม)

```dbml
// =============================================================================
// DBML Specification for MissionNote
// Paste this code into https://dbdiagram.io
// =============================================================================

// 1. Table: users (ระบบจัดการบัญชีผู้ใช้งาน)
Table users {
  id uuid [pk, default: `gen_random_uuid()`]
  email varchar(255) [unique, not null]
  password_hash varchar(255) [not null]
  display_name varchar(100)
  created_at timestamp [default: `now()`]
}

// 2. Table: tags (คลังชื่อ Tag ทั้งหมด)
Table tags {
  id integer [pk, increment]
  name varchar(50) [unique, not null, note: 'เช่น cafe, drink, ขายของงานอีเว้น']
  created_at timestamp [default: `now()`]
}

// 3. Table: vocabularies (คลังคำศัพท์หลัก)
Table vocabularies {
  id integer [pk, increment]
  word varchar(100) [unique, not null]
  part_of_speech varchar(50) [not null]
  meaning_th varchar(255) [not null]
  created_at timestamp [default: `now()`]
}

// 4. Table: vocabulary_tags (Junction Table: 1 ศัพท์ มีได้หลาย Tag)
Table vocabulary_tags {
  vocabulary_id integer [ref: > vocabularies.id, note: 'Cascade Delete']
  tag_id integer [ref: > tags.id, note: 'Cascade Delete']

  Indexes {
    (vocabulary_id, tag_id) [pk]
    tag_id [name: 'idx_vocab_tags_tag_id']
  }
}

// 5. Table: flashcards (Flashcard แต่ละใบที่ User สร้าง)
Table flashcards {
  id uuid [pk, default: `gen_random_uuid()`]
  user_id uuid [not null, ref: > users.id]
  vocabulary_id integer [not null, ref: > vocabularies.id]

  image_url text [not null, note: 'ภาพถ่าย / ภาพประกอบ AI']
  user_sentence text [note: 'ประโยคที่ User พิมพ์หรือพูด Voice-to-Text เอง']
  selected_ai_sentence text [note: 'กรณี User ไม่แต่งเอง แต่กดเลือกจากประโยค AI']

  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]

  Indexes {
    user_id [name: 'idx_flashcards_user_id']
  }
}

// 6. Table: ai_suggested_sentences (ประโยคตัวอย่าง 3-4 ประโยคที่ AI เจนให้เลือก)
Table ai_suggested_sentences {
  id integer [pk, increment]
  flashcard_id uuid [not null, ref: > flashcards.id, note: 'Cascade Delete']
  sentence_text text [not null, note: 'ประโยคตัวอย่างที่ AI เจน']
  created_at timestamp [default: `now()`]
}
```

> **หมายเหตุ:** DBML ด้านบนคือดีไซน์ตั้งต้น — schema จริงที่รันอยู่ตอนนี้ต่างจากนี้
> เล็กน้อย (`users.id`/`flashcards.user_id` เป็น `BIGINT` ไม่ใช่ `uuid`) ดู
> ของจริงที่ `../../projact/database.md`

---

## 📜 Example Database SQL Queries

### 1. สุ่มคำศัพท์ 4 คำตาม Tag ที่เลือก (PostgreSQL)
```sql
SELECT
    w.id,
    w.word,
    w.part_of_speech,
    w.meaning_th,
    t.name AS tag_name
FROM vocabularies w
JOIN vocabulary_tags wt ON w.id = wt.vocabulary_id
JOIN tags t ON t.id = wt.tag_id
WHERE t.name = 'ร้านกาแฟ'
ORDER BY RANDOM()
LIMIT 4;
```

### 2. Transaction บันทึกข้อมูลเมื่อเจอ Tag ใหม่จาก AI
```sql
BEGIN;

-- 1. เพิ่ม Tag ใหม่ (ถ้ามีแล้วให้ข้าม)
INSERT INTO tags (name) VALUES ('ขายของงานอีเว้น')
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id; -- สมมติได้ tag_id = 10

-- 2. เพิ่มคำศัพท์ใหม่
INSERT INTO vocabularies (word, part_of_speech, meaning_th)
VALUES ('booth', 'noun', 'บูธแสดงสินค้า')
ON CONFLICT (word) DO UPDATE SET word = EXCLUDED.word
RETURNING id; -- สมมติได้ vocabulary_id = 101

-- 3. จับคู่ Word กับ Tag
INSERT INTO vocabulary_tags (vocabulary_id, tag_id)
VALUES (101, 10)
ON CONFLICT DO NOTHING;

COMMIT;
```

---

## ⚖️ Legal & Copyright Considerations (การจัดการคลังศัพท์ถูกกฎหมาย)
หลีกเลี่ยงการก๊อปปี้ Oxford 3000 ทั้งชุดตรงๆ เพื่อป้องกันปัญหาการละเมิดลิขสิทธิ์ประเภทจัดหมวดหมู่ (Compilation Copyright)

**ทางออกคลังศัพท์ตั้งต้น (Initial Pre-load Data):**
- ใช้คลังศัพท์เปิดฟรีเพื่อการค้า เช่น NGSL (New General Service List) หรือ WordNet
- ใช้ AI เจนชุดคำศัพท์พื้นฐาน A1-B2 แยกตาม Tag ตั้งต้นไว้ ~1,000–2,000 คำ บันทึกลง SQLite ในเครื่องตั้งแต่ติดตั้งแอป
- Dynamic Scaling: ใช้ gpt-4o-mini เจนศัพท์ใหม่สดๆ เมื่อเจอ Tag ใหม่ ซึ่งไม่มีปัญหาเรื่องลิขสิทธิ์ผูกขาด

---

## 💰 Cost & Performance Optimization Summary
- **AI Cost:** ใช้ gpt-4o-mini + Prompt ขนาดสั้น บังคับ JSON Output — ค่าใช้จ่ายเฉลี่ย ~0.005 บาท ($0.00014 USD) ต่อ 1 Tag ใหม่
- **Economy of Scale:** ยิ่งมีผู้ใช้เยอะ และมีการ Caching คลังศัพท์ลง DB มากขึ้น ค่าใช้จ่าย AI ต่อหัวจะวิ่งเข้าใกล้ 0 บาท
- **Offline Resilience:** ข้อมูลศัพท์และ Tag ที่เคยถูกดึง/เจนแล้ว จะถูกแอบ Caching ลง SQLite ในมือถือ ช่วยให้แอปทำงานแบบ Offline ได้ไหลลื่น
