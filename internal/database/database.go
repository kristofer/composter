package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	*sql.DB
}

type User struct {
	ID        int
	Username  string
	Password  string
	IsAdmin   bool
	CreatedAt time.Time
}

type Outline struct {
	ID        int
	UserID    int
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func New(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		is_admin BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS outlines (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_outlines_user_id ON outlines(user_id);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	// Create default admin user if no users exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		_, err = db.Exec("INSERT INTO users (username, password, is_admin) VALUES (?, ?, ?)",
			"admin", string(hashedPassword), true)
		if err != nil {
			return err
		}
		fmt.Println("Created default admin user (username: admin, password: admin)")
	}

	return nil
}

// User methods
func (db *DB) CreateUser(username, password string, isAdmin bool) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (username, password, is_admin) VALUES (?, ?, ?)",
		username, string(hashedPassword), isAdmin)
	return err
}

func (db *DB) GetUser(username string) (*User, error) {
	user := &User{}
	err := db.QueryRow("SELECT id, username, password, is_admin, created_at FROM users WHERE username = ?",
		username).Scan(&user.ID, &user.Username, &user.Password, &user.IsAdmin, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) GetUserByID(id int) (*User, error) {
	user := &User{}
	err := db.QueryRow("SELECT id, username, password, is_admin, created_at FROM users WHERE id = ?",
		id).Scan(&user.ID, &user.Username, &user.Password, &user.IsAdmin, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) GetAllUsers() ([]User, error) {
	rows, err := db.Query("SELECT id, username, password, is_admin, created_at FROM users ORDER BY username")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.IsAdmin, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (db *DB) UpdateUser(id int, username string, password string, isAdmin bool) error {
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = db.Exec("UPDATE users SET username = ?, password = ?, is_admin = ? WHERE id = ?",
			username, string(hashedPassword), isAdmin, id)
		return err
	}

	_, err := db.Exec("UPDATE users SET username = ?, is_admin = ? WHERE id = ?",
		username, isAdmin, id)
	return err
}

func (db *DB) DeleteUser(id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func (db *DB) VerifyPassword(username, password string) (*User, error) {
	user, err := db.GetUser(username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Outline methods
func (db *DB) CreateOutline(userID int, title, content string) (int64, error) {
	result, err := db.Exec("INSERT INTO outlines (user_id, title, content) VALUES (?, ?, ?)",
		userID, title, content)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) GetOutline(id, userID int) (*Outline, error) {
	outline := &Outline{}
	err := db.QueryRow("SELECT id, user_id, title, content, created_at, updated_at FROM outlines WHERE id = ? AND user_id = ?",
		id, userID).Scan(&outline.ID, &outline.UserID, &outline.Title, &outline.Content, &outline.CreatedAt, &outline.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return outline, nil
}

func (db *DB) GetUserOutlines(userID int) ([]Outline, error) {
	rows, err := db.Query("SELECT id, user_id, title, content, created_at, updated_at FROM outlines WHERE user_id = ? ORDER BY updated_at DESC",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outlines []Outline
	for rows.Next() {
		var outline Outline
		if err := rows.Scan(&outline.ID, &outline.UserID, &outline.Title, &outline.Content, &outline.CreatedAt, &outline.UpdatedAt); err != nil {
			return nil, err
		}
		outlines = append(outlines, outline)
	}
	return outlines, nil
}

func (db *DB) UpdateOutline(id, userID int, title, content string) error {
	_, err := db.Exec("UPDATE outlines SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?",
		title, content, id, userID)
	return err
}

func (db *DB) DeleteOutline(id, userID int) error {
	_, err := db.Exec("DELETE FROM outlines WHERE id = ? AND user_id = ?", id, userID)
	return err
}
