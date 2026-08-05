# Product Specification
## Project Name: CueNote (Core MVP)

---

## 1. Product Concept & Vision
**CueNote** คือ Productivity Application สำหรับฝึกฝนทักษะภาษาอังกฤษ (เน้นการพูด Shadowing และแต่งประโยคจากบริบทจริง) ในรูปแบบ **"Notion-style Workspace"**

เกิดมาเพื่อแก้ Pain Point ของคนที่ต้องการเก่งภาษาเพื่อไป **"ลุยสถานการณ์หรือภารกิจเฉพาะหน้าในชีวิตจริง"** (เช่น เตรียมตัวคุมบูธ/แนะนำสินค้าในอีเวนต์, เตรียมสัมภาษณ์งาน หรือเตรียมตัวไปเรียนต่อ) โดยไม่เน้นการเรียนตามหลักสูตรทั่วไป แต่เน้นการเรียนรู้ด้วยตัวเองแบบคุมเวลาและเป้าหมายได้เองโดยไม่ยืดเยื้อ

### 🎯 Target Audience & Scale
* **Solo Developer / Tech Professionals / Self-Learners:** กลุ่มคนที่ต้องการพัฒนาภาษาอังกฤษแบบมีเป้าหมาย นำไปใช้พูดคุยได้จริงในสถานการณ์นั้นๆ และชอบจัดตารางเรียนรู้ด้วยตัวเอง
* **Target Scale:** รองรับผู้ใช้งานเริ่มต้น 3,000 คน+ บนโครงสร้างระบบที่เสถียร ประหยัดต้นทุน และขยายต่อได้ง่าย

---

## 2. Core MVP Feature Specifications

### Feature 1: Tag-Based Vocabulary & Flashcard Generator — ✅ Implemented
* **สถานะ:** Backend + Database ทำงานจบ flow แล้ว รายละเอียด endpoint ทั้งหมดดู `database.md` หัวข้อ "API Endpoints", รายละเอียด flow การเจนดู `../fureture/vocabulary/random_word.md`
* **รวมเป็น 1 โมดูล** (2026-07-27, อัปเดตจากที่เคยแยก `vocabulary`/`flashcard` ไว้ก่อนหน้า): โค้ดทั้งหมดอยู่ใน `internal/app/flashcard` — โมดูล `vocabulary` เดิมถูกลบทิ้งแล้ว ยุบรวมเป็นโค้ดภายในของ `flashcard` เพราะ API ใหม่ทำให้ tag/vocabulary/flashcard/sentence กลายเป็น resource เดียวกันที่มี lifecycle ผูกกันแน่นแล้ว (สร้าง tag → เจน flashcard ในนั้น → แต่งประโยคในการ์ด)
* **UX/UI Concept:** ผู้ใช้พิมพ์ชื่อ Tag/ภารกิจ (เช่น `ร้านกาแฟ`, `ขายของงานอีเว้นท์`) ระบบสร้าง tag แล้วค้นหาคำศัพท์ที่เคยเจนไว้ในหมวดนั้นก่อน ถ้าไม่พบจะยิงหา AI (`gpt-4o-mini` ผ่าน Groq) เพื่อเจนคำศัพท์ใหม่ พร้อมสร้าง **flashcard** ให้ทันที — 1 คำศัพท์ = 1 flashcard ที่มีทั้งรูปประกอบและประโยคตัวอย่างของตัวเอง แต่งประโยคเพิ่มได้หลายประโยคต่อการ์ด (พิมพ์เอง หรือกดเลือกจากประโยคที่ AI เจนให้)
* **Data Flow (เปลี่ยนเป็น multi-step ตาม API spec ใหม่ 2026-07-27):**
  1. Client เรียก `POST /tags {name}` สร้าง tag (idempotent — เรียกซ้ำชื่อเดิมไม่พัง)
  2. Client เรียก `POST /tags/{tag_id}/flashcards/generate {limit}` — ค้นใน `vocabularies`/`vocabulary_tags` ก่อน (cache) ถ้าไม่พบยิง AI เจนคำศัพท์ใหม่ (พร้อม CEFR `level`) แล้วบันทึกลง DB
  3. ต่อคำ: เจนรูป (ผ่าน `openai.GenerateImage` — **ตอนนี้ mock อยู่** ด้วย `AI_MOCK_IMAGES=true`) และเจนประโยคตัวอย่างใหม่ (`openai.GenerateSentences`) บันทึกเป็นแถวใหม่ใน `flashcards` + `ai_suggested_sentences` ผูกกับ user ID จาก JWT จริง
  4. หน้ารายละเอียดการ์ด: `GET /flashcards/{card_id}`, เพิ่มประโยค `POST /flashcards/{card_id}/sentences`, ให้ AI เจนประโยคเพิ่ม `POST /flashcards/{card_id}/sentences/generate`, ลบประโยค `DELETE /sentences/{sentence_id}`
* **ยังไม่ทำ:** `PUT /flashcards/{card_id}` (แก้ไข flashcard — ยังไม่ชัดว่าจะแก้ field ไหน), การเพิ่มคำศัพท์ใหม่เข้า tag เดิมแบบ "รับประกันไม่ซ้ำคำเก่า" (ตอนนี้ generate ซ้ำอาจได้คำเดิมที่เคยมีอยู่แล้วแบบสุ่ม)

### Feature 2: Photo Block & Voice-to-Text Flip Card Workspace — ❌ Removed (2026-07-27)
เดิมมีโครง Go module (`internal/app/workspace`) สแกฟโฟลด์ไว้แต่ไม่เคยมีตาราง DB รองรับเลย และ storage client ก็ยังเป็น stub — ตัดสินใจแล้วว่าไม่อยู่ใน scope ต่อไป (ดีไซน์ UI ล่าสุดใน `cue-note-mvp-spec.md` ไม่ได้พูดถึง Photo Flip Card เลย เน้น Tag→Vocab→Sentence→Shadowing แทน) **ลบโมดูลทิ้งแล้ว** ถ้าจะกลับมาทำฟีเจอร์นี้ในอนาคตต้องออกแบบ/เขียนใหม่ทั้งหมด

### Feature 3: Shadowing — 📝 Two competing designs, need a decision before implementing
มีเอกสารดีไซน์ 2 ชุดที่ไม่ตรงกัน ต้องเลือกก่อนเริ่ม implement backend:

**(A) แนวทางเดิม — Pronunciation Score แบบ Spotify Lyrics**
* **สถานะ:** มีโครง Go module (`internal/app/shadowing`) กับ route `GET /shadowing/sentences/:id`, `POST /shadowing/attempts` แล้ว แต่คะแนนถูก hardcode เป็น `0`, ยังไม่มีตาราง DB, ยังไม่ต่อ TTS/Whisper จริง
* **UX/UI Concept:** ระบบแสดงผลประโยคที่ถูกต้องในรูปแบบเนื้อร้องคาราโอเกะ (Spotify Lyrics) โดยจะมี AI กดอ่านออกเสียงประโยคนั้นให้ฟัง และตัวหนังสือบนหน้าจอจะไฮไลต์วิ่งตามทีละคำตามจังหวะเสียงพูดจริง จากนั้นผู้ใช้กดปุ่มอัดเสียงเพื่อทำ Shadowing (พูดตาม) เมื่อพูดจบระบบจะแสดงการคำนวณคะแนนความถูกต้อง (0-100) พร้อมไฮไลต์คำที่ออกเสียงชัดเป็น **สีเขียว** และคำที่ออกเสียงเพี้ยนเป็น **สีแดง**
* **Data Flow:** หน้าบ้านใช้ข้อมูลคำและตำแหน่งเวลา (Timestamps) ในการขยับไฮไลต์ตัวหนังสือตามไฟล์เสียง AI -> เมื่อผู้ใช้อัดเสียง Shadowing เสร็จ ระบบส่งไฟล์เสียงไปแปลงเป็นข้อความและทำ Diff Match เปรียบเทียบกับต้นฉบับเพื่อคิดคะแนนส่งกลับมาโชว์ที่หน้าบ้าน

**(B) แนวทางใหม่ — Scenario-based Read-Along Chat** (ตาม `cue-note-mvp-spec.md` และ UI mockup ล่าสุด `web-test/app-flow-mockup.html`)
* **สถานะ:** มีแค่ UI mockup (คลิกดูได้จริง ไม่มี backend) ยังไม่มีตาราง DB เลย
* **UX/UI Concept:** ไม่มีการอัดเสียง/ให้คะแนนการออกเสียง — แทนที่ด้วยบทสนทนาสถานการณ์ (Scenario) แบบ Chat ที่ปลดล็อกทีละสถานการณ์ตามจำนวนคำศัพท์สะสมใน Tag นั้นๆ (5/10/15 คำ) ผู้ใช้กด "สร้างสถานการณ์" ให้ AI เจนบทสนทนาจากคำศัพท์ + ประโยคที่เคยแต่งเอง แล้วกด Play ให้ไฮไลต์วิ่งตามทีละบทพร้อม auto-scroll (ไม่มีการประเมินการออกเสียง)
* **Data Flow (ยังไม่ implement):** ต้องมีตารางเก็บ scenario ต่อ tag (เช่น `shadowing_scenarios`: tag_id, scenario_number, สถานะ created/locked) และตารางเก็บบทสนทนาแต่ละบรรทัด (เช่น `shadowing_dialogue_lines`: scenario_id, speaker, text_en, text_th, order)

**⚠️ ต้องตัดสินใจก่อน implement:** จะทำ (A), (B), หรือทำทั้งคู่แบบคนละ endpoint? การเลือกกระทบ DB schema และ effort ต่างกันมาก — (A) ต้องมี audio pipeline (Whisper transcription + diff scoring + TTS), (B) ไม่ต้องจัดการเสียงเลย เน้น content generation + unlock logic เท่านั้น
