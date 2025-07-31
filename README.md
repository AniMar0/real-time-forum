# Real-Time Forum

A modern web-based forum with real-time chat, posts, and comments. Built with Go (backend), SQLite (database), and vanilla JavaScript (frontend).

## Features

- **User Registration & Login:** Secure authentication with sessions.
- **Forum Posts:** Create, view, and categorize posts.
- **Comments:** Add comments to posts.
- **Real-Time Chat:** Chat with other online users using WebSockets.
- **Online User List:** See who’s online and start chats.
- **Responsive UI:** Futuristic design, works on desktop and mobile.

## Project Structure

- `main.go` — Entry point for the Go backend server.
- `backend/` — Backend logic (database, handlers, server, tools, models).
- `database/forum.db` — SQLite database file.
- `static/` — Frontend static files (HTML, CSS, JS).
- `templates/` — HTML templates for error pages.

## Getting Started

### Prerequisites

- Go 1.24+
- SQLite3

### Installation

1. **Clone the repository:**
   ```sh
   git clone https://github.com/yourusername/real-time-forum.git
   cd real-time-forum
   ```

2. **Install dependencies:**
   ```sh
   go mod tidy
   ```

3. **Run the server:**
   ```sh
   go run main.go
   ```
   The server will start at [http://localhost:8080](http://localhost:8080).

4. **Open in browser:**
   Visit [http://localhost:8080](http://localhost:8080) to use the forum.

## Usage

- **Register:** Create a new account.
- **Login:** Access posts and chat features.
- **Create Posts:** Share your thoughts or questions.
- **Comment:** Engage in discussions.
- **Chat:** Click on an online user to start a real-time chat.

## Technologies Used

- **Backend:** Go, Gorilla WebSocket, SQLite
- **Frontend:** HTML, CSS, JavaScript

## License

MIT

---

Feel free to contribute or open issues for enhancements.