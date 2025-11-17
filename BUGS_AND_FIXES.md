# 🐛 Bugs, Issues & Recommended Fixes

## 🔴 Critical Issues

### 1. WebSocket CORS Vulnerability

**File**: `backend/Server.go`  
**Line**: WebSocket upgrader initialization  
**Issue**: No origin check for WebSocket connections, allowing potential CSRF attacks.

**Current Code**:

```go
upgrader websocket.Upgrader
```

**Fix**:

```go
upgrader: websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        allowedOrigins := []string{"http://localhost:8080"} // Add your production domain
        for _, allowed := range allowedOrigins {
            if origin == allowed {
                return true
            }
        }
        return false
    },
}
```

---

### 2. Goroutine Leak - Send Channel Not Closed

**File**: `backend/Server.go`  
**Function**: `removeClient`  
**Issue**: The `Send` channel is never closed when removing a client, causing goroutine leaks.

**Current Code**:

```go
func (s *Server) removeClient(client *Client) {
    s.Lock()
    defer s.Unlock()

    clients, ok := s.clients[client.Username]
    if !ok {
        return
    }

    for i, c := range clients {
        if c.ID == client.ID {
            s.clients[client.Username] = append(clients[:i], clients[i+1:]...)
            break
        }
    }
    // Missing: close(c.Send) and c.Conn.Close()
}
```

**Fix**:

```go
func (s *Server) removeClient(client *Client) {
    s.Lock()
    defer s.Unlock()

    clients, ok := s.clients[client.Username]
    if !ok {
        return
    }

    for i, c := range clients {
        if c.ID == client.ID {
            s.clients[client.Username] = append(clients[:i], clients[i+1:]...)
            close(c.Send)  // Close the channel to terminate StartWriter goroutine
            c.Conn.Close() // Close WebSocket connection
            break
        }
    }

    if len(s.clients[client.Username]) == 0 {
        delete(s.clients, client.Username)
    }

    fmt.Println(client.Username, "disconnected")

    go func() {
        time.Sleep(100 * time.Millisecond)
        s.broadcastUserStatusChange()
    }()
}
```

---

### 3. SQL Scan Order Mismatch

**File**: `backend/Server.go`  
**Function**: `GetHashedPasswordFromDB`  
**Issue**: The order of variables in `Scan()` doesn't match the SELECT statement order.

**Current Code**:

```go
err := S.db.QueryRow(`
    SELECT password, nickname FROM users
    WHERE nickname = ? OR email = ?
`, identifier, identifier).Scan(&nickname, &hashedPassword)
```

**Fix**:

```go
err := S.db.QueryRow(`
    SELECT password, nickname FROM users
    WHERE nickname = ? OR email = ?
`, identifier, identifier).Scan(&hashedPassword, &nickname)  // Correct order
```

---

### 4. No Input Validation in Backend

**File**: `backend/handlers.go`  
**Multiple Functions**: `RegisterHandler`, `CreatePostHandler`, `CreateCommentHandler`  
**Issue**: Insufficient validation for:

- Text length limits
- Email format validation
- Password strength in backend
- Age validation

**Recommended Fix**:

```go
// Add validation package
package backend

import (
    "regexp"
    "strings"
    "unicode"
)

// Validate email format
func isValidEmail(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}

// Validate password strength
func isValidPassword(password string) bool {
    if len(password) < 8 {
        return false
    }
    var hasUpper, hasLower, hasNumber bool
    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        }
    }
    return hasUpper && hasLower && hasNumber
}

// Validate text length
func isValidTextLength(text string, min, max int) bool {
    length := len(strings.TrimSpace(text))
    return length >= min && length <= max
}

// Apply in RegisterHandler
func (S *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
    // ... existing code ...

    // Add validation
    if !isValidEmail(user.Email) {
        renderErrorPage(w, r, "Invalid email format", http.StatusBadRequest)
        return
    }

    if !isValidPassword(user.Password) {
        renderErrorPage(w, r, "Password must be at least 8 characters with uppercase, lowercase, and number", http.StatusBadRequest)
        return
    }

    if user.Age < 13 || user.Age > 120 {
        renderErrorPage(w, r, "Invalid age", http.StatusBadRequest)
        return
    }

    // ... rest of the code ...
}
```

---

## 🟠 Medium Priority Issues

### 5. Typo in Filename

**File**: `static/regester.js`  
**Issue**: Should be `register.js` (spelling error)

**Fix**: Rename file and update imports:

```bash
mv static/regester.js static/register.js
```

Update `app.js`:

```javascript
import { handleRegister } from "./register.js"; // Fixed spelling
```

---

### 6. Missing Import in register.js

**File**: `static/regester.js`  
**Issue**: `ErrorPage` function used but not imported

**Current Code**:

```javascript
// Missing import
export function handleRegister(event) {
  // ...
  ErrorPage(res); // Used but not imported
}
```

**Fix**:

```javascript
import { showSection } from "./app.js";
import { ErrorPage } from "./error.js"; // Add this import

export function handleRegister(event) {
  // ... rest of the code
}
```

---

### 7. Unclear Error Messages

**File**: `backend/handlers.go`  
**Function**: `LoginHandler`

**Current Code**:

```go
renderErrorPage(w, r, "User Undif", http.StatusBadRequest)
```

**Fix**:

```go
renderErrorPage(w, r, "User not found", http.StatusBadRequest)
```

**Also update**:

```go
renderErrorPage(w, r, "Inccorect password", http.StatusInternalServerError)
// Fix to:
renderErrorPage(w, r, "Incorrect password", http.StatusUnauthorized)
```

---

### 8. StartWriter Goroutine Not Handling Closed Channel

**File**: `backend/Server.go`  
**Function**: `StartWriter`  
**Issue**: If `Send` channel is closed, the function panics.

**Current Code**:

```go
func StartWriter(c *Client) {
    for msg := range c.Send {
        err := c.Conn.WriteJSON(msg)
        if err != nil {
            fmt.Println(err)
        }
    }
}
```

**Fix**:

```go
func StartWriter(c *Client) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Writer panic recovered:", r)
        }
    }()

    for msg := range c.Send {
        if err := c.Conn.WriteJSON(msg); err != nil {
            fmt.Println("Write error:", err)
            return // Exit on write error
        }
    }
}
```

---

## 🟡 Low Priority Issues

### 9. No Rate Limiting

**Issue**: No protection against brute force or spam attacks.

**Recommended Implementation**:

```go
import (
    "golang.org/x/time/rate"
    "sync"
)

type IPRateLimiter struct {
    ips map[string]*rate.Limiter
    mu  *sync.RWMutex
    r   rate.Limit
    b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
    return &IPRateLimiter{
        ips: make(map[string]*rate.Limiter),
        mu:  &sync.RWMutex{},
        r:   r,
        b:   b,
    }
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()

    limiter, exists := i.ips[ip]
    if !exists {
        limiter = rate.NewLimiter(i.r, i.b)
        i.ips[ip] = limiter
    }

    return limiter
}

// Apply in middleware
func (S *Server) RateLimitMiddleware(next http.Handler) http.Handler {
    limiter := NewIPRateLimiter(1, 5) // 1 request per second, burst of 5

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr
        limiter := limiter.GetLimiter(ip)

        if !limiter.Allow() {
            http.Error(w, "Too many requests", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

---

### 10. Console.log Statements in Production

**Files**: Multiple JavaScript files  
**Issue**: Debug logs left in production code

**Files to clean**:

- `static/app.js`
- `static/chat.js`
- `static/login.js`

**Recommended**: Create a logger utility:

```javascript
// static/logger.js
const isDevelopment = window.location.hostname === "localhost";

export const logger = {
  log: (...args) => {
    if (isDevelopment) {
      console.log(...args);
    }
  },
  error: (...args) => {
    console.error(...args); // Always log errors
  },
  warn: (...args) => {
    if (isDevelopment) {
      console.warn(...args);
    }
  },
};
```

---

### 11. No Database Connection Pooling

**File**: `backend/Server.go`  
**Issue**: Database connection not optimized for concurrent requests

**Current Code**:

```go
S.db, err = sql.Open("sqlite3", "database/forum.db")
```

**Fix**:

```go
S.db, err = sql.Open("sqlite3", "database/forum.db")
if err != nil {
    log.Fatal(err)
}

// Set connection pool limits
S.db.SetMaxOpenConns(25)
S.db.SetMaxIdleConns(5)
S.db.SetConnMaxLifetime(5 * time.Minute)
```

---

### 12. Missing Database Indexes

**File**: `backend/DataBase.go`  
**Issue**: No indexes for frequently queried columns

**Recommended Additions**:

```go
func createIndexes(db *sql.DB) error {
    indexes := []string{
        `CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,
        `CREATE INDEX IF NOT EXISTS idx_users_nickname ON users(nickname);`,
        `CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);`,
        `CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);`,
        `CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);`,
        `CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender);`,
        `CREATE INDEX IF NOT EXISTS idx_messages_receiver ON messages(receiver);`,
        `CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);`,
    }

    for _, index := range indexes {
        if _, err := db.Exec(index); err != nil {
            return err
        }
    }

    return nil
}

// Call in MakeDataBase after creating tables
if err := createIndexes(db); err != nil {
    log.Fatalf("Failed to create indexes: %v", err)
}
```

---

### 13. Error Response Inconsistency

**File**: `backend/handlers.go`  
**Issue**: Mix of `http.Error`, `renderErrorPage`, and JSON responses

**Recommended**: Standardize error responses:

```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
    Code    int    `json:"code"`
}

func sendJSONError(w http.ResponseWriter, message string, code int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(ErrorResponse{
        Error:   http.StatusText(code),
        Message: message,
        Code:    code,
    })
}
```

---

### 14. No Request Timeout

**File**: `backend/Server.go`  
**Function**: `Run`

**Current Code**:

```go
err = http.ListenAndServe(":"+port, S.Mux)
```

**Fix**:

```go
server := &http.Server{
    Addr:         ":" + port,
    Handler:      S.Mux,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}

fmt.Println("Server running on http://localhost:" + port)
err = server.ListenAndServe()
```

---

## 📋 Testing Recommendations

### Unit Tests Needed:

1. User registration validation
2. Password hashing and verification
3. Session creation and validation
4. Message sending and receiving
5. Post and comment creation

### Example Test:

```go
// backend/handlers_test.go
package backend

import (
    "testing"
    "golang.org/x/crypto/bcrypt"
)

func TestCheckPassword(t *testing.T) {
    password := "TestPass123"
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

    err := CheckPassword(string(hashedPassword), password)
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }

    err = CheckPassword(string(hashedPassword), "WrongPassword")
    if err == nil {
        t.Error("Expected error for wrong password")
    }
}
```

---

## 🎯 Priority Fixes Summary

### Must Fix (Critical):

1. ✅ Add WebSocket origin check
2. ✅ Close Send channels properly
3. ✅ Fix SQL Scan order
4. ✅ Add backend input validation

### Should Fix (Medium):

5. ✅ Rename regester.js to register.js
6. ✅ Add missing imports
7. ✅ Fix error message typos
8. ✅ Handle closed channels in StartWriter

### Nice to Have (Low):

9. ⚠️ Implement rate limiting
10. ⚠️ Remove debug console.logs
11. ⚠️ Add database connection pooling
12. ⚠️ Create database indexes
13. ⚠️ Standardize error responses
14. ⚠️ Add request timeouts

---

**Total Issues Found**: 14  
**Critical**: 4  
**Medium**: 4  
**Low**: 6

---

Last Updated: November 17, 2025
