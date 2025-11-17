# Real-Time Forum

A real-time forum application with instant messaging capabilities, built with Go and vanilla JavaScript.

## 🚀 Features

- **User Authentication**: Register, login, and session management with bcrypt password hashing
- **Posts & Comments**: Create and view posts with threaded comments
- **Real-time Chat**: WebSocket-based instant messaging with online/offline status
- **Notifications**: Unread message badges and real-time updates
- **Multi-session Support**: Users can be connected from multiple devices simultaneously
- **Message History**: Pagination for loading previous messages

## 🛠️ Technologies

### Backend

- **Go 1.24.1**
- **SQLite3** - Database
- **Gorilla WebSocket** - Real-time communication
- **bcrypt** - Password hashing
- **UUID** - Session management

### Frontend

- **Vanilla JavaScript** (ES6 Modules)
- **HTML5** & **CSS3**
- **WebSocket API**

## 📁 Project Structure

```
real-time-forum/
├── backend/
│   ├── DataBase.go      # Database initialization and schema
│   ├── handlers.go      # HTTP request handlers
│   ├── Objects.go       # Data structures
│   ├── Server.go        # Server setup and WebSocket handling
│   └── Tools.go         # Helper functions
├── database/            # SQLite database (auto-generated)
├── static/              # Frontend files
│   ├── app.js          # Main application logic
│   ├── chat.js         # Real-time chat features
│   ├── comments.js     # Comment system
│   ├── error.js        # Error page handler
│   ├── login.js        # Login functionality
│   ├── logout.js       # Logout functionality
│   ├── posts.js        # Post management
│   ├── regester.js     # User registration
│   ├── style.css       # Styling
│   └── index.html      # Main HTML page
├── main.go             # Entry point
├── go.mod              # Go dependencies
├── PROJECT_ANALYSIS.md # Comprehensive project analysis (Arabic)
└── BUGS_AND_FIXES.md   # Detailed bug report and fixes

```

## 🗄️ Database Schema

### Tables:

- **users** - User accounts (nickname, email, password, age, gender)
- **posts** - Forum posts with categories
- **comments** - Comments on posts
- **messages** - Private messages between users
- **sessions** - User session management
- **notifications** - Unread message notifications

## 📥 Installation

### Prerequisites

- Go 1.24 or higher
- Modern web browser

### Steps

1. **Clone the repository**

```bash
git clone <repository-url>
cd real-time-forum
```

2. **Install Go dependencies**

```bash
go mod download
```

3. **Build the project**

```bash
go build -o real-time-forum.exe .
```

4. **Run the server**

```bash
./real-time-forum.exe
# or
go run main.go
```

5. **Open in browser**

```
http://localhost:8080
```

## 🔧 Configuration

The server runs on port **8080** by default. To change the port, modify `main.go`:

```go
func main() {
    var Server backend.Server
    backend.MakeDataBase()
    Server.Run("8080") // Change port here
}
```

## 📝 API Endpoints

### Authentication

- `POST /register` - Register new user
- `POST /login` - User login
- `POST /logout` - User logout
- `POST /logged` - Check session status

### Posts & Comments

- `GET /posts` - Get all posts
- `POST /createPost` - Create new post (authenticated)
- `GET /comments?post_id={id}` - Get comments for a post
- `POST /createComment` - Add comment (authenticated)

### Messaging

- `POST /sendMessage` - Send message (authenticated)
- `POST /messages?from={user}&to={user}` - Get message history
- `POST /notification` - Update notification status
- `GET /ws` - WebSocket connection (authenticated)

## 🐛 Known Issues

See [BUGS_AND_FIXES.md](BUGS_AND_FIXES.md) for detailed information about:

- 4 Critical security and stability issues
- 4 Medium priority improvements
- 6 Low priority enhancements

## 📊 Testing

The project currently builds without compilation errors. To test:

```bash
# Build and check for errors
go build -o real-time-forum.exe .

# Run the application
./real-time-forum.exe
```

Visit `http://localhost:8080` and test:

1. User registration and login
2. Creating posts and comments
3. Real-time chat functionality
4. Multi-device sessions
5. Online/offline status updates

## 🔐 Security Features

- ✅ bcrypt password hashing
- ✅ Session-based authentication with expiration (24 hours)
- ✅ HTML escaping to prevent XSS attacks
- ✅ SQL parameterized queries
- ✅ Foreign key constraints
- ⚠️ CORS protection needed for WebSocket
- ⚠️ Rate limiting recommended

## 🎯 Future Improvements

- [ ] Add input validation in backend
- [ ] Implement rate limiting
- [ ] Add database indexes for performance
- [ ] Create comprehensive test suite
- [ ] Add HTTPS support
- [ ] Implement file upload for posts
- [ ] Add user profiles and avatars
- [ ] Email verification system
- [ ] Password reset functionality

## 📖 Documentation

- **[PROJECT_ANALYSIS.md](PROJECT_ANALYSIS.md)** - Comprehensive analysis in Arabic
- **[BUGS_AND_FIXES.md](BUGS_AND_FIXES.md)** - Detailed bug reports and solutions

## 📄 License

This project is available for educational purposes.

## 👥 Contributing

Issues and pull requests are welcome!

---

**Status**: ✅ Working - Build successful  
**Version**: 1.0  
**Last Updated**: November 17, 2025
