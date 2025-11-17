package database

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	// Create a temporary database file
	dbPath := "/tmp/test_composter.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test that we can ping the database
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}
}

func TestInit(t *testing.T) {
	dbPath := "/tmp/test_composter_init.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initialize the database
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Verify that tables were created by checking for default admin user
	user, err := db.GetUser("admin")
	if err != nil {
		t.Fatalf("Failed to get admin user: %v", err)
	}

	if user.Username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", user.Username)
	}

	if !user.IsAdmin {
		t.Error("Expected admin user to have IsAdmin = true")
	}
}

func TestCreateAndGetUser(t *testing.T) {
	dbPath := "/tmp/test_composter_user.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a new user
	username := "testuser"
	password := "testpass"
	err = db.CreateUser(username, password, false)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Retrieve the user
	user, err := db.GetUser(username)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user.Username != username {
		t.Errorf("Expected username '%s', got '%s'", username, user.Username)
	}

	if user.IsAdmin {
		t.Error("Expected non-admin user to have IsAdmin = false")
	}
}

func TestVerifyPassword(t *testing.T) {
	dbPath := "/tmp/test_composter_password.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Test with default admin user
	user, err := db.VerifyPassword("admin", "admin")
	if err != nil {
		t.Fatalf("Failed to verify admin password: %v", err)
	}

	if user.Username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", user.Username)
	}

	// Test with wrong password
	_, err = db.VerifyPassword("admin", "wrongpassword")
	if err == nil {
		t.Error("Expected error when verifying wrong password")
	}
}

func TestCreateAndGetOutline(t *testing.T) {
	dbPath := "/tmp/test_composter_outline.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Get admin user
	user, err := db.GetUser("admin")
	if err != nil {
		t.Fatalf("Failed to get admin user: %v", err)
	}

	// Create an outline
	title := "Test Outline"
	content := "<div>Item 1</div><div style=\"margin-left: 30px;\">Sub Item</div>"
	id, err := db.CreateOutline(user.ID, title, content)
	if err != nil {
		t.Fatalf("Failed to create outline: %v", err)
	}

	if id <= 0 {
		t.Error("Expected positive outline ID")
	}

	// Retrieve the outline
	outline, err := db.GetOutline(int(id), user.ID)
	if err != nil {
		t.Fatalf("Failed to get outline: %v", err)
	}

	if outline.Title != title {
		t.Errorf("Expected title '%s', got '%s'", title, outline.Title)
	}

	if outline.Content != content {
		t.Errorf("Expected content '%s', got '%s'", content, outline.Content)
	}

	if outline.UserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, outline.UserID)
	}
}

func TestUpdateOutline(t *testing.T) {
	dbPath := "/tmp/test_composter_update.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	user, err := db.GetUser("admin")
	if err != nil {
		t.Fatalf("Failed to get admin user: %v", err)
	}

	// Create an outline
	title := "Original Title"
	content := "<div>Original Content</div>"
	id, err := db.CreateOutline(user.ID, title, content)
	if err != nil {
		t.Fatalf("Failed to create outline: %v", err)
	}

	// Update the outline
	newTitle := "Updated Title"
	newContent := "<div>Updated Content</div>"
	err = db.UpdateOutline(int(id), user.ID, newTitle, newContent)
	if err != nil {
		t.Fatalf("Failed to update outline: %v", err)
	}

	// Verify the update
	outline, err := db.GetOutline(int(id), user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated outline: %v", err)
	}

	if outline.Title != newTitle {
		t.Errorf("Expected title '%s', got '%s'", newTitle, outline.Title)
	}

	if outline.Content != newContent {
		t.Errorf("Expected content '%s', got '%s'", newContent, outline.Content)
	}
}

func TestDeleteOutline(t *testing.T) {
	dbPath := "/tmp/test_composter_delete.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	user, err := db.GetUser("admin")
	if err != nil {
		t.Fatalf("Failed to get admin user: %v", err)
	}

	// Create an outline
	id, err := db.CreateOutline(user.ID, "Test", "Content")
	if err != nil {
		t.Fatalf("Failed to create outline: %v", err)
	}

	// Delete the outline
	err = db.DeleteOutline(int(id), user.ID)
	if err != nil {
		t.Fatalf("Failed to delete outline: %v", err)
	}

	// Verify deletion
	_, err = db.GetOutline(int(id), user.ID)
	if err == nil {
		t.Error("Expected error when getting deleted outline")
	}
}

func TestGetUserOutlines(t *testing.T) {
	dbPath := "/tmp/test_composter_list.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	user, err := db.GetUser("admin")
	if err != nil {
		t.Fatalf("Failed to get admin user: %v", err)
	}

	// Create multiple outlines
	_, err = db.CreateOutline(user.ID, "Outline 1", "Content 1")
	if err != nil {
		t.Fatalf("Failed to create outline 1: %v", err)
	}

	_, err = db.CreateOutline(user.ID, "Outline 2", "Content 2")
	if err != nil {
		t.Fatalf("Failed to create outline 2: %v", err)
	}

	// Get all outlines
	outlines, err := db.GetUserOutlines(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user outlines: %v", err)
	}

	if len(outlines) != 2 {
		t.Errorf("Expected 2 outlines, got %d", len(outlines))
	}

	// Verify both outlines exist (don't rely on specific ordering when created quickly)
	titles := make(map[string]bool)
	for _, outline := range outlines {
		titles[outline.Title] = true
	}

	if !titles["Outline 1"] {
		t.Error("Expected to find 'Outline 1'")
	}
	if !titles["Outline 2"] {
		t.Error("Expected to find 'Outline 2'")
	}
}

func TestBeginnerTemplatesSeeded(t *testing.T) {
	dbPath := "/tmp/test_composter_beginner_templates.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Get all system templates
	templates, err := db.GetSystemTemplates()
	if err != nil {
		t.Fatalf("Failed to get system templates: %v", err)
	}

	// Should have original 6 templates + 5 new beginner templates = 11 total
	if len(templates) != 11 {
		t.Errorf("Expected 11 system templates, got %d", len(templates))
	}

	// Check for beginner templates
	beginnerTemplates := []string{
		"Word Guess Game",
		"CLI Text Processor",
		"Command-Line Notes App",
		"Text-Based Dungeon Game",
		"LLM Chat Terminal",
	}

	templateNames := make(map[string]bool)
	for _, tmpl := range templates {
		templateNames[tmpl.Name] = true
	}

	for _, name := range beginnerTemplates {
		if !templateNames[name] {
			t.Errorf("Expected to find beginner template '%s'", name)
		}
	}

	// Verify beginner category exists in templates
	beginnerCategoryFound := false
	for _, tmpl := range templates {
		if tmpl.Category == CategoryBeginner {
			beginnerCategoryFound = true
			break
		}
	}

	if !beginnerCategoryFound {
		t.Error("Expected to find at least one template with CategoryBeginner")
	}
}
