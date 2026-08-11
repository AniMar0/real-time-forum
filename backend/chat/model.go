package chat

type Message struct {
	ID         int
	SenderID   int64
	ReceiverID int64
	From       string
	To         string
	Content    string
	Timestamp  string
}
