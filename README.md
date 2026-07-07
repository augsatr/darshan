# Darshan

Virtual temple exploration platform — 3D, 360°, and guided virtual visits to temples across India.

## Stack

| Layer     | Technology                          |
|-----------|-------------------------------------|
| Frontend  | Next.js (React), Tailwind CSS v4    |
| Backend   | Go (chi router), pgx                |
| Database  | PostgreSQL                          |
| Auth      | JWT (golang-jwt), bcrypt            |
| Maps      | Leaflet (CartoDB dark tiles)        |
| Storage   | Cloudflare R2 / AWS S3              |
| Deploy    | Fly.io (API), Vercel (frontend)     |

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | — | Health check |
| GET | `/temples` | — | List all published temples |
| GET | `/temples/{slug}` | — | Get temple by slug |
| POST | `/auth/signup` | — | Create account |
| POST | `/auth/login` | — | Sign in |
| GET | `/auth/me` | JWT | Current user |
| GET | `/favorites` | JWT | List user's favorites |
| POST | `/favorites/{slug}` | JWT | Add favorite |
| DELETE | `/favorites/{slug}` | JWT | Remove favorite |
| GET | `/reviews/{slug}` | — | List reviews for temple |
| POST | `/reviews/{slug}` | JWT | Create review |
| POST | `/views` | — | Record page view (analytics) |
| GET | `/popular` | — | Most viewed temples |
| POST | `/admin/temples` | Admin | Create/update temple |
| DELETE | `/admin/temples/{slug}` | Admin | Delete temple |

## Getting started

```bash
# Start Postgres
make db-up

# Run migrations
make migrate

# Seed temple data (20 temples)
make seed

# Seed 34 more temples (54 total)
make seed2

# Start API (with air hot-reload)
make dev-api

# Start frontend (another terminal)
make dev-frontend

# Or run everything via Docker
make up
```

## Seed data

```bash
# Both seed commands are idempotent — safe to re-run
make seed       # 20 temples (temples.json)
make seed2      # 34 more temples (temples2.json)
```

## Deploy

### API (Fly.io)
```bash
fly launch --dockerfile api/Dockerfile
fly secrets set DATABASE_URL=<your-postgres-url>
fly secrets set JWT_SECRET=<random-64-chars>
fly deploy
```

### Frontend (Vercel)
```bash
cd nextjs
vercel --prod
# Set NEXT_PUBLIC_API_URL to your Fly.io app URL
```

### Docker Compose (bare metal)
```bash
cp .env.prod.example .env.prod
# edit .env.prod
docker compose -f docker-compose.prod.yml up -d
```

## Project structure

```
darshan/
├── api/                          # Go backend
│   ├── cmd/
│   │   ├── server/main.go        # Entry point (chi router, graceful shutdown)
│   │   └── seed/main.go          # CLI seed tool
│   ├── internal/
│   │   ├── auth/jwt.go           # JWT generation and validation
│   │   ├── db/                   # Database connection + queries
│   │   │   ├── postgres.go       # pgx pool connection
│   │   │   ├── temples.go        # Temple CRUD queries
│   │   │   ├── users.go          # User auth queries
│   │   │   ├── favorites.go      # Favorites queries
│   │   │   ├── reviews.go        # Reviews queries
│   │   │   └── analytics.go      # Page view tracking
│   │   ├── handlers/             # HTTP handlers
│   │   │   ├── handler.go        # Handler struct
│   │   │   ├── health.go         # GET /health
│   │   │   ├── temples.go        # GET /temples, GET /temples/{slug}
│   │   │   ├── auth.go           # POST /auth/signup, /auth/login, /auth/me
│   │   │   ├── favorites.go      # CRUD /favorites
│   │   │   ├── reviews.go        # POST + GET /reviews
│   │   │   ├── admin.go          # POST + DELETE /admin/temples
│   │   │   └── analytics.go      # POST /views, GET /popular
│   │   ├── middleware/auth.go    # JWT auth middleware (Auth, OptionalAuth, Admin)
│   │   └── models/               # Data types (Temple, User, Review, Favorite)
│   ├── migrations/
│   │   ├── 001_initial.sql       # Temples + images tables
│   │   └── 002_users.sql         # Users, favorites, reviews, page_views
│   ├── seed/
│   │   ├── temples.json          # 20 initial temples
│   │   └── temples2.json         # 34 additional temples
│   └── Dockerfile
├── nextjs/                       # Next.js frontend
│   ├── src/
│   │   ├── app/
│   │   │   ├── page.tsx                 # Homepage — temple grid
│   │   │   ├── layout.tsx               # Root layout + AuthProvider
│   │   │   ├── globals.css              # Tailwind v4 + Leaflet overrides
│   │   │   ├── temples/[slug]/          # Temple detail page
│   │   │   │   ├── page.tsx             # SSR + generateMetadata
│   │   │   │   └── TempleDetailClient.tsx # 3D viewer, reviews, favorites
│   │   │   ├── map/                     # Full-screen Leaflet map
│   │   │   ├── auth/login/              # Sign in page
│   │   │   ├── auth/signup/             # Sign up page
│   │   │   ├── favorites/               # User's saved temples
│   │   │   ├── planner/                 # Visit planner
│   │   │   └── admin/add-temple/        # Admin form (role-gated)
│   │   ├── components/
│   │   │   ├── Navbar.tsx               # Sticky nav with auth state
│   │   │   ├── TempleCard.tsx           # Grid card
│   │   │   ├── TempleGallery.tsx        # Client search/filter + grid
│   │   │   ├── SearchFilters.tsx        # Search + state/deity filters
│   │   │   ├── Viewer3D.tsx             # 360°/3D embed + audio toggle
│   │   │   └── MapView.tsx              # Leaflet map (client-only)
│   │   └── lib/
│   │       ├── types.ts                 # TypeScript interfaces
│   │       ├── api.ts                   # API client (all endpoints)
│   │       └── AuthContext.tsx           # Auth state management
│   └── Dockerfile
├── docker-compose.yml            # Dev stack
├── docker-compose.prod.yml       # Production stack
├── fly.toml                      # Fly.io config
├── vercel.json                   # Vercel config
├── Makefile
└── README.md
```
