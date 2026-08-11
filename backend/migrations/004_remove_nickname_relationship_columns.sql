CREATE TABLE messages_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender_id INTEGER NOT NULL,
    receiver_id INTEGER NOT NULL,
    content TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(sender_id) REFERENCES users(id),
    FOREIGN KEY(receiver_id) REFERENCES users(id),
    CHECK (sender_id != receiver_id)
);

INSERT INTO messages_new (id, sender_id, receiver_id, content, timestamp)
SELECT id, sender_id, receiver_id, content, timestamp
FROM messages;

DROP TABLE messages;
ALTER TABLE messages_new RENAME TO messages;

CREATE INDEX idx_messages_conversation_id
    ON messages(sender_id, receiver_id, id DESC);

CREATE INDEX idx_messages_reverse_conversation_id
    ON messages(receiver_id, sender_id, id DESC);

CREATE TABLE sessions_new (
    session_id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    expires_at DATETIME,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

INSERT INTO sessions_new (session_id, user_id, expires_at)
SELECT session_id, user_id, expires_at
FROM sessions;

DROP TABLE sessions;
ALTER TABLE sessions_new RENAME TO sessions;

CREATE INDEX idx_sessions_expiry
    ON sessions(session_id, expires_at);

CREATE INDEX idx_sessions_user_id
    ON sessions(user_id, expires_at);

CREATE TABLE notifications_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    receiver_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,
    unread_messages INTEGER DEFAULT 0,
    FOREIGN KEY(receiver_id) REFERENCES users(id),
    FOREIGN KEY(sender_id) REFERENCES users(id)
);

INSERT INTO notifications_new (id, receiver_id, sender_id, unread_messages)
SELECT id, receiver_id, sender_id, unread_messages
FROM notifications;

DROP TABLE notifications;
ALTER TABLE notifications_new RENAME TO notifications;

CREATE INDEX idx_notifications_pair
    ON notifications(receiver_id, sender_id);
