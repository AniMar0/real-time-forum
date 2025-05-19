package backend

type User struct {
	Nickname  string `json:"Nickname"`
	Age       int    `json:"Age"`
	Gender    string `json:"Gender"`
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`
	Email     string `json:"Email"`
	Password  string `json:"Password"`
}
