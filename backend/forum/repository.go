package forum

import (
	"database/sql"
	"errors"
)

var ErrPostNotFound = errors.New("post not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreatePost(userNickname, title, content, category string) error {
	_, err := r.db.Exec(`
		INSERT INTO posts (user_id, title, content, category)
		VALUES ((SELECT id FROM users WHERE nickname = ?), ?, ?, ?)`,
		userNickname, title, content, category)
	return err
}

func (r *Repository) ListPosts() ([]Post, error) {
	rows, err := r.db.Query(`
		SELECT posts.id, posts.title, posts.content, posts.category,
		       posts.created_at, users.nickname
		FROM posts
		JOIN users ON posts.user_id = users.id
		ORDER BY posts.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.CreatedAt, &post.Author); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *Repository) CreateComment(postID int, userNickname, content string) error {
	var exists int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM posts WHERE id = ?", postID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return ErrPostNotFound
	}

	_, err := r.db.Exec(`
		INSERT INTO comments (post_id, user_id, content)
		VALUES (?, (SELECT id FROM users WHERE nickname = ?), ?)`,
		postID, userNickname, content)
	return err
}

func (r *Repository) ListComments(postID string) ([]Comment, error) {
	rows, err := r.db.Query(`
		SELECT comments.id, comments.post_id, comments.content,
		       comments.created_at, users.nickname
		FROM comments
		JOIN users ON comments.user_id = users.id
		WHERE comments.post_id = ?
		ORDER BY comments.created_at ASC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.Content, &comment.CreatedAt, &comment.Author); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}
