Role: You are generating a complete, production-ready codebase.

Goal: Implement a full web forum in Go with SQLite, strictly no JavaScript and no frontend frameworks. Only HTML, CSS, Go, SQLite. Include Docker support, unit tests, and a polished neon dark theme using only CSS.

Technical Constraints

Languages: Go, HTML, CSS, SQL (SQLite)

No JavaScript of any kind.

Allowed Go packages:

Standard library

github.com/mattn/go-sqlite3 (CGO OK)

golang.org/x/crypto/bcrypt or github.com/google/uuid / github.com/gofrs/uuid (use UUID for sessions)

Must compile and run with Go 1.22+.

Must run in Docker (provide a Dockerfile and the run commands).

No binary assets (no images/video/audio).

Server-rendered HTML using html/template.

Proper HTTP status codes and error pages (400/404/500).

Unit tests using testing and net/http/httptest.

Functional Requirements
Authentication & Sessions

Registration requires email, username, password.

Detect duplicate email and duplicate username and return clear errors.

Login checks credentials; wrong email/password returns a clear error response.

Passwords stored hashed with bcrypt.

Sessions via secure cookie (HttpOnly). Store session rows in DB:

id (uuid), user_id, created_at, expires_at, revoked_at NULL.

Enforce single active session per user: when a user logs in, revoke any existing active session and issue a new one.

Cookie must include expiry aligned to expires_at (e.g., 24h).

Middleware that fetches session by cookie, rejects if missing, expired, or revoked.

Non-registered users can browse but cannot post/comment/like.

Posts, Comments, Categories

Registered users can create posts (title + body) and associate one or more categories per post.

Registered users can create comments on posts.

Empty title/body/comments are rejected with validation messages.

Posts and comments visible to everyone.

Categories are predefined in DB and can be multi-select on the post form.

Filtering:

By category (e.g., /filter?category=golang).

My Posts (created by the logged-in user).

My Liked Posts (liked by the logged-in user).

Likes/Dislikes:

Registered users can like or dislike posts and comments.

Show counts to all users.

A user cannot like and dislike the same entity simultaneously.

Toggling:

If user likes something they previously disliked, switch to like.

If user clicks like again, it should remove the like (same for dislike).

All mutating actions via POST forms (no JS). Use method="post".

Error Handling & Robustness

Proper HTTP methods: GET for pages, POST for actions.

Return 400 on bad input; 500 on internal errors; 404 for not found.

Central error pages/templates: /error/400, /error/404, /error/500 (or render inline).

Graceful DB errors with user-friendly messages.

Unit tests: at least for registration/login, session middleware, creating posts/comments, like/dislike toggling, and key handlers.

SQLite & Queries

Use SQLite and include migrations (Go init() or explicit SQL files executed on start).

Must include at least:

One CREATE TABLE …

One INSERT …

One SELECT …

Provide a CLI/dev instruction to inspect DB with sqlite3 and run:

SELECT * FROM users;

SELECT * FROM posts;

SELECT * FROM comments;

Docker

Provide a Dockerfile (multi-stage recommended).

The image must run the server on port 8080.

Data stored in /app/data/forum.db (mountable volume).

Commands documented to build image and run container (and a simple shell script to build & run).

Project Organization

Create this structure:

forum/
  cmd/server/
    main.go
  internal/
    app/
      app.go            // server wiring: routes, middleware, templates
      routes.go
      middleware.go
      validators.go
      render.go
      errors.go
    db/
      schema.sql        // CREATE TABLEs, indexes, seed categories
      migrate.go        // run migrations at startup
      queries.go        // helper CRUD methods with prepared statements
      models.go         // Go structs (User, Session, Post, Comment, Category, Vote)
    auth/
      sessions.go       // session cookie + store logic
      password.go       // bcrypt helpers
    handlers/
      auth.go           // register/login/logout
      posts.go          // list/create/view/filter posts
      comments.go       // add comments, list
      votes.go          // like/dislike for posts and comments
      me.go             // my posts, my likes
  web/
    templates/
      layout.tmpl
      partials/
        flash.tmpl
        nav.tmpl
        pagination.tmpl
      auth/
        register.tmpl
        login.tmpl
      posts/
        index.tmpl
        show.tmpl
        new.tmpl
        filter.tmpl
        mine.tmpl
        liked.tmpl
      comments/
        list.tmpl
      error/
        400.tmpl
        404.tmpl
        500.tmpl
    static/
      css/
        base.css
        theme-dark-neon.css
  Dockerfile
  go.mod
  go.sum
  README.md
  Makefile                 // optional: build, run, test
  scripts/
    docker_build.sh
    docker_run.sh
  tests/
    auth_test.go
    posts_test.go
    votes_test.go

Database Schema (DDL)

Implement (and index) at minimum:

-- users
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT NOT NULL UNIQUE,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- sessions (single active per user)
CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,                 -- uuid
  user_id INTEGER NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME NOT NULL,
  revoked_at DATETIME,
  FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_user_active
  ON sessions(user_id) WHERE revoked_at IS NULL;

-- categories (seed with a few defaults)
CREATE TABLE IF NOT EXISTS categories (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL UNIQUE
);

-- posts
CREATE TABLE IF NOT EXISTS posts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

-- post_categories (many-to-many)
CREATE TABLE IF NOT EXISTS post_categories (
  post_id INTEGER NOT NULL,
  category_id INTEGER NOT NULL,
  PRIMARY KEY (post_id, category_id),
  FOREIGN KEY (post_id) REFERENCES posts(id),
  FOREIGN KEY (category_id) REFERENCES categories(id)
);

-- comments
CREATE TABLE IF NOT EXISTS comments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  post_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  body TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (post_id) REFERENCES posts(id),
  FOREIGN KEY (user_id) REFERENCES users(id)
);

-- votes for posts
CREATE TABLE IF NOT EXISTS post_votes (
  user_id INTEGER NOT NULL,
  post_id INTEGER NOT NULL,
  value INTEGER NOT NULL CHECK (value IN (-1,1)), -- dislike=-1, like=+1
  PRIMARY KEY (user_id, post_id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (post_id) REFERENCES posts(id)
);

-- votes for comments
CREATE TABLE IF NOT EXISTS comment_votes (
  user_id INTEGER NOT NULL,
  comment_id INTEGER NOT NULL,
  value INTEGER NOT NULL CHECK (value IN (-1,1)),
  PRIMARY KEY (user_id, comment_id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (comment_id) REFERENCES comments(id)
);


Seed categories (example): golang, databases, docker, web, security.

HTTP Routes (no JS; form-based)

Public:

GET / — home (list posts, filter form by category)

GET /post/{id} — post details with comments and like/dislike counts

GET /login GET /register

GET /error/400, /error/404, /error/500 (or render from handler)

Auth:

POST /register — validate, create user, hash password, login

POST /login — validate, create session, set cookie (HttpOnly, SameSite=Lax)

POST /logout — revoke session, clear cookie

Posts:

GET /post/new — form (auth required)

POST /post — create post (auth required, validate non-empty)

POST /post/{id}/comment — create comment (auth required, validate)

POST /post/{id}/like — like toggle

POST /post/{id}/dislike — dislike toggle

Comments:

POST /comment/{id}/like

POST /comment/{id}/dislike

Filters:

GET /filter — by ?category=slug

GET /me/posts — posts by current user

GET /me/likes — posts liked by current user

All mutating endpoints must verify session and return 401/403 when missing.

Template & UX Requirements (CSS-only, no JS)

Theme: dark with neon accents — black base with yellow, green, orange, purple highlights.

Background: near-black #0b0f14 or #0a0a0a.

Text: #e6f1ff for body; muted secondary #94a3b8.

Accent palette:

Yellow: #ffd400

Green: #00ff95

Orange: #ff7a18

Purple: #9b5cff

Neon effects (subtle): box-shadow: 0 0 8px var(--accent).

Use pure CSS for:

Buttons: .btn, .btn--primary, .btn--danger, .btn--ghost (hover: glow).

Cards: .card with rounded corners, soft neon border.

Forms: .field, .label, .input, .help with focus outline neon.

Layout: centered max-width container, responsive using flex/grid (no floats).

Nav bar with active state, login/logout buttons, user badge.

Badges for categories (chips with neon border).

Like/Dislike as form buttons with counts (no JS).

Typography: system sans or Inter fallback stack; large headings, tight spacing.

Accessibility: high contrast; focus states; larger hit areas.

Provide two CSS files:

base.css — CSS reset + utilities + layout + components.

theme-dark-neon.css — variables + theme colors + neon effects.

Validation Rules

Registration:

Email required, valid email shape.

Username 3–24 chars, no spaces.

Password 8+ chars.

Duplicate email/username → show field-level error.

Post:

Title 3–120 chars, Body 3+ chars, ≥1 category.

Comment:

Body 1+ char (non-empty).

Return 400 on validation errors with error template showing messages.

Query Examples (must exist in code)

CREATE: tables above in schema.sql.

INSERT: user, post, comment, votes, etc.

SELECT:

List posts with joined like/dislike counts.

Get single post with categories and aggregated comment counts.

Filter by category (JOIN post_categories).

“My Posts” by user_id.

“My Likes” via post_votes where value=1.

Unit Tests (add in tests/)

auth_test.go:

Register flow success, duplicate email/username.

Login wrong password → 400 with error message.

Single-session enforcement (second login revokes first).

posts_test.go:

Create post requires auth.

Empty title/body rejected (400).

Filter by category returns expected post.

votes_test.go:

Like a post increments count.

Like then dislike switches state (never both set).

Re-click like removes like.

Use httptest.Server, temporary SQLite DB (file-backed under a temp dir) and clean up.

Docker

Create a multi-stage Dockerfile:

Stage 1: build Go binary server.

Stage 2: minimal base with CA certs, copy /app/server, /app/web, /app/internal/db/schema.sql, create /app/data for forum.db.

Expose port 8080 and run ./server.

Provide scripts:

scripts/docker_build.sh:

#!/usr/bin/env bash
set -euo pipefail
docker build -t forum:latest .
docker images | head -n 20


scripts/docker_run.sh:

#!/usr/bin/env bash
set -euo pipefail
mkdir -p data
docker run --rm -d \
  -p 8080:8080 \
  -v "$(pwd)/data:/app/data" \
  --name forum \
  forum:latest
docker ps -a


Document manual commands in README:

Build: docker build -t forum:latest .

Run: docker run --rm -d -p 8080:8080 -v "$(pwd)/data:/app/data" --name forum forum:latest

README.md

Generate a comprehensive README that includes:

Overview & features.

Tech stack & constraints (no JS).

How to run locally (go run ./cmd/server).

Docker build/run instructions.

Database path & how to inspect with sqlite3.

Routes overview.

Styling/theme notes.

Test instructions go test ./....

Audit checklist mapping to features (see below).

Troubleshooting.

Security & Good Practices

Use bcrypt with cost 12+.

Set cookies HttpOnly, SameSite=Lax. (Secure flag when behind TLS is a note.)

Input validation & context timeouts on DB calls.

Prepared statements for all inserts/updates.

Template auto-escaping enabled.

Deliverables

Full working codebase as per structure above.

No binary assets.

Polished CSS-only neon dark theme.

Passing go build and go test ./....

Dockerfile + scripts.

README with audit checklist.

Now implement the full project accordingly.

🧾 Project Description (copy-paste if needed)

A server-rendered forum in Go with SQLite and zero JavaScript. Users can register, log in, create posts with multi-category tagging, comment, and like/dislike both posts and comments. Non-authenticated users can browse and view counts but cannot interact.

Authentication uses bcrypt-hashed passwords and cookie-based sessions with UUID IDs stored in SQLite. Only one active session per user is enforced; the most recent login revokes older sessions. The UI is a responsive, accessible neon-dark theme (black base with yellow, green, orange, purple accents) implemented purely in CSS.

The app supports filtering by category, showing the current user’s posts, and showing posts the user liked. All mutations are handled through POST forms; there is no client-side JavaScript. Proper HTTP statuses and error pages are provided. The project includes unit tests, a multi-stage Dockerfile, and simple build/run scripts.

✅ Audit Checklist (ready to use)

Authentication

 Registration asks for email, username, password.

 Duplicate email/username returns error.

 Wrong email/password shows error.

 Registration works end to end.

 Login works; registered rights available.

 Login without credentials shows warning.

 Sessions exist with expiry.

 Single active session per user enforced across browsers.

 Creating post/comment in one browser appears in another.

SQLite

 At least one CREATE, INSERT, SELECT in code.

 After registering, SELECT * FROM users; shows new user.

 After posting, SELECT * FROM posts; shows new post.

 After commenting, SELECT * FROM comments; shows comment.

Docker

 Dockerfile exists.

 docker build -t forum:latest . succeeds; docker images lists image.

 docker run -p 8080:8080 forum:latest runs; docker ps -a shows container.

 No unused objects left in repo.

Functional

 Non-registered cannot create post/comment/like/dislike.

 Registered can create comment; empty comment forbidden.

 Registered can create post; empty post forbidden.

 Multiple categories selectable for a post.

 Can like/dislike posts and comments.

 Refresh updates like/dislike counts.

 Cannot like and dislike the same item simultaneously.

 “My Posts” shows expected posts.

 “My Likes” shows expected posts.

 Everyone sees like/dislike counts on comments.

 Category filter shows correct posts.

 Server stable (no crashes).

 Correct HTTP methods used.

 No broken pages (404 only where appropriate).

 Handles 400 & 500 with proper pages.

 Only allowed packages used.

 Meets auditor standards (not empty/incomplete/invalid/cheating/crashing/leaking).

General (+)

 Build & run scripts provided.

 Passwords encrypted in DB (bcrypt).

 Efficient handlers and DB usage; good practices.

 Unit tests included.

 Suitable to open-source; reusable.

 Worthy as an example project.

📘 README.md (drop-in draft)
# Forum (Go + SQLite, No JS)

A server-rendered web forum built with Go, SQLite, HTML, and CSS — **no JavaScript**. Users can register, log in, create posts with categories, comment, and like/dislike posts and comments. Non-authenticated users can browse and see counts but cannot interact.

## Features

- Registration & Login (email, username, password)
- Bcrypt password hashing
- Cookie sessions with UUID; **single active session per user**
- Create posts with **multiple categories**
- Comment on posts
- Like/Dislike for **posts and comments**
- Filters: by category, **My Posts**, **My Liked Posts**
- Fully server-rendered (no JS); POST forms for all actions
- Robust error handling (400/404/500)
- Polished dark neon theme (CSS-only)

## Tech Stack & Constraints

- Go 1.22+, `net/http`, `html/template`, `database/sql`
- SQLite via `github.com/mattn/go-sqlite3`
- Passwords via `golang.org/x/crypto/bcrypt`
- UUID via `github.com/google/uuid` (or gofrs/uuid)
- **No JavaScript** and no frontend frameworks
- Dockerized

## Getting Started (Local)

```bash
# 1) Run migrations and start the server
go run ./cmd/server

# Server defaults
# PORT: 8080
# DB:   ./data/forum.db (created on first run)


Visit: http://localhost:8080

Inspect the Database
sqlite3 ./data/forum.db
sqlite> .tables
sqlite> SELECT * FROM users;
sqlite> SELECT * FROM posts;
sqlite> SELECT * FROM comments;

Docker
Build the image
docker build -t forum:latest .
docker images

Run the container
mkdir -p data
docker run --rm -d \
  -p 8080:8080 \
  -v "$(pwd)/data:/app/data" \
  --name forum \
  forum:latest
docker ps -a


Or use scripts:

./scripts/docker_build.sh
./scripts/docker_run.sh

Project Structure
cmd/server/           # main entry
internal/
  app/                # wiring, middleware, render, errors
  auth/               # sessions, bcrypt
  db/                 # schema, migrations, queries
  handlers/           # route handlers
web/
  templates/          # HTML templates
  static/css/         # CSS files (base + dark neon theme)
scripts/              # docker scripts
tests/                # unit tests

Styling (Dark Neon, CSS-only)

Background: #0b0f14

Text: #e6f1ff, secondary #94a3b8

Accents: yellow #ffd400, green #00ff95, orange #ff7a18, purple #9b5cff

Components: .btn, .card, .chip, .input, .label, .help

Hover/focus: subtle neon glow via box-shadow: 0 0 8px var(--accent)

Files:

web/static/css/base.css

web/static/css/theme-dark-neon.css

HTTP Routes

Public: /, /post/{id}, /login, /register

Auth POST actions: /login, /register, /logout

Posts: GET /post/new, POST /post

Comments: POST /post/{id}/comment

Votes: POST /post/{id}/like, POST /post/{id}/dislike,
POST /comment/{id}/like, POST /comment/{id}/dislike

Filters: GET /filter?category=slug, GET /me/posts, GET /me/likes

Sessions

Cookie: HttpOnly, SameSite=Lax; aligned with DB expires_at

DB table sessions with UUID primary key

Only one active session per user; new login revokes prior session(s)

Tests

Run all tests:

go test ./...


Includes tests for:

Registration/login (including duplicates and wrong passwords)

Session middleware and single-session rule

Creating posts/comments with validation

Like/dislike toggling logic

Audit Checklist

See AUDIT.md or the section below; the app is designed to satisfy all items:

Auth validation and error paths

SQLite presence of CREATE/INSERT/SELECT

Dockerfile build/run

Functional permissions & filters

HTTP methods & error handling

Allowed packages only

Scripts, bcrypt encryption, good practices, unit tests

Troubleshooting

CGO/SQLite: Ensure your environment can compile github.com/mattn/go-sqlite3. Docker build handles this.

Port in use? Change port with PORT env or stop the conflicting service.

DB locked: avoid running multiple instances writing to the same file.

License

MIT


---

# 🎨 CSS Style Notes for Codex (copy to your CSS files)

Use custom properties in `theme-dark-neon.css`:

```css
:root {
  --bg: #0b0f14;
  --bg-soft: #0f1420;
  --text: #e6f1ff;
  --muted: #94a3b8;
  --edge: #1f2937;

  --yellow: #ffd400;
  --green:  #00ff95;
  --orange: #ff7a18;
  --purple: #9b5cff;

  --radius: 14px;
  --shadow: 0 0 8px;
}
body {
  background: var(--bg);
  color: var(--text);
}
a { color: var(--purple); text-decoration: none; }
a:hover { text-decoration: underline; }

.container { max-width: 1100px; margin: 0 auto; padding: 24px; }

.card {
  background: var(--bg-soft);
  border: 1px solid var(--edge);
  border-radius: var(--radius);
  padding: 20px;
  box-shadow: 0 0 0 transparent;
}
.card:hover { box-shadow: var(--shadow) rgba(155,92,255,0.35); }

.btn {
  display: inline-block;
  padding: 10px 16px;
  border-radius: 12px;
  border: 1px solid var(--edge);
  font-weight: 600;
  background: transparent;
  color: var(--text);
}
.btn--primary { border-color: var(--green); }
.btn--primary:hover { box-shadow: var(--shadow) var(--green); }
.btn--danger  { border-color: var(--orange); }
.btn--danger:hover { box-shadow: var(--shadow) var(--orange); }

.input, select, textarea {
  width: 100%;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid var(--edge);
  background: #0c111b;
  color: var(--text);
}
.input:focus, select:focus, textarea:focus {
  outline: none;
  border-color: var(--purple);
  box-shadow: var(--shadow) rgba(155,92,255,0.4);
}

.chip {
  display: inline-block;
  padding: 4px 10px;
  border: 1px solid var(--edge);
  border-radius: 999px;
  margin-right: 8px;
}
.chip--yellow { border-color: var(--yellow); box-shadow: var(--shadow) rgba(255,212,0,0.25); }
.chip--green  { border-color: var(--green);  box-shadow: var(--shadow) rgba(0,255,149,0.25); }
.chip--orange { border-color: var(--orange); box-shadow: var(--shadow) rgba(255,122,24,0.25); }
.chip--purple { border-color: var(--purple); box-shadow: var(--shadow) rgba(155,92,255,0.25); }

.help { color: var(--muted); font-size: 0.9rem; }
.error { color: var(--orange); }

🧰 Dockerfile Notes for Codex

Ensure a multi-stage build, example outline:

# ---- build ----
FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

# ---- final ----
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/server ./server
COPY web ./web
COPY internal/db/schema.sql ./internal/db/schema.sql
ENV PORT=8080
ENV DB_PATH=/app/data/forum.db
EXPOSE 8080
# create data dir at runtime if not present
ENTRYPOINT ["./server"]


(If distroless is an issue for sqlite3 CGO runtime, use a minimal Debian/Alpine with glibc as needed. Ensure it runs successfully.)
