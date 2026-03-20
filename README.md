# 🚀 Real-Time Forum | Go & Vanilla JS SPA

A high-performance, real-time social platform built with **Go** (Standard Library) and **Vanilla JavaScript**. This project demonstrates a deep understanding of web fundamentals, concurrency, and real-time communication without relying on heavy frameworks.

## 🛠️ Tech Stack

- **Backend:** Go (Golang) - Pure Standard Library for Routing & Logic.
- **Real-Time:** WebSockets (Gorilla WebSocket).
- **Database:** SQLite3 with Foreign Key constraints.
- **Frontend:** Vanilla JS (Single Page Application architecture), CSS3, HTML5.
- **Security:** Bcrypt for password hashing, Cookie-based session management (UUID).

## 🌟 Key Features

- **Full SPA Experience:** Smooth navigation and page transitions without browser refreshes.
- **Real-Time Messaging:** Private chat system with instant delivery.
- **Typing Indicators:** Live "user is typing..." visual feedback via WebSockets.
- **Presence System:** Real-time online/offline status tracking.
- **Forum Core:** Post creation, categorization, and threaded comments.
- **Smart Loading:** Infinite scroll with throttling and debouncing for message history.

## 🏗️ Technical Highlights (The "Under the Hood")

### 1. Concurrency Management (Go)

The backend utilizes a robust **Hub/Client pattern**. Each connection is managed by dedicated **Goroutines** (Reader/Writer) and synchronized using **RWMutex** and **Buffered Channels** to prevent data races and ensure high throughput.

### 2. Custom SPA Router

Implemented a frontend "Router" from scratch that handles DOM manipulation and state management, mimicking modern framework behavior while maintaining a zero-dependency footprint.

### 3. Optimized Data Layer

- Complex SQLite schemas with `CHECK` constraints and `FOREIGN KEYS`.
- Paginated message retrieval to optimize memory and network usage.

## 📁 Architecture

- `/backend`: Core logic, WebSocket handling, and DB operations.
- `/static`: Frontend assets (ES Modules, CSS, and SPA logic).
- `/database`: SQLite schema definitions and migrations.

## 🚀 How to Run

1. Clone the repository: `git clone https://github.com/AniMar0/real-time-forum`
2. Install Go dependencies: `go mod tidy`
3. Run the server: `go run main.go`
4. Open `http://localhost:8080`
