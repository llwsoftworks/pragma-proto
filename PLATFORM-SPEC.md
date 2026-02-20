# Student Grading Platform — Technical Specification

This is the authoritative spec for building a web-based SaaS student grading platform. Follow it precisely.

---

## 1. Priorities (Strict Order)

1. **Security** — Defense-in-depth. Assume every layer can fail. Student data is protected under FERPA.
2. **Performance** — Sub-100ms grade entry. Sub-3s report generation. 100K+ req/sec on the API.
3. **Stability** — Near-zero downtime. Explicit error handling. No silent failures. No partial data writes.
4. **Simplicity** — Least code possible. Least dependencies possible. Every file and function should be obvious in purpose.
5. **User Experience** — Fewest clicks per action. Keyboard-first. Glanceable dashboards. Seamless and pleasing UI.

---

## 2. Technology Stack

| Layer | Technology |
|---|---|
| **Frontend** | SvelteKit with TypeScript |
| **UI** | Tailwind CSS + shadcn-svelte (or Skeleton UI) |
| **Backend API** | Go 1.22+ with Chi router |
| **Database Access** | sqlc (SQL → type-safe Go code) |
| **Database** | PostgreSQL 16 (via Neon — serverless, branching) |
| **Authentication** | Custom Go middleware: Argon2id password hashing + JWT (Ed25519) + TOTP MFA |
| **File Storage** | Cloudflare R2 (S3-compatible, zero egress fees) |
| **AI** | Anthropic Claude API (claude-sonnet-4-5-20250929) via Go proxy |
| **PDF Generation** | chromedp (headless Chrome in Go) or go-pdf |
| **Frontend Hosting** | Vercel or Cloudflare Pages |
| **API Hosting** | Railway or Fly.io |
| **Monitoring** | Sentry (errors) + Prometheus + Grafana (metrics) |
| **Email** | Resend (transactional email) |

---

## 3. Architecture

```
Browser → SvelteKit (Vercel) → Go API (Railway/Fly.io) → PostgreSQL (Neon)
                                       ↓          ↓
                                  Cloudflare R2   Claude API
                                   (files)       (AI features)
```

**Rules:**
- SvelteKit is a display layer ONLY. Zero business logic. Zero database access. It calls the Go API for all data.
- The Go API is the ONLY thing that talks to the database.
- The Go API is the ONLY thing that enforces permissions, calculates grades, generates PDFs, and proxies AI calls.
- SvelteKit's `+page.server.ts` files call the Go API. These files are physically excluded from the client bundle by the Svelte compiler.
- `+page.svelte` files render the UI. They receive data as props from `+page.server.ts` and handle user interactions.
- All state-changing operations go: Svelte UI → `+page.server.ts` → Go API → database.

---

## 4. User Roles

Five roles with strict hierarchical permissions:

| Role | Can Do | Cannot Do |
|---|---|---|
| **SUPER_ADMIN** | Everything. Manage multiple schools. Platform-level settings. | — |
| **ADMIN** | Manage their school: users, courses, grade locks, settings, reports, branding, document templates. View all data within their school. | Access other schools. Platform-level settings. |
| **TEACHER** | Create/edit assignments. Enter/edit grades. Attach files. View grades for their assigned courses only. Generate report cards for their students. Use AI features. Build their schedule. Send parent communications. | View grades outside their courses. Lock/unlock student grades. Manage users. Access admin settings. |
| **PARENT** | View linked children's grades (subject to grade lock). View/download report cards. Generate enrollment certificates, attendance letters, and other documents. View child's schedule. Receive teacher communications. | View other students' data. Edit any data. Access teacher or admin functions. |
| **STUDENT** | View their own grades (subject to grade lock). View their schedule. View their digital ID. Request documents. | View other students' data. Edit any data. Access teacher, parent, or admin functions. |

**Parent-Student linking:** A parent account can be linked to one or more student accounts. The `parent_students` table defines which parents can see which students. A parent sees ONLY their linked children's data. This is enforced at the Go API level on every request.

**Grade lock behavior by role:**
- When a student's grades are locked: `STUDENT` and `PARENT` roles see a generic restriction message. They cannot see any grades.
- `TEACHER` and `ADMIN` roles can still view and edit grades for locked students. The lock only affects student/parent visibility.

---

## 5. Project Structure

### 5.1 Frontend (SvelteKit)

```
src/
├── routes/
│   ├── (auth)/
│   │   ├── login/
│   │   │   ├── +page.svelte
│   │   │   └── +page.server.ts
│   │   ├── register/
│   │   │   ├── +page.svelte
│   │   │   └── +page.server.ts
│   │   └── forgot-password/
│   │       ├── +page.svelte
│   │       └── +page.server.ts
│   │
│   ├── (dashboard)/
│   │   ├── +layout.svelte                ← Shared layout: sidebar nav, top bar, role-based menu
│   │   ├── +layout.server.ts             ← Auth guard: verify JWT, load user, redirect if unauthorized
│   │   │
│   │   ├── teacher/
│   │   │   ├── +page.svelte              ← Dashboard: today's schedule, alerts, quick actions, recent activity
│   │   │   ├── +page.server.ts           ← Load dashboard data from Go API
│   │   │   ├── grades/
│   │   │   │   ├── +page.svelte          ← Course selector → grade entry grid
│   │   │   │   ├── +page.server.ts
│   │   │   │   └── [courseId]/
│   │   │   │       ├── +page.svelte      ← Full grade grid for one course (inline editing, keyboard nav)
│   │   │   │       └── +page.server.ts
│   │   │   ├── assignments/
│   │   │   │   ├── +page.svelte          ← Assignment list by course
│   │   │   │   ├── +page.server.ts
│   │   │   │   ├── new/
│   │   │   │   │   ├── +page.svelte      ← Create assignment form (title, rubric, due date, file upload)
│   │   │   │   │   └── +page.server.ts
│   │   │   │   └── [assignmentId]/
│   │   │   │       ├── +page.svelte      ← Edit assignment, view submissions
│   │   │   │       └── +page.server.ts
│   │   │   ├── schedule/
│   │   │   │   ├── +page.svelte          ← Drag-and-drop weekly schedule builder
│   │   │   │   └── +page.server.ts
│   │   │   ├── reports/
│   │   │   │   ├── +page.svelte          ← Generate report cards (single or batch)
│   │   │   │   └── +page.server.ts
│   │   │   └── ai/
│   │   │       ├── +page.svelte          ← AI assistant: grading help, comments, insights
│   │   │       └── +page.server.ts
│   │   │
│   │   ├── student/
│   │   │   ├── +page.svelte              ← Dashboard: current grades overview, upcoming assignments, schedule
│   │   │   ├── +page.server.ts
│   │   │   ├── grades/
│   │   │   │   ├── +page.svelte          ← View grades by course (checks grade lock)
│   │   │   │   └── +page.server.ts
│   │   │   ├── schedule/
│   │   │   │   ├── +page.svelte          ← View weekly schedule
│   │   │   │   └── +page.server.ts
│   │   │   ├── id-card/
│   │   │   │   ├── +page.svelte          ← Digital ID card (QR code, offline-capable)
│   │   │   │   └── +page.server.ts
│   │   │   └── documents/
│   │   │       ├── +page.svelte          ← Request/download certificates
│   │   │       └── +page.server.ts
│   │   │
│   │   ├── parent/
│   │   │   ├── +page.svelte              ← Dashboard: overview of all linked children's grades, alerts
│   │   │   ├── +page.server.ts
│   │   │   ├── grades/
│   │   │   │   ├── +page.svelte          ← View child's grades (child selector if multiple children)
│   │   │   │   └── +page.server.ts
│   │   │   ├── reports/
│   │   │   │   ├── +page.svelte          ← View/download child's report cards
│   │   │   │   └── +page.server.ts
│   │   │   ├── documents/
│   │   │   │   ├── +page.svelte          ← Generate enrollment certs, attendance letters
│   │   │   │   └── +page.server.ts
│   │   │   └── messages/
│   │   │       ├── +page.svelte          ← View messages from teachers
│   │   │       └── +page.server.ts
│   │   │
│   │   └── admin/
│   │       ├── +page.svelte              ← Dashboard: school-wide stats, pending actions
│   │       ├── +page.server.ts
│   │       ├── students/
│   │       │   ├── +page.svelte          ← Student roster, search, filter
│   │       │   └── +page.server.ts
│   │       ├── teachers/
│   │       │   ├── +page.svelte          ← Teacher management
│   │       │   └── +page.server.ts
│   │       ├── parents/
│   │       │   ├── +page.svelte          ← Parent accounts, link parents to students
│   │       │   └── +page.server.ts
│   │       ├── courses/
│   │       │   ├── +page.svelte          ← Course management, assign teachers
│   │       │   └── +page.server.ts
│   │       ├── grade-locks/
│   │       │   ├── +page.svelte          ← Lock/unlock student grade access (single + bulk)
│   │       │   └── +page.server.ts
│   │       ├── reports/
│   │       │   ├── +page.svelte          ← School-wide reports, batch generation
│   │       │   └── +page.server.ts
│   │       └── settings/
│   │           ├── +page.svelte          ← School branding, AI toggles, grading scales, document templates
│   │           └── +page.server.ts
│   │
│   └── verify/
│       └── [code]/
│           └── +page.server.ts           ← Public document/ID verification endpoint
│
├── lib/
│   ├── components/
│   │   ├── ui/                           ← Base UI primitives (Button, Input, Modal, Toast, etc.)
│   │   ├── GradeInput.svelte             ← Single grade entry cell (inline, keyboard-navigable)
│   │   ├── GradeGrid.svelte             ← Full grade grid (students × assignments matrix)
│   │   ├── ScheduleBuilder.svelte        ← Drag-and-drop weekly calendar
│   │   ├── ScheduleView.svelte          ← Read-only schedule display
│   │   ├── FileUpload.svelte            ← Drag-and-drop file upload with R2 presigned URLs
│   │   ├── ReportCard.svelte            ← Report card preview component
│   │   ├── DigitalId.svelte             ← Digital student ID card with QR code
│   │   ├── ChildSelector.svelte         ← Parent: dropdown/tabs to switch between linked children
│   │   ├── AlertBadge.svelte            ← Dashboard alert/notification badge
│   │   └── DocumentGenerator.svelte     ← Document type selector + generate button
│   │
│   ├── stores/
│   │   ├── auth.ts                       ← Current user session (role, school_id, name)
│   │   ├── notifications.ts             ← Toast notifications (success, error, undo)
│   │   └── theme.ts                     ← Dark mode toggle (system-preference-aware)
│   │
│   ├── api.ts                            ← Typed Go API client (fetch wrapper with auth headers, error handling)
│   └── utils.ts                          ← Shared utilities (date formatting, grade calculations for display only)
│
├── app.html                              ← Root HTML shell
├── app.css                               ← Tailwind base + global styles
└── hooks.server.ts                       ← Global: security headers, auth cookie parsing, request logging
```

### 5.2 Backend (Go)

```
api/
├── cmd/
│   └── server/
│       └── main.go                        ← Entry point: init DB, init router, start server
│
├── internal/
│   ├── auth/
│   │   ├── jwt.go                         ← JWT creation (Ed25519), validation, refresh
│   │   ├── middleware.go                  ← Auth middleware: extract JWT from cookie, validate, set context
│   │   ├── mfa.go                         ← TOTP generation, verification (crypto/hmac, crypto/sha1)
│   │   └── passwords.go                  ← Argon2id hash + verify. HIBP breached password check.
│   │
│   ├── handlers/
│   │   ├── auth.go                        ← Login, register, logout, password reset, MFA setup/verify
│   │   ├── grades.go                      ← CRUD grades. Weighted grade calculation. GPA calculation.
│   │   ├── assignments.go                ← CRUD assignments. Attachment management (R2 presigned URLs).
│   │   ├── courses.go                    ← CRUD courses. Enrollment management.
│   │   ├── students.go                   ← Student profile management. Parent linking.
│   │   ├── parents.go                    ← Parent account management. Child linking. Grade viewing (with lock check).
│   │   ├── schedule.go                   ← CRUD schedule blocks. Conflict detection. iCal export.
│   │   ├── reports.go                    ← Generate report cards (PDF). Batch generation. Historical storage.
│   │   ├── documents.go                  ← Generate enrollment certs, attendance letters (PDF). Verification codes.
│   │   ├── digital_id.go                ← Generate digital IDs. QR code creation. Verification endpoint.
│   │   ├── admin.go                      ← Grade locking (single + bulk). User management. School settings.
│   │   ├── ai.go                         ← AI proxy: anonymize data → call Claude → de-anonymize → return.
│   │   └── dashboard.go                 ← Dashboard data aggregation per role (today's schedule, alerts, stats).
│   │
│   ├── middleware/
│   │   ├── ratelimit.go                  ← Per-user: 100 req/min. Per-IP: 10 login attempts/hr. AI: 20 req/min.
│   │   ├── rbac.go                       ← Role check per endpoint. Returns 403 if unauthorized.
│   │   ├── audit.go                      ← Auto-log all state-changing requests to audit_logs table.
│   │   ├── tenant.go                     ← Extract school_id from JWT. Inject into all DB queries. Enforced on every request.
│   │   └── cors.go                       ← CORS: allow only the SvelteKit frontend origin.
│   │
│   ├── models/
│   │   ├── user.go                        ← User, Student, Teacher, Parent structs
│   │   ├── grade.go                       ← Grade, GradeCalculation, LetterGradeMapping structs
│   │   ├── assignment.go                 ← Assignment, Attachment structs
│   │   ├── course.go                     ← Course, Enrollment structs
│   │   ├── schedule.go                   ← ScheduleBlock struct
│   │   ├── report.go                     ← ReportCard, Document structs
│   │   ├── digital_id.go                ← DigitalId struct
│   │   └── school.go                     ← School, SchoolSettings structs
│   │
│   ├── database/
│   │   ├── queries/
│   │   │   ├── users.sql                  ← User CRUD, role queries, parent-student linking
│   │   │   ├── grades.sql                ← Grade CRUD, aggregations, GPA calculations
│   │   │   ├── assignments.sql           ← Assignment CRUD, attachment queries
│   │   │   ├── courses.sql               ← Course CRUD, enrollment queries
│   │   │   ├── schedule.sql              ← Schedule block CRUD, conflict detection queries
│   │   │   ├── reports.sql               ← Report card queries, document queries
│   │   │   ├── admin.sql                 ← Grade lock queries, bulk operations
│   │   │   ├── audit.sql                 ← Audit log insert (append-only), read queries
│   │   │   └── dashboard.sql             ← Dashboard aggregation queries per role
│   │   │
│   │   ├── migrations/
│   │   │   ├── 001_create_schools.sql
│   │   │   ├── 002_create_users.sql
│   │   │   ├── 003_create_students_teachers.sql
│   │   │   ├── 004_create_parent_students.sql
│   │   │   ├── 005_create_courses_enrollments.sql
│   │   │   ├── 006_create_assignments_attachments.sql
│   │   │   ├── 007_create_grades.sql
│   │   │   ├── 008_create_schedule_blocks.sql
│   │   │   ├── 009_create_grade_locks.sql
│   │   │   ├── 010_create_report_cards_documents.sql
│   │   │   ├── 011_create_digital_ids.sql
│   │   │   ├── 012_create_ai_interactions.sql
│   │   │   ├── 013_create_audit_logs.sql
│   │   │   ├── 014_create_sessions.sql
│   │   │   └── 015_create_rls_policies.sql
│   │   │
│   │   └── db.go                          ← Connection pool setup, health check, migration runner
│   │
│   └── services/
│       ├── grading.go                     ← Weighted grade calculation, GPA, letter grade mapping
│       ├── pdf.go                         ← PDF generation: report cards, enrollment certs, IDs
│       ├── storage.go                     ← R2 operations: presigned upload/download URLs, delete
│       ├── ai.go                          ← Claude API client: anonymize, call, de-anonymize, validate
│       ├── email.go                       ← Transactional email via Resend: notifications, reports, resets
│       └── verification.go               ← HMAC-based document/ID verification code generation + validation
│
├── config/
│   └── config.go                          ← Env-based config: DB URL, R2 creds, JWT secret, Claude API key, etc.
│
├── go.mod
└── go.sum
```

---

## 6. Database Schema

All tables have `school_id` for multi-tenancy. All queries are scoped by `school_id` via Go middleware + PostgreSQL RLS as defense-in-depth. The `audit_logs` table is append-only (no UPDATE/DELETE permissions).

### 6.1 Core Tables

```sql
-- schools
CREATE TABLE schools (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    address         TEXT,
    logo_url        TEXT,
    settings        JSONB DEFAULT '{}',  -- grading scales, branding, AI toggle, etc.
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- users (all roles share this table)
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id),
    role            TEXT NOT NULL CHECK (role IN ('super_admin', 'admin', 'teacher', 'parent', 'student')),
    email           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    first_name      TEXT NOT NULL,
    last_name       TEXT NOT NULL,
    phone           TEXT,
    profile_photo   TEXT,
    mfa_secret      TEXT,               -- encrypted TOTP secret, NULL if MFA not enabled
    mfa_enabled     BOOLEAN DEFAULT FALSE,
    is_active       BOOLEAN DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (school_id, email)
);

-- students (extends users)
CREATE TABLE students (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    school_id       UUID NOT NULL REFERENCES schools(id),
    student_number  TEXT NOT NULL,       -- school-issued student ID number
    grade_level     TEXT NOT NULL,       -- e.g., "9th", "10th", "K", "Pre-K"
    enrollment_date DATE NOT NULL,
    enrollment_status TEXT DEFAULT 'active' CHECK (enrollment_status IN ('active', 'withdrawn', 'graduated', 'transferred')),
    is_grade_locked BOOLEAN DEFAULT FALSE,
    lock_reason     TEXT,               -- NULL when not locked. Never shown to student/parent.
    date_of_birth   DATE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (school_id, student_number)
);

-- teachers (extends users)
CREATE TABLE teachers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    school_id       UUID NOT NULL REFERENCES schools(id),
    department      TEXT,
    title           TEXT,               -- e.g., "Lead Teacher", "Department Head"
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- parent_students (links parents to their children)
CREATE TABLE parent_students (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    school_id       UUID NOT NULL REFERENCES schools(id),
    relationship    TEXT NOT NULL CHECK (relationship IN ('mother', 'father', 'guardian', 'other')),
    is_primary_contact BOOLEAN DEFAULT FALSE,
    can_view_grades BOOLEAN DEFAULT TRUE,
    can_generate_docs BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (parent_id, student_id)
);
```

### 6.2 Academic Tables

```sql
-- courses
CREATE TABLE courses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id),
    teacher_id      UUID NOT NULL REFERENCES teachers(id),
    name            TEXT NOT NULL,       -- e.g., "Algebra II"
    subject         TEXT NOT NULL,       -- e.g., "Mathematics"
    period          TEXT,               -- e.g., "Period 1", "Block A"
    room            TEXT,
    academic_year   TEXT NOT NULL,       -- e.g., "2025-2026"
    semester        TEXT,               -- e.g., "Fall", "Spring", "Full Year"
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- enrollments (students in courses)
CREATE TABLE enrollments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id       UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    school_id       UUID NOT NULL REFERENCES schools(id),
    enrolled_at     TIMESTAMPTZ DEFAULT NOW(),
    dropped_at      TIMESTAMPTZ,
    status          TEXT DEFAULT 'active' CHECK (status IN ('active', 'dropped', 'completed')),
    UNIQUE (student_id, course_id)
);

-- assignments
CREATE TABLE assignments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id       UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    school_id       UUID NOT NULL REFERENCES schools(id),
    title           TEXT NOT NULL,
    description     TEXT,
    due_date        TIMESTAMPTZ,
    max_points      DECIMAL(8,2) NOT NULL,
    category        TEXT NOT NULL CHECK (category IN ('homework', 'quiz', 'test', 'exam', 'project', 'classwork', 'participation', 'other')),
    weight          DECIMAL(5,4) DEFAULT 1.0, -- weight within its category
    is_published    BOOLEAN DEFAULT FALSE,    -- students can't see until published
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- assignment_attachments
CREATE TABLE assignment_attachments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id   UUID NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    school_id       UUID NOT NULL REFERENCES schools(id),
    file_name       TEXT NOT NULL,
    file_key        TEXT NOT NULL,       -- R2 object key
    file_size       BIGINT NOT NULL,     -- bytes
    mime_type       TEXT NOT NULL,
    uploaded_by     UUID NOT NULL REFERENCES users(id),
    version         INT DEFAULT 1,
    is_current      BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- grades
CREATE TABLE grades (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id   UUID NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    school_id       UUID NOT NULL REFERENCES schools(id),
    points_earned   DECIMAL(8,2),        -- NULL if not yet graded
    letter_grade    TEXT,                -- calculated from points, stored for historical snapshots
    comment         TEXT,                -- teacher comment on this grade
    graded_by       UUID REFERENCES users(id),
    graded_at       TIMESTAMPTZ,
    ai_suggested    DECIMAL(8,2),        -- AI suggested grade (NULL if not used)
    ai_accepted     BOOLEAN,             -- did teacher accept AI suggestion?
    is_excused      BOOLEAN DEFAULT FALSE,
    is_missing      BOOLEAN DEFAULT FALSE,
    is_late         BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (assignment_id, student_id)
);
```

### 6.3 Schedule, Documents, IDs

```sql
-- schedule_blocks
CREATE TABLE schedule_blocks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    course_id       UUID REFERENCES courses(id),   -- NULL for personal blocks
    day_of_week     INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sunday
    start_time      TIME NOT NULL,
    end_time        TIME NOT NULL,
    room            TEXT,
    label           TEXT,               -- for personal blocks: "Study Hall", "Lunch"
    color           TEXT,               -- hex color code
    semester        TEXT,
    is_recurring    BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    CHECK (end_time > start_time)
);

-- grade_locks
CREATE TABLE grade_locks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id      UUID NOT NULL REFERENCES students(id),
    school_id       UUID NOT NULL REFERENCES schools(id),
    locked_by       UUID NOT NULL REFERENCES users(id), -- admin who locked
    reason          TEXT NOT NULL,       -- e.g., "Outstanding tuition - January 2026"
    locked_at       TIMESTAMPTZ DEFAULT NOW(),
    unlocked_at     TIMESTAMPTZ,         -- NULL while active
    unlocked_by     UUID REFERENCES users(id),
    is_active       BOOLEAN DEFAULT TRUE
);

-- report_cards
CREATE TABLE report_cards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id      UUID NOT NULL REFERENCES students(id),
    school_id       UUID NOT NULL REFERENCES schools(id),
    academic_period TEXT NOT NULL,        -- e.g., "2025-2026 Fall Semester"
    gpa             DECIMAL(4,3),
    teacher_comments TEXT,
    admin_comments  TEXT,
    is_finalized    BOOLEAN DEFAULT FALSE,
    pdf_url         TEXT,                -- R2 key for generated PDF
    generated_by    UUID NOT NULL REFERENCES users(id),
    generated_at    TIMESTAMPTZ DEFAULT NOW()
);

-- documents (enrollment certs, attendance letters, etc.)
CREATE TABLE documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id),
    student_id      UUID NOT NULL REFERENCES students(id),
    type            TEXT NOT NULL CHECK (type IN ('enrollment_certificate', 'attendance_letter', 'academic_standing', 'tuition_confirmation', 'custom')),
    verification_code TEXT NOT NULL UNIQUE, -- HMAC-signed code for authenticity verification
    pdf_url         TEXT NOT NULL,        -- R2 key
    generated_by    UUID NOT NULL REFERENCES users(id), -- parent, student, or admin who requested
    expires_at      TIMESTAMPTZ,         -- some documents expire
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- digital_ids
CREATE TABLE digital_ids (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id      UUID NOT NULL REFERENCES students(id),
    school_id       UUID NOT NULL REFERENCES schools(id),
    id_number       TEXT NOT NULL,       -- formatted display number
    qr_code_data    TEXT NOT NULL,        -- data encoded in QR (verification URL)
    barcode_data    TEXT,
    photo_url       TEXT,                -- R2 key for student photo
    issued_at       TIMESTAMPTZ DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    is_valid        BOOLEAN DEFAULT TRUE,
    revoked_at      TIMESTAMPTZ,
    UNIQUE (school_id, id_number)
);
```

### 6.4 System Tables

```sql
-- sessions
CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    school_id       UUID NOT NULL REFERENCES schools(id),
    token_hash      TEXT NOT NULL,       -- hash of JWT, for invalidation
    ip_address      INET,
    user_agent      TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- audit_logs (APPEND-ONLY — no UPDATE or DELETE permissions)
CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id),
    user_id         UUID REFERENCES users(id),
    action          TEXT NOT NULL,       -- e.g., "grade.update", "grade_lock.create", "document.generate"
    entity_type     TEXT NOT NULL,       -- e.g., "grade", "assignment", "student", "grade_lock"
    entity_id       UUID,
    old_value       JSONB,               -- previous state (NULL for creates)
    new_value       JSONB,               -- new state (NULL for deletes)
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ai_interactions
CREATE TABLE ai_interactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    feature         TEXT NOT NULL CHECK (feature IN ('grading_assistant', 'student_insights', 'report_comments', 'smart_scheduling', 'parent_communication')),
    input_summary   TEXT,                -- what was sent (anonymized)
    output_summary  TEXT,                -- what was returned
    tokens_used     INT,
    accepted        BOOLEAN,             -- did user accept the AI output?
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

### 6.5 Key Indexes

```sql
-- Multi-tenancy (on every table)
CREATE INDEX idx_users_school ON users(school_id);
CREATE INDEX idx_students_school ON students(school_id);
CREATE INDEX idx_courses_school ON courses(school_id);
CREATE INDEX idx_grades_school ON grades(school_id);

-- Performance-critical queries
CREATE INDEX idx_grades_student ON grades(student_id, school_id);
CREATE INDEX idx_grades_assignment ON grades(assignment_id);
CREATE INDEX idx_enrollments_student ON enrollments(student_id, status);
CREATE INDEX idx_enrollments_course ON enrollments(course_id, status);
CREATE INDEX idx_assignments_course ON assignments(course_id, is_published);
CREATE INDEX idx_parent_students_parent ON parent_students(parent_id);
CREATE INDEX idx_parent_students_student ON parent_students(student_id);
CREATE INDEX idx_schedule_blocks_user ON schedule_blocks(user_id, day_of_week);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_sessions_user ON sessions(user_id, expires_at);
CREATE INDEX idx_documents_verification ON documents(verification_code);
```

---

## 7. Security Requirements

### 7.1 Authentication

- Argon2id for password hashing (tunable memory: 64MB, iterations: 3, parallelism: 4).
- Minimum 12-character passwords. Check against Have I Been Pwned breached password API on registration and password change.
- JWT tokens signed with Ed25519. Stored in HTTP-only, Secure, SameSite=Strict cookies.
- Token expiration: 15 minutes (teachers/admins), 24 hours (parents/students). Sliding window refresh.
- TOTP-based MFA required for `teacher`, `admin`, `super_admin`. Optional for `parent`, `student`.
- Account lockout: 5 failed attempts → 15-minute lockout. 15 failures → account locked, admin reset required.
- Password change invalidates all other sessions immediately.

### 7.2 Authorization (RBAC)

Every Go API handler MUST check the user's role before doing anything. Use middleware that:
1. Extracts JWT from cookie → validates signature and expiration.
2. Extracts `school_id` and `role` from JWT claims.
3. Injects `school_id` into all database queries (tenant scoping).
4. Checks role against the endpoint's required permission.
5. Returns 403 Forbidden if unauthorized.

Parent-specific: when a parent requests student data, the handler MUST verify the parent-student link exists in `parent_students` before returning any data.

### 7.3 Data Protection

- All database queries scoped by `school_id` at Go middleware level. PostgreSQL RLS as defense-in-depth backup.
- Field-level encryption (AES-256-GCM) for any PII beyond name/email: date of birth, medical info, SSN if ever stored. Separate encryption keys per school.
- Audit log is append-only. Revoke UPDATE and DELETE on the `audit_logs` table at the database level.
- Every grade change logged with: user_id, old_value, new_value, ip_address, timestamp.

### 7.4 Input Validation

- Go struct validation using `go-playground/validator` on every API request body.
- SvelteKit validates on the frontend for UX. The Go API re-validates everything — trust nothing from the client.
- Reject requests with unexpected fields (strict JSON decoding in Go: `DisallowUnknownFields`).

### 7.5 File Security

- Uploads go directly from browser to R2 via presigned URLs generated by Go API. Files never pass through the Go server.
- Presigned upload URLs: valid for 15 minutes, scoped to specific file path, max size 25MB.
- Presigned download URLs: valid for 1 hour, regenerated on each request.
- MIME type validation in Go before generating upload URL. Allowed: PDF, DOCX, PPTX, XLSX, JPG, PNG, GIF, MP3, MP4.
- Virus scanning on upload completion (ClamAV webhook or on-access scan).

### 7.6 Rate Limiting

| Endpoint Category | Limit |
|---|---|
| General API | 100 requests/minute per user |
| Login | 10 attempts/hour per IP |
| Password reset | 5 requests/hour per email |
| AI features | 20 requests/minute per user |
| Document generation | 5 per day per user |
| File upload | 20 per hour per user |

### 7.7 Security Headers

Set in both SvelteKit `hooks.server.ts` and Go API response middleware:

```
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'
```

---

## 8. Feature Specifications

### 8.1 Printable Grade Reports

- Generate report cards as PDF server-side via Go.
- Templates customizable per school: logo, colors, grading scale, comment fields, signature line.
- Single report: Teacher/admin selects student + academic period → Go generates PDF → returns download URL.
- Batch: Admin selects grade level or entire school + period → Go generates all reports concurrently via goroutines → returns ZIP download or individual links.
- Parents can view/download their children's finalized report cards.
- Each generated report is stored immutably in R2. If grades change later, the historical PDF is preserved.
- Grade calculations are in Go ONLY. The frontend never calculates grades.

**Permissions:** Teachers can generate for their course students. Admins for any student in their school. Parents can download finalized reports for their linked children. Students can view their own finalized reports.

### 8.2 Administrative Grade Locking

- Admin Dashboard → Students → Select student(s) → Lock/Unlock.
- Lock = student and parent roles cannot see ANY grades for that student. Teachers and admins still see everything.
- Lock reason stored in database. Never displayed to student or parent.
- Student/parent sees: *"Your grade access has been temporarily restricted. Please contact your school administration."*
- Bulk lock: Admin uploads CSV or selects filter (e.g., "all students with overdue January tuition") → lock all matching.
- Unlock sends email notification to student and linked parent(s).
- Every lock/unlock is audit-logged with: admin who acted, student affected, reason, timestamp.

**Permissions:** Only `admin` and `super_admin` can lock/unlock.

### 8.3 Direct Assignment Attachments

- Teacher creates assignment → drag-and-drop files into upload zone → files upload directly to R2 via presigned URL → file metadata saved in `assignment_attachments`.
- Multiple files per assignment (up to 100MB total per assignment, 25MB per file).
- In-browser preview for PDFs and images. Download button for all other types.
- Version history: replacing a file sets `is_current = FALSE` on the old record, creates a new record. Old files retained in R2 for 90 days.
- Students see attachments when assignment is published (`is_published = TRUE`).

**Permissions:** Upload restricted to course teacher. Download restricted to enrolled students, course teacher, and school admins. Parents can view attachments for their linked children's courses.

### 8.4 Native Schedule Builder

- Teacher: drag-and-drop time blocks onto a weekly grid. Assign course, room, color. Set recurring (every week) or one-time.
- Student: view enrolled course schedule (auto-populated from course data). Add personal blocks (study time, lunch, clubs).
- Admin: view any teacher or student schedule. Detect room/teacher conflicts across the school.
- Conflict detection: when saving a schedule block, Go API checks for overlapping blocks for the same room or teacher. Returns specific conflict details.
- Export to .ics (iCal) for syncing to Google Calendar, Apple Calendar, Outlook.
- Parents can view their linked children's schedules.

**UI:** Weekly view is default. Day view available. Color-coded by subject/course. Mobile: swipe between days.

### 8.5 AI Integration

All AI features are opt-in per school (admin toggle in settings). All AI outputs require human confirmation before taking effect.

**Grading Assistant:**
- Teacher selects assignment + rubric → selects student submissions to grade.
- Go API anonymizes data (strips names → "Student A", "Student B").
- Go calls Claude API with: anonymized responses, rubric, max points.
- Claude returns: suggested grade + reasoning per student.
- Go de-anonymizes, validates (grades within 0 to max_points), stores suggestions in `grades.ai_suggested`.
- Teacher sees suggestions in the grade grid with "AI Suggested" badge. Accepts, modifies, or rejects each.
- Every interaction logged in `ai_interactions`.

**Student Insights:**
- Automatic: Go runs nightly analysis of grade trends per student (or on teacher request).
- Sends anonymized grade trajectories to Claude with prompt: "Identify at-risk students and explain the trend."
- Returns: alerts like "Student C's math average has dropped from 88% to 71% over the last 4 weeks."
- Alerts shown on teacher dashboard in "Needs Attention" section.
- Teacher can dismiss or act on each alert.

**Report Card Comments:**
- Teacher clicks "Generate Comment" for a student.
- Go sends to Claude: anonymized grade summary, attendance data, trend direction.
- Claude returns: professional, personalized comment.
- Teacher edits and approves before it's saved to the report card.

**Smart Scheduling:**
- Admin or teacher requests schedule optimization.
- Go sends to Claude: available rooms, teacher preferences, course enrollment counts, time constraints.
- Claude returns: candidate schedules ranked by optimization criteria (minimal conflicts, balanced load).
- Admin/teacher reviews and selects one to apply.

**Parent Communication:**
- Teacher describes a concern about a student.
- Go anonymizes and sends to Claude with tone preference (formal, warm, urgent).
- Claude drafts an email.
- Teacher reviews, edits, and sends via the platform (through Resend).

**Privacy rules (non-negotiable):**
- Student names, IDs, and PII are NEVER sent to Claude. Only anonymized identifiers and numerical data.
- AI can be fully disabled per school in admin settings.
- "AI-assisted" badge on any content generated with AI help.
- All interactions logged with input/output summary and acceptance status.

### 8.6 Digital Student IDs

- Each student gets a digital ID card accessible at `/student/id-card`.
- Card displays: student photo, full name, grade level, student number, school name, school logo, barcode, QR code, issue date, expiration date.
- QR code encodes a verification URL: `https://platform.com/verify/[HMAC-signed-code]`.
- Verification endpoint (public, no auth required): returns ONLY valid/invalid status + student name + photo. No sensitive data.
- Offline capable: cache the ID card page as a PWA so it works without internet.
- Printable: CSS print stylesheet formats the card to standard ID dimensions (3.375" × 2.125").
- Admin can revoke any ID instantly → `is_valid = FALSE`, `revoked_at = NOW()`. Scanning revoked ID returns "INVALID".
- IDs expire annually. Admin or system triggers renewal.

**Permissions:** Students see their own ID. Parents see their linked children's IDs. Admins can view/revoke/regenerate any ID.

### 8.7 Automated Document Generation

Available document types:
- **Proof of Enrollment Certificate** — confirms student is currently enrolled.
- **Proof of Attendance Letter** — confirms student has been attending.
- **Academic Standing Letter** — reports current GPA and academic status.
- **Tuition Payment Confirmation** — confirms payment (if payment tracking is integrated).
- **Custom** — admin-defined templates.

**Flow:** Parent/student goes to Documents → selects type → Go generates PDF with school letterhead, student info, date, verification code, QR code → user downloads.

- Each document includes a unique `verification_code` (HMAC-SHA256 of document ID + school secret).
- Anyone can verify at `https://platform.com/verify/[code]` — returns document type, student name, issue date, and validity status.
- Admin can customize document templates (wording, logo, signature image) in school settings.
- Rate limited: 5 documents per day per user.

**Permissions:** Students can generate for themselves. Parents can generate for linked children. Admins can generate for any student.

---

## 9. UI/UX Principles

### 9.1 General

- **Fewest clicks wins.** Every common action should take ≤ 3 clicks. See click targets below.
- **Keyboard-first.** Grade entry: Tab between cells, Enter to save, arrow keys to navigate. No mouse required.
- **Inline editing.** Click a grade cell → type → Tab to next. No modal dialogs for simple edits.
- **Auto-save.** All changes save automatically. Visual confirmation: subtle green flash on the saved cell. No "Save" button for routine tasks.
- **Undo everywhere.** Every destructive or state-changing action shows a 10-second undo toast.
- **Empty states.** Every page with no data shows a helpful message with a clear call to action: "No assignments yet. Create your first one →"
- **Loading skeletons.** Never show a blank page. Show skeleton UI during data loading.
- **Responsive.** Full functionality on desktop and tablet. Read-only essentials on phone. Grade entry is desktop/tablet only to prevent accidental edits.
- **Dark mode.** System-preference-aware toggle. Persisted in local storage.
- **Accessibility.** WCAG 2.1 AA. Screen reader tested. Keyboard navigable. Color is never the only indicator of state.

### 9.2 Click-Count Targets

| Action | Max Clicks |
|---|---|
| View a student's current grade | 2 |
| Enter grades for an assignment | 2 |
| Create a new assignment | 3 |
| Generate a report card | 3 |
| Lock/unlock a student's grades | 2 |
| View today's schedule | 0 (visible on dashboard) |
| Generate enrollment certificate | 2 |
| View digital ID | 1 |
| Attach a file to an assignment | 1 (drag and drop) |
| Parent: view child's grades | 1 (default dashboard view) |
| Parent: switch between children | 1 (child selector dropdown/tabs) |
| Parent: generate a document | 2 |

### 9.3 Teacher Dashboard ("Glanceable")

Teachers must get critical information from a 3-second glance. Dashboard layout:

```
┌──────────────────────────────────────────────────────────────────┐
│  [Logo] Platform     Dashboard  Grades  Schedule  Reports  AI   │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─ TODAY'S SCHEDULE ────────┐  ┌─ NEEDS ATTENTION ───────────┐ │
│  │ Period 1: Algebra II      │  │ ⚠ 12 ungraded assignments   │ │
│  │ Period 2: Free            │  │ ⚠ 3 past-due grades         │ │
│  │ Period 3: Geometry        │  │ 🔔 Report cards due Friday  │ │
│  │ Period 4: Algebra I       │  │ 📊 Grade drop: Maria S.     │ │
│  │                           │  │    (Algebra II, -18%)        │ │
│  └───────────────────────────┘  └─────────────────────────────┘ │
│                                                                  │
│  ┌─ QUICK ACTIONS ───────────────────────────────────────────┐  │
│  │ [+ New Assignment]  [Enter Grades]  [Generate Report]     │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─ RECENT GRADE ACTIVITY ───────────────────────────────────┐  │
│  │ Algebra II - Homework 5    Avg: 82%  ▓▓▓▓▓▓▓░░░  12/15   │  │
│  │ Geometry - Quiz 3          Avg: 74%  ▓▓▓▓▓░░░░░  28/28   │  │
│  │ Algebra I - Test 2         Avg: 88%  ▓▓▓▓▓▓▓▓░░  30/30   │  │
│  └───────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

### 9.4 Parent Dashboard

Parents must see all children's grades at a glance. If multiple children, show a tabbed or card-based layout:

```
┌──────────────────────────────────────────────────────────────────┐
│  [Logo] Platform     Dashboard  Grades  Reports  Documents      │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─ [Sofia M. (9th)] ─ [Diego M. (6th)] ────────────────────┐  │
│  │                                                            │  │
│  │  Sofia Martinez — 9th Grade — GPA: 3.6                    │  │
│  │                                                            │  │
│  │  Algebra II ........... A-  (92%)                          │  │
│  │  English 9 ............ B+  (88%)                          │  │
│  │  Biology .............. A   (95%)                          │  │
│  │  World History ........ B   (84%)                          │  │
│  │                                                            │  │
│  │  ⚠ Missing assignment: English 9 — Essay #3 (due Feb 18)  │  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─ QUICK ACTIONS ───────────────────────────────────────────┐  │
│  │ [View Report Card]  [Download Enrollment Cert]            │  │
│  └───────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

### 9.5 Design System

| Element | Specification |
|---|---|
| **Font** | Inter |
| **Colors** | Slate gray base. Blue primary actions. Green success/saved. Amber warnings. Red destructive only. |
| **Radius** | 8px cards, 6px buttons, 4px inputs |
| **Spacing** | 4px base unit. All spacing is multiples of 4. |
| **Animations** | 150ms ease-out for micro-interactions. Max 300ms. Use Svelte's `transition:` and `animate:` directives. |
| **Hierarchy** | Size → Weight → Color. Most important number is largest. |

---

## 10. Performance Targets

| Metric | Target |
|---|---|
| First Contentful Paint | < 0.8s |
| Largest Contentful Paint | < 1.5s |
| Time to Interactive | < 2.0s |
| Grade entry save (perceived) | < 100ms (optimistic UI) |
| Grade entry save (actual) | < 200ms (Go API + Neon roundtrip) |
| Single report card PDF | < 3s |
| Batch 500 report cards | < 30s |
| Search results | < 150ms |
| Dashboard data load | < 500ms |
| AI response (grading) | < 10s |

**Key strategies:**
- Optimistic UI for grade entry (save instantly in Svelte, sync to Go in background, revert on failure).
- SvelteKit SSR for fast first paint. Svelte compiles away — ~15KB JS shipped to browser.
- Go's goroutines for concurrent PDF generation and AI calls. No external queue needed.
- PostgreSQL composite indexes on all frequent query patterns.
- R2 presigned URLs for direct browser ↔ storage transfers (no Go server bottleneck for files).

---

## 11. Multi-Tenancy

- Every table has a `school_id` column.
- Go tenant middleware extracts `school_id` from JWT and injects it into every database query.
- PostgreSQL RLS policies enforce `school_id` scoping as defense-in-depth (even if Go middleware has a bug, Postgres blocks cross-tenant access).
- R2 file paths are prefixed: `school-{id}/attachments/...`, `school-{id}/reports/...`, `school-{id}/ids/...`
- No cross-school data access is ever possible.

---

## 12. Audit Logging

Every state-changing action is logged to `audit_logs`:

| action format | triggers |
|---|---|
| `grade.create` | New grade entered |
| `grade.update` | Grade modified |
| `grade.delete` | Grade removed |
| `assignment.create/update/delete` | Assignment changes |
| `grade_lock.create` | Student grades locked |
| `grade_lock.release` | Student grades unlocked |
| `document.generate` | Certificate/letter generated |
| `digital_id.issue` | New digital ID created |
| `digital_id.revoke` | Digital ID revoked |
| `user.login` | Successful login |
| `user.login_failed` | Failed login attempt |
| `user.password_change` | Password changed |
| `user.mfa_enable/disable` | MFA toggled |
| `report_card.generate` | Report card PDF generated |
| `ai.request` | AI feature used (logged separately in ai_interactions too) |
| `settings.update` | School settings changed |

Each log entry contains: `user_id`, `school_id`, `action`, `entity_type`, `entity_id`, `old_value` (JSONB), `new_value` (JSONB), `ip_address`, `user_agent`, `created_at`.

The `audit_logs` table has no UPDATE or DELETE permissions granted to the application database role. Logs are immutable.
