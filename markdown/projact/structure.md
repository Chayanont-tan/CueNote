# Project Directory & File Structure (Advanced Layout)
## Backend Architecture: Standard Go Project Layout (Feature-Based Modules)

โครงสร้างนี้จัดหมวดหมู่โค้ดตามฟังก์ชันธุรกิจ (Feature-Based) และแยกโครงสร้างทางเทคนิค (Infrastructure/Third-party SDKs) ออกจาก Business Logic อย่างเด็ดขาด เพื่อให้ง่ายต่อการขยายระบบและการเชื่อมต่อ Dependencies

---

## 📂 Backend Directory Tree

```text
mission-note/
├── cmd/
│   └── api/
│       └── main.go                 # Setup Gin, Wire Route Groups, Entry Point
│
├── internal/
│   ├── app/                        # Core Business Logic (Feature Modules)
│   │   ├── flashcard/              # Feature 1: tags + vocabulary catalog + flashcards + sentences
│   │   │   │                        # (vocabulary module was folded in here 2026-07-27 — see product.md)
│   │   │   ├── flashcard.go        # Module Initializer — 3 route groups: /tags, /flashcards, /sentences
│   │   │   ├── interfaces.go       # Service/Repository Interfaces
│   │   │   ├── entity.go           # Database Models (Tag, Vocabulary, Flashcard, *Sentence)
│   │   │   ├── dto.go              # Request/Response Data Transfer Objects
│   │   │   ├── service.go          # Core Business Logic
│   │   │   ├── http_handler.go     # Gin HTTP Handlers & Controllers
│   │   │   └── pg_repository.go    # PostgreSQL Database Queries
│   │   │
│   │   ├── shadowing/              # Feature 2: Shadowing (design not yet chosen — see product.md)
│   │   │   ├── shadowing.go
│   │   │   ├── interfaces.go
│   │   │   ├── entity.go
│   │   │   ├── dto.go
│   │   │   ├── service.go
│   │   │   ├── http_handler.go
│   │   │   └── pg_repository.go
│   │   │
│   │   └── auth/                   # Identity & Account Management (3,000+ Users)
│   │       ├── auth.go
│   │       ├── interfaces.go
│   │       ├── entity.go
│   │       ├── dto.go
│   │       ├── service.go
│   │       ├── http_handler.go
│   │       └── pg_repository.go
│   │
│   ├── core/                       # Shared Utilities (Cross-cutting Concerns)
│   │   ├── response/               # Standard API JSON Responses (Success/Error)
│   │   ├── pagination/             # Pagination Helpers for Lists (not yet used by any endpoint)
│   │   ├── validation/             # Bind-error → clean message conversion (BindErrorMessage)
│   │   ├── middleware/              # RequireAuth — JWT bearer validation, sets user ID in context
│   │   └── ratelimiting/            # Per-IP token bucket rate limiter
│   │
│   ├── infra/                      # Infrastructure & Drivers
│   │   ├── db/                     # PostgreSQL Connection & Pool Initialization
│   │   │   └── postgres.go
│   │   └── storage/                # Cloud Storage Client (Supabase/Firebase) — still a stub;
│   │       └── storage.go          # kept for `shadowing`'s future audio-upload needs
│   │
│   ├── pkg/                        # Third-Party Wrapper Packages (Shared Libraries)
│   │   └── openai/                 # OpenAI SDK Wrapper (Whisper, TTS, GPT Client)
│   │       └── client.go
│   │
│   ├── config/                     # Configuration Management
│   │   └── config.go               # Load .env / Environment Variables
│   │
│   └── bootstrap/                  # Dependency Injection & Wire-up
│       └── deps.go                 # Initialize Infra, Packages, Modules, CORS middleware, Inject Them
│
├── assets/
│   └── icons/                      # Static tag/badge icon files — served at /assets/icons/<file>
│                                    # (route wired in bootstrap/deps.go; folder empty for now)
│
├── db/
│   ├── migrations/                 # SQL Migration Files — 000001-000006 applied (users, vocabularies,
│   │                                # tags, vocabulary_tags, flashcards, ai_suggested_sentences,
│   │                                # vocabularies.level, flashcard_sentences)
│   └── seed/                       # Seed data for vocabulary/tags (dev convenience)
│
├── bruno/                          # REST API test collection (auth, flashcard — more to add)
│
├── web-test/                       # Static HTML/CSS/JS mockups — NOT part of the backend/production
│   │                                # frontend. Manual test tools & UX flow review only.
│   ├── mobile-test.html            # Hits the real auth + tags/flashcards generate endpoints
│   └── app-flow-mockup.html        # Full app flow mockup (mock data, no backend calls)
│
├── .env / .env.local
├── go.mod
└── go.sum
```