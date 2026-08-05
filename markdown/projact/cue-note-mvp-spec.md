# 🚀 CueNote MVP - Product & User Flow Specification

> **สถานะ: อัปเดต 2026-07-27** — Flow ทั้ง 6 หน้าจอด้านล่างมี clickable mockup
> แล้วที่ `web-test/app-flow-mockup.html` (ไฟล์เดียวจบ, mock data ล้วนๆ,
> ไม่เรียก backend จริง) ฝั่ง backend/DB ตอนนี้ทำตามทันแล้วส่วนใหญ่ (ดู
> `product.md` Feature 1 และ `database.md`):
> - ✅ CEFR `level` ต่อคำ — มีคอลัมน์จริงแล้ว (migration 000006)
> - ✅ แต่งได้หลายประโยคต่อคำ — ตาราง `flashcard_sentences` รองรับแล้ว
>   (`POST/DELETE .../sentences`)
> - ⏳ การเพิ่มคำศัพท์เข้า tag เดิมแบบ "ไม่ซ้ำคำเก่า" — endpoint
>   `POST /tags/{id}/flashcards/generate` เรียกซ้ำได้ แต่ยังไม่การันตีว่าจะ
>   ได้คำใหม่ที่ไม่เคยมีมาก่อนเสมอ
> - ❌ Shadowing แบบ scenario-based — ยังไม่มี backend/DB เลย (ดูหัวข้อ 4
>   ด้านล่างสำหรับ schema ที่ยังต้องสร้าง)

## 1. Executive Summary & Core Concept
* **Project Name:** CueNote
* **Concept:** Productivity Application สำหรับฝึกทักษะภาษาอังกฤษเพื่อใช้ในสถานการณ์จริง (Scenario-based Learning) ผ่านระบบ **Dynamic Tag Generation**, **Vocab Context Cards**, และ **Read-Along Shadowing Chat**
* **Target User:** Solo Developer / Self-Learners ที่ต้องการเก่งภาษาเพื่อภารกิจเฉพาะหน้า
* **Development Strategy:** Solo Full-Stack Architecture (เน้น Lean MVP, ไม่เก็บไฟล์เสียง User, ไม่ใช้ Image Gen API เพื่อลด Latency และ Cost)

---

## 2. Core Architecture & Navigation System

### 🧭 Global Navigation Bar (Bottom Tabs - 4 Main Tabs)
1. 🏠 **Home:** สร้าง Tag ใหม่ (In-place Generation), แสดงรายการ Tag ทั้งหมด และการ์ดคำศัพท์ล่าสุด
2. 🏷️ **All Cards / Tags:** คลังรวม Tag และการ์ดคำศัพท์ทั้งหมด สามารถกดเข้าไปดูรายละเอียดและแต่งประโยคได้
3. 🎙️ **Shadowing:** คลังรวมบทสนทนาโต้ตอบตามสถานการณ์ (Scenarios) สำหรับฝึก Read-Along
4. ⚙️ **Settings:** การตั้งค่าบัญชีผู้ใช้และค่าเริ่มต้นของแอป

---

## 3. Comprehensive Screen Lists & User Flows

### 📱 Screen 1: Authentication (Register / Login Page)
* **UI Elements:**
  * Input Fields: `Email`, `Password`, `Confirm Password` (เฉพาะ Register)
  * Action Buttons: `Login` / `Register`, Social Auth (Optional)
* **User Flow:**
  * ยูสเซอร์เข้าสู่ระบบสำเร็จ ➔ Navigate ไปยัง **[Home Page]**

---

### 📱 Screen 2: Home Page (Single-Page Tag Creation & Live Update)
* **UI Elements:**
  * **Onboarding Card:** การ์ดต้อนรับและแนะนำวิธีใช้งาน
  * **Tag Input Bar:** ช่องพิมพ์ชื่อภารกิจ/Tag (เช่น *"ขายของงานอีเว้นท์"*, *"ร้านกาแฟ"*) + ปุ่มส่ง `Submit`
  * **Tag Grid Collection:** พื้นที่แสดงการ์ด Tag ทั้งหมด (มี Tag Name, Count, Progress Badge)
  * **Skeleton / Loading Card:** การ์ดสีขาวโปร่งแสงแสดงสถานะระหว่าง AI กำลังเจนข้อมูล
  * **Recent Cards Carousel:** การ์ดคำศัพท์ล่าสุดที่เพิ่งสร้าง/แก้ไข
* **User Actions & System Logic:**
  1. ยูสเซอร์พิมพ์ชื่อ Tag แล้วกดส่ง `Submit`
  2. หน้าจอ **ไม่เปลี่ยนหน้า (Stay on Home)** ➔ แทรก **Skeleton Card** ลง Grid ทันที
  3. Go Backend ส่ง Prompt หา AI เพื่อเจนคำศัพท์ 5 คำประจำ Tag
  4. เมื่อ AI ตอบกลับ ➔ อัปเดตการ์ดใหม่แทนที่ Skeleton Card แบบ Real-time (In-place Update)

---

### 📱 Screen 3: Vocab Card List (By Tag)
* **UI Elements:**
  * **Tag Header:** ชื่อ Tag และสถิติคำศัพท์ (เช่น `☕ ร้านกาแฟ (5/5 Words)`)
  * **Level Filter Badges:** ปุ่มกรองระดับภาษาตามมาตรฐาน CEFR (`All`, `A1-A2`, `B1-B2`, `C1-C2`)
  * **Vocab Cards Grid:** การ์ดแสดงรายการคำศัพท์
    * `Word` (คำศัพท์) + `Part of Speech` (ชนิดของคำ)
    * `Meaning TH` (คำแปลไทย) + `Level Badge` (A1-C1)
  * **Action Button:** ปุ่ม `+ Create / Edit Sentence` บนการ์ดแต่ละใบ
* **User Actions & Navigation:**
  * กดที่การ์ดคำศัพท์ ➔ เปิดไปยัง **[Create / Edit Sentence Screen]**

---

### 📱 Screen 4: Create / Edit Sentence Screen
* **UI Elements:**
  * Target Word Display & Definition
  * Textarea สำหรับพิมพ์/แต่งประโยคด้วยตัวเอง
  * Button: `✨ AI Generate 3 Sentences` (ให้ AI เจนประโยคตัวอย่าง 3 แบบ)
  * Button: `💾 Save Card`
* **User Actions & Data Flow:**
  1. ยูสเซอร์แต่งประโยคเอง หรือกดให้ AI เจนประโยค 3 แบบแล้วเลือกประโยคที่ชอบ
  2. กดปุ่ม `Save Card` ➔ บันทึกข้อมูลลง DB (ผูกกับ `vocabulary_id`)
  3. แสดง Pop-up Toast: *"Saved Successfully"*
  4. ระบบ Auto Navigate ย้อนกลับไปหน้า **[Vocab Card List]** ทันที

---

### 📱 Screen 5: All Shadowing Card Page (Scenario Hub)
* **UI Elements:**
  * **Tag Filter Chips:** แถบเลือกกรองสถานการณ์ตาม Tag (`All`, `ร้านกาแฟ`, `ขายของ`)
  * **Scenario Cards Grid:** การ์ดแสดงบทสนทนาสถานการณ์
    * `Scenario Title` (เช่น *"สถานการณ์ 1: เดินไปสั่งกาแฟและระบุความหวาน"*)
    * `Tag Badge` + `Dialogue Count` (เช่น `4 Dialogues`)
    * `Status Badge` (เช่น `Read-Along Ready`)
    * Button: `▶ Start Practice`
  * **Locked Card State:** การ์ดที่ยังไม่ปลดล็อกเนื่องจากคำศัพท์ใน Tag ยังไม่ครบตามเกณฑ์
* **Unlock Conditions Logic:**
  * **Scenario 1:** ปลดล็อกเมื่อสะสมคำศัพท์ใน Tag ครบ **5 คำ**
  * **Scenario 2:** ปลดล็อกเมื่อสะสมคำศัพท์ใน Tag ครบ **10 คำ**
  * **Scenario 3:** ปลดล็อกเมื่อสะสมคำศัพท์ใน Tag ครบ **15 คำ**
* **User Actions:**
  * หากเงื่อนไขครบ ➔ กด `+ Create Scenario` ให้ AI เจนบทสนทนาใหม่จากคำศัพท์ที่มี
  * กดปุ่ม `▶ Start Practice` ➔ Navigate ไปยัง **[Shadowing Practice Screen]**

---

### 📱 Screen 6: Shadowing Practice Screen (Chat-Style Read-Along Mode)
* **UI Elements:**
  * **Header:** ชื่อ Scenario + ปุ่ม Back `<`
  * **Chat Dialogue Container:** กล่องแสดงบทสนทนาโต้ตอบแบบ Chat
    * **AI Line (ฝั่งซ้าย):** บทพูดของคู่สนทนา (เช่น Barista / Customer)
    * **User Line (ฝั่งขวา):** บทพูดที่ยูสเซอร์ต้องฝึกอ่านตาม (เน้นไฮไลต์คำศัพท์เป้าหมาย)
  * **Active Bubble Highlight:** กล่องแชทเปลี่ยนสีเน้นตามประโยคที่เสียง AI กำลังอ่าน
  * **Control Panel (Bottom Bar):**
    * ปุ่ม `Play / Pause`
    * ปุ่มปรับความเร็วเสียง `Speed (x0.8, x1.0)`
    * ปุ่ม `Translate Toggle` (เปิด/ปิด คำแปลไทยใต้กล่องแชท)
    * ปุ่ม `Replay` (เมื่ออ่านจบสคริปต์)
* **User Actions & Auto-Flow:**
  1. กด `Play` ➔ ระบบเล่นเสียง AI อ่านบทสนทนาฝั่ง AI
  2. เสียง AI สลับมาอ่านบทสนทนาฝั่ง User (ใช้จังหวะความเร็วเหมาะสม)
  3. ยูสเซอร์อ่านออกเสียงพูดตามบทตัวเองไปพร้อมกับเสียง AI (Read-Along)
  4. หน้าจอทำการ **Auto-scroll** ตามประโยคที่กำลังอ่านไปเรื่อยๆ จนจบสคริปต์

---

## 4. Entity-Relationship Data Structure (Draft for DB Design)

Schema ปัจจุบัน (`../database.md`) รองรับแค่ Screen 1-4 บางส่วน (auth, สร้าง
flashcard ต่อคำ, ประโยคเดียวต่อ flashcard) ยังขาดสิ่งเหล่านี้ถ้าจะทำตาม flow
เต็มในเอกสารนี้:

### 4.1 CEFR Level ต่อคำ (Screen 3: Level Filter Badges) — ✅ ทำแล้ว (2026-07-27)
เพิ่มคอลัมน์ `vocabularies.level VARCHAR(2) NOT NULL DEFAULT 'A1'` แล้วใน
migration `000006` และปรับ prompt ของ `openai.GenerateVocabulariesByTag`
ให้ AI คืน `level` มาด้วยตอนเจนคำใหม่แล้ว — ดู schema จริงใน `database.md`

### 4.2 แต่งประโยคได้หลายประโยคต่อคำ (Screen 4) — ✅ ทำแล้ว (2026-07-27)
`flashcards.user_sentence`/`selected_ai_sentence` ถูก drop ทิ้งแล้ว แทนที่
ด้วยตาราง `flashcard_sentences` (migration `000006`) — เพิ่มประโยคผ่าน
`POST /flashcards/{card_id}/sentences`, ลบผ่าน `DELETE /sentences/{id}`

### 4.3 เพิ่มคำศัพท์เข้า Tag เดิมทีหลัง (Screen 2/3: "+ เพิ่มคำศัพท์") — ⏳ ยังไม่การันตีไม่ซ้ำคำเก่า
มี endpoint แล้ว (`POST /tags/{tag_id}/flashcards/generate` เรียกซ้ำได้ตาม
`database.md`) แต่ตอนนี้ logic การ์ดใหม่ยังเป็นการสุ่มจาก `vocabularies` ที่มี
อยู่ในหมวดนั้นอยู่ก่อน (`GetRandomVocabByTagID`) ยังไม่ได้กันไม่ให้ได้คำซ้ำกับ
ที่ user เคยสร้าง flashcard ไปแล้ว — ต้องปรับ query ให้ตัดคำที่มี flashcard
ของ user คนนี้อยู่แล้วออกก่อนค่อยสุ่ม ถ้าจะให้ "เพิ่มคำศัพท์" ได้คำใหม่จริงๆ ทุกครั้ง

### 4.4 Shadowing แบบ Scenario-based (Screen 5-6)
ตารางใหม่ทั้งหมด (ยังไม่มีเลยตอนนี้):
```sql
CREATE TABLE shadowing_scenarios (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tag_id        BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    scenario_num  SMALLINT NOT NULL,              -- 1, 2, 3
    title         VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tag_id, scenario_num)
);

CREATE TABLE shadowing_dialogue_lines (
    id            SERIAL PRIMARY KEY,
    scenario_id   UUID NOT NULL REFERENCES shadowing_scenarios(id) ON DELETE CASCADE,
    line_order    SMALLINT NOT NULL,
    speaker       VARCHAR(10) NOT NULL,            -- 'ai' หรือ 'user'
    text_en       TEXT NOT NULL,
    text_th       TEXT NOT NULL,
    source_flashcard_id UUID REFERENCES flashcards(id)  -- ถ้าบรรทัดนี้ดึงมาจากประโยคที่ user เคยแต่งเอง
);
```
Unlock logic (5/10/15 คำต่อ tag) คำนวณจาก `COUNT(*) FROM vocabulary_tags
WHERE tag_id = ?` ไม่ต้องเก็บ state แยก — แค่เช็คว่ามีแถวใน
`shadowing_scenarios` สำหรับ (tag_id, scenario_num) นั้นหรือยังเพื่อรู้ว่า
"unlocked แต่ยังไม่ create" กับ "created แล้ว" ต่างกันตรงมีแถวหรือไม่