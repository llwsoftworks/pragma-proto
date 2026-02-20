# Pragma — Student Grading Platform

A web-based SaaS student grading platform built according to the `PLATFORM-SPEC.md` specification.

## Architecture

```
Browser → SvelteKit (Vercel) → Go API (Railway/Fly.io) → PostgreSQL (Neon)
                                       ↓          ↓
                                  Cloudflare R2   Claude API
                                   (files)       (AI features)
```

## Project Structure

```
pragma-proto/
├── api/                    ← Go 1.22+ backend
│   ├── cmd/server/         ← Entry point
│   ├── internal/
│   │   ├── auth/           ← JWT (Ed25519), Argon2id, TOTP MFA
│   │   ├── handlers/       ← HTTP route handlers
│   │   ├── middleware/     ← RBAC, rate limiting, audit, tenant, CORS
│   │   ├── models/         ← Go structs
│   │   ├── database/       ← pgx pool, migrations, SQL queries
│   │   └── services/       ← Grading, PDF, R2 storage, Claude AI, email
│   ├── config/             ← Env-based configuration
│   └── .env.example
│
├── frontend/               ← SvelteKit + TypeScript
│   ├── src/
│   │   ├── routes/         ← File-based routing
│   │   │   ├── (auth)/     ← Login, register, forgot-password
│   │   │   ├── (dashboard)/← Role-specific dashboards
│   │   │   └── verify/     ← Public verification endpoint
│   │   ├── lib/
│   │   │   ├── api.ts      ← Go API client (server-side only)
│   │   │   ├── utils.ts    ← Display-only utilities
│   │   │   ├── stores/     ← Auth, notifications, theme
│   │   │   └── components/ ← GradeGrid, GradeInput, FileUpload, DigitalId, etc.
│   │   ├── hooks.server.ts ← Security headers, JWT parsing
│   │   └── app.css         ← Tailwind + design system tokens
│   └── .env.example
│
└── PLATFORM-SPEC.md        ← Authoritative specification
```

## Getting Started

### Backend (Go API)

```bash
cd api
cp .env.example .env
# Fill in .env values

go run ./cmd/server
```

### Frontend (SvelteKit)

```bash
cd frontend
cp .env.example .env
# Fill in .env values

npm install
npm run dev
```

## Key Design Decisions

- **SvelteKit is display-only.** Zero business logic, zero DB access. All data flows through Go API.
- **Defense-in-depth multi-tenancy.** Every DB query is scoped by `school_id` at the Go middleware level AND PostgreSQL RLS.
- **Grade calculations in Go only.** The frontend never computes grades — it only displays what the API returns.
- **Student PII anonymized before AI calls.** Names, IDs, and PII never reach the Claude API.
- **Audit log is append-only.** `UPDATE`/`DELETE` permissions revoked at the database level.
- **Files go directly to R2.** The Go server generates presigned URLs; files never pass through it.

## Security

See `PLATFORM-SPEC.md` §7 for full security requirements including:
- Argon2id password hashing (64MB, 3 iterations, 4 parallelism)
- Ed25519 JWT (15 min for teachers/admins, 24h for parents/students)
- TOTP MFA required for teacher/admin/super_admin roles
- Account lockout: 5 attempts → 15-min lock; 15 → admin reset required
- HIBP breached password check on registration and password change

## User Roles

| Role | Access |
|------|--------|
| `super_admin` | Everything |
| `admin` | School-level management |
| `teacher` | Courses they're assigned to |
| `parent` | Linked children only |
| `student` | Own data only |
