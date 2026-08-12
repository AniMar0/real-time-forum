# Architecture

This document describes the current architecture of the Real-time Forum after the refactor and cleanup pass.

## System overview

The application is a small Go HTTP server with a static JavaScript frontend, SQLite persistence, and a WebSocket endpoint for real-time chat and presence updates.

```mermaid
flowchart LR
    Browser[Browser frontend\nstatic/*.js + style.css]
    HTTP[HTTP handlers\nbackend/handlers.go]
    WS[WebSocket lifecycle\nbackend/Server.go]
    Hub[WebSocket Hub\nbackend/Hub.go]
    Account[account package\nusers + sessions]
    Forum[forum package\nposts + comments]
    Chat[chat package\nmessages + conversations]
    Notification[notification package\nunread counters]
    SQLite[(SQLite database)]

    Browser -->|HTTP| HTTP
    Browser -->|WebSocket| WS
    WS --> Hub
    WS --> Chat
    WS --> Notification
    HTTP --> Account
    HTTP --> Forum
    HTTP --> Chat
    HTTP --> Notification
    Account --> SQLite
    Forum --> SQLite
    Chat --> SQLite
    Notification --> SQLite
```

## Package responsibilities

### Root `backend` package

The root package owns application wiring and transport orchestration:

- HTTP route registration and middleware.
- WebSocket upgrade and connection lifecycle.
- Mapping authenticated identities to transport clients.
- Routing WebSocket events to the chat service and Hub.
- Converting repository results into HTTP/WebSocket response shapes.

It should not contain feature-specific SQL or persistence rules.

### `backend/account`

Owns account and session persistence:

- `UserRepository`: user existence checks, account creation, and credential lookup.
- `SessionRepository`: create, validate, and delete sessions.
- `Identity`: authenticated `UserID`, nickname, and session ID passed through request context.

### `backend/forum`

Owns forum persistence:

- Creating and listing posts.
- Creating and listing comments.
- Post and comment data models.

### `backend/chat`

Owns chat persistence and business rules:

- Validating recipients and message content.
- Resolving recipient nicknames to user IDs.
- Persisting messages.
- Listing message history and conversation users.
- Coordinating message persistence with unread notification updates through one transaction.

### `backend/notification`

Owns unread notification persistence:

- Incrementing unread message counters.
- Listing unread counters by sender.
- Marking a sender conversation as read.

### `backend/migrations`

Contains the ordered SQLite migrations. Migrations `003` and `004` represent the identity migration from nickname relationships to user IDs. The current schema uses IDs for messages, sessions, and notifications.

## Dependency direction

```mermaid
flowchart TD
    Root[backend transport/wiring]
    Account[account]
    Forum[forum]
    Chat[chat]
    Notification[notification]
    DB[(SQLite)]

    Root --> Account
    Root --> Forum
    Root --> Chat
    Root --> Notification
    Chat --> Notification
    Account --> DB
    Forum --> DB
    Chat --> DB
    Notification --> DB
```

The direction is intentionally simple:

- Handlers depend on repositories and services.
- `chat.Service` depends on `chat.Repository` and `notification.Repository` because sending a message updates both message history and unread state atomically.
- Repositories depend on `database/sql` and do not depend on HTTP or WebSocket code.
- Feature packages do not depend on the root `backend` package.

## Authentication and session flow

```mermaid
sequenceDiagram
    participant B as Browser
    participant H as HTTP handler
    participant A as account.UserRepository
    participant S as account.SessionRepository
    participant DB as SQLite

    B->>H: POST /register
    H->>H: Validate nickname, email, password, age, names, gender
    H->>A: Exists + Create
    A->>DB: Check uniqueness + insert hashed password
    DB-->>A: Result
    A-->>H: Result
    H-->>B: Registration response

    B->>H: POST /login
    H->>A: PasswordByIdentifier
    A->>DB: Load password hash and nickname
    DB-->>A: Credentials
    H->>H: Compare password hash
    H->>S: Create session
    S->>DB: Insert session with user_id and expiry
    H-->>B: HttpOnly session cookie

    B->>H: Authenticated request
    H->>S: FindValid(session cookie)
    S->>DB: Validate session and resolve user identity
    S-->>H: account.Identity
    H->>B: Continue with identity in request context
```

The authenticated identity is the source of truth. Client-provided sender identity is not trusted for protected operations.

## WebSocket Hub and client design

The Hub is keyed by `UserID`, not nickname:

```text
Hub
└── map[userID]map[connectionID]*Client
```

This supports multiple browser sessions for one user and avoids using mutable display data as an internal identity key.

Each `Client` owns:

- The WebSocket connection.
- The authenticated `UserID`, nickname, and session ID.
- A buffered outbound channel.
- Close synchronization to avoid closing the channel more than once.

Each connection has two goroutines:

- Reader: applies read limits, pong deadlines, parses incoming events, and delegates chat handling.
- Writer: serializes outbound messages and periodic ping frames with write deadlines.

The browser still receives nicknames in the existing message contract. IDs are used internally for Hub lookup, persistence, presence, and delivery.

## Message persistence and notification transaction

```mermaid
sequenceDiagram
    participant C as WebSocket client
    participant R as receiveMessages
    participant S as chat.Service
    participant CR as chat.Repository
    participant NR as notification.Repository
    participant DB as SQLite transaction
    participant H as Hub

    C->>R: chat_message {to, content}
    R->>S: SendMessage(authenticated senderID, receiver nickname, content)
    S->>S: Validate sender, recipient, and content
    S->>CR: Resolve receiver ID
    S->>DB: Begin transaction
    S->>CR: Insert message using sender_id/receiver_id
    S->>NR: Increment unread counter
    S->>DB: Commit
    S-->>R: Stored message with IDs and display names
    R->>H: Deliver to receiver sessions and sender sessions
    R->>H: Refresh presence/conversation lists
```

If message insertion or notification update fails, the transaction rolls back and no partial message state is committed.

Messages are stored as raw text and rendered by the frontend with `textContent`. This keeps storage separate from presentation escaping and avoids double-escaping history or live events.

## Data model

The current identity relationships are:

```mermaid
erDiagram
    USERS ||--o{ POSTS : creates
    USERS ||--o{ COMMENTS : writes
    USERS ||--o{ MESSAGES : sends
    USERS ||--o{ MESSAGES : receives
    USERS ||--o{ SESSIONS : owns
    USERS ||--o{ NOTIFICATIONS : receives
    USERS ||--o{ NOTIFICATIONS : triggers
    POSTS ||--o{ COMMENTS : contains

    USERS {
        int id PK
        string nickname UK
        string email UK
        string password
        string first_name
        string last_name
        int age
        string gender
    }
    POSTS {
        int id PK
        int user_id FK
        string title
        string content
        string category
        datetime created_at
    }
    COMMENTS {
        int id PK
        int post_id FK
        int user_id FK
        string content
        datetime created_at
    }
    MESSAGES {
        int id PK
        int sender_id FK
        int receiver_id FK
        string content
        datetime timestamp
    }
    SESSIONS {
        string session_id PK
        int user_id FK
        datetime expires_at
    }
    NOTIFICATIONS {
        int id PK
        int receiver_id FK
        int sender_id FK
        int unread_messages
    }
```

## Important boundaries

- HTTP handlers authenticate and validate transport input, then call a repository or service.
- `account.Identity` is used for authenticated user context.
- `chat.Service` is the only business entry point for sending a message.
- The Hub handles connection lookup and delivery, not database persistence.
- Repositories own SQL queries.
- SQLite migrations own schema changes; runtime code should not silently alter the schema.

## Verification commands

```bash
gofmt -l .
go test ./...
go test -race ./...
go vet ./...
go build .
```
