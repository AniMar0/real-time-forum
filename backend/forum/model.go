package forum

type Post struct {
	ID        int
	Title     string
	Content   string
	Category  string
	CreatedAt string
	Author    string
}

type Comment struct {
	ID        int
	PostID    int
	Content   string
	CreatedAt string
	Author    string
}
