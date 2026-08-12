package account

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestUserRepositoryCreatesAndFindsCredentials(t *testing.T) {
	db, err := sql.Open("sqlite", "file:user-repository-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nickname TEXT UNIQUE NOT NULL,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			age INTEGER NOT NULL,
			gender TEXT NOT NULL
		);`)
	if err != nil {
		t.Fatal(err)
	}

	repository := NewUserRepository(db)
	user := UserRecord{
		Nickname: "alice", FirstName: "Alice", LastName: "Example",
		Email: "alice@example.com", Password: "Password1", Age: 30, Gender: "female",
	}
	if err := repository.Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repository.Exists(user.Email, user.Nickname)
	if err != nil || !found {
		t.Fatalf("Exists returned found=%v, err=%v", found, err)
	}

	hashed, nickname, err := repository.PasswordByIdentifier(user.Email)
	if err != nil {
		t.Fatalf("PasswordByIdentifier failed: %v", err)
	}
	if hashed == user.Password || nickname != user.Nickname {
		t.Fatalf("unexpected credentials result: nickname=%q password hash=%q", nickname, hashed)
	}
}
