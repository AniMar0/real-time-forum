# Real-time Forum

A small forum application with posts, comments, real-time direct messaging over WebSocket, and unread notifications.

See [Architecture](docs/architecture.md) for the system overview, package boundaries, dependency direction, data model, authentication flow, WebSocket design, and message transaction flow.

## Requirements

- Go 1.25 or later
- SQLite
- Docker and Docker Compose (optional)

## Run locally

```bash
go run .
```

Open:

```text
http://localhost:8080
```

The application creates the database directory and runs pending SQLite migrations automatically at startup.

## Run with Docker

```bash
docker compose up --build
```

Then open:

```text
http://localhost:8080
```

SQLite is stored in a Docker volume named `forum-data`, so data survives container recreation.

Stop the container with:

```bash
docker compose down
```

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `FORUM_HTTP_ADDRESS` | `:8080` | HTTP address and port |
| `FORUM_DATABASE_PATH` | `database/forum.db` | SQLite database path |
| `FORUM_STATIC_PATH` | `static` | Frontend files path |
| `FORUM_ENV` | `development` | Runtime environment; `production` enables secure cookies |
| `FORUM_WS_ORIGINS` | localhost origins | Allowed WebSocket origins, separated by commas |

Example:

```bash
FORUM_HTTP_ADDRESS=:9090 \
FORUM_DATABASE_PATH=database/forum.db \
FORUM_WS_ORIGINS=http://localhost:9090 \
go run .
```

## Current features

- Account registration and login using email or nickname.
- SQLite-backed login sessions.
- Create and view posts.
- Add and view comments.
- Real-time direct messaging over WebSocket.
- Persistent messages and unread notifications in one database transaction.
- Paginated chat history loading.
- Closeable, mobile-responsive chat interface.

## Main routes

| Route | Method | Purpose |
|---|---|---|
| `/register` | POST | Create an account |
| `/login` | POST | Log in |
| `/logout` | POST | Log out |
| `/logged` | POST | Check the current session |
| `/posts` | GET | Fetch posts |
| `/createPost` | POST | Create a post |
| `/comments` | GET | Fetch comments |
| `/createComment` | POST | Create a comment |
| `/messages` | POST | Fetch chat history |
| `/notifications` | GET | Fetch unread notifications |
| `/notifications/mark-read` | POST | Mark notifications as read |
| `/ws` | WebSocket | Messaging, presence, and typing events |

## Project structure

```text
.
├── backend/
│   ├── account/       # Accounts and sessions
│   ├── chat/          # Messages and chat history
│   ├── forum/         # Posts and comments
│   ├── notification/  # Unread notifications
│   └── migrations/    # SQLite migrations
├── static/            # HTML, CSS, and frontend JavaScript
├── docs/              # Architecture and project documentation
├── database/          # Local SQLite database
├── Dockerfile
└── docker-compose.yml
```

## Tests and checks

```bash
go test ./...
go test -race ./...
go vet ./...
go build .
```

The same checks run in `.github/workflows/ci.yml` on pushes and pull requests.

## Database notes

The local `database/forum.db` file is intended for local development. Docker uses a separate volume. Do not delete or replace the database while the application is running.
