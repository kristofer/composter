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

type Template struct {
	ID          int
	Name        string
	Description string
	Content     string
	Category    string
	IsSystem    bool
	UserID      int // 0 for system templates
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Template categories
const (
	CategoryMVC          = "MVC"
	CategoryAPI          = "API"
	CategoryMicroservice = "Microservice"
	CategoryDataPipeline = "DataPipeline"
	CategoryFeature      = "Feature"
	CategoryBugFix       = "BugFix"
	CategoryGeneral      = "General"
	CategoryBeginner     = "Beginner"
)

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

	CREATE TABLE IF NOT EXISTS templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		content TEXT NOT NULL,
		category TEXT NOT NULL,
		is_system BOOLEAN DEFAULT 0,
		user_id INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_outlines_user_id ON outlines(user_id);
	CREATE INDEX IF NOT EXISTS idx_templates_category ON templates(category);
	CREATE INDEX IF NOT EXISTS idx_templates_user_id ON templates(user_id);
	CREATE INDEX IF NOT EXISTS idx_templates_is_system ON templates(is_system);
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

	// Seed system templates
	if err := db.SeedSystemTemplates(); err != nil {
		return err
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

// Template methods
func (db *DB) CreateTemplate(name, description, content, category string, isSystem bool, userID int) (int64, error) {
	result, err := db.Exec("INSERT INTO templates (name, description, content, category, is_system, user_id) VALUES (?, ?, ?, ?, ?, ?)",
		name, description, content, category, isSystem, userID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) GetTemplate(id int) (*Template, error) {
	template := &Template{}
	err := db.QueryRow("SELECT id, name, description, content, category, is_system, user_id, created_at, updated_at FROM templates WHERE id = ?",
		id).Scan(&template.ID, &template.Name, &template.Description, &template.Content, &template.Category, &template.IsSystem, &template.UserID, &template.CreatedAt, &template.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return template, nil
}

func (db *DB) GetAllTemplates() ([]Template, error) {
	rows, err := db.Query("SELECT id, name, description, content, category, is_system, user_id, created_at, updated_at FROM templates ORDER BY is_system DESC, category, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []Template
	for rows.Next() {
		var template Template
		if err := rows.Scan(&template.ID, &template.Name, &template.Description, &template.Content, &template.Category, &template.IsSystem, &template.UserID, &template.CreatedAt, &template.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}
	return templates, nil
}

func (db *DB) GetSystemTemplates() ([]Template, error) {
	rows, err := db.Query("SELECT id, name, description, content, category, is_system, user_id, created_at, updated_at FROM templates WHERE is_system = 1 ORDER BY category, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []Template
	for rows.Next() {
		var template Template
		if err := rows.Scan(&template.ID, &template.Name, &template.Description, &template.Content, &template.Category, &template.IsSystem, &template.UserID, &template.CreatedAt, &template.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}
	return templates, nil
}

func (db *DB) GetUserTemplates(userID int) ([]Template, error) {
	rows, err := db.Query("SELECT id, name, description, content, category, is_system, user_id, created_at, updated_at FROM templates WHERE user_id = ? ORDER BY category, name",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []Template
	for rows.Next() {
		var template Template
		if err := rows.Scan(&template.ID, &template.Name, &template.Description, &template.Content, &template.Category, &template.IsSystem, &template.UserID, &template.CreatedAt, &template.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}
	return templates, nil
}

func (db *DB) GetTemplatesByCategory(category string) ([]Template, error) {
	rows, err := db.Query("SELECT id, name, description, content, category, is_system, user_id, created_at, updated_at FROM templates WHERE category = ? ORDER BY is_system DESC, name",
		category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []Template
	for rows.Next() {
		var template Template
		if err := rows.Scan(&template.ID, &template.Name, &template.Description, &template.Content, &template.Category, &template.IsSystem, &template.UserID, &template.CreatedAt, &template.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}
	return templates, nil
}

func (db *DB) UpdateTemplate(id int, name, description, content, category string, userID int) error {
	_, err := db.Exec("UPDATE templates SET name = ?, description = ?, content = ?, category = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ? AND is_system = 0",
		name, description, content, category, id, userID)
	return err
}

func (db *DB) DeleteTemplate(id, userID int) error {
	_, err := db.Exec("DELETE FROM templates WHERE id = ? AND user_id = ? AND is_system = 0", id, userID)
	return err
}

// SeedSystemTemplates populates the database with pre-built system templates
func (db *DB) SeedSystemTemplates() error {
	templates := []struct {
		name        string
		description string
		category    string
		content     string
	}{
		{
			name:        "MVC Application",
			description: "Model-View-Controller architecture decomposition",
			category:    CategoryMVC,
			content: `<div>Project: [Application Name]</div>
<div style="margin-left: 30px">Models</div>
<div style="margin-left: 60px">Data structures</div>
<div style="margin-left: 60px">Database schema</div>
<div style="margin-left: 60px">Validation rules</div>
<div style="margin-left: 60px">Business logic</div>
<div style="margin-left: 30px">Views</div>
<div style="margin-left: 60px">UI components</div>
<div style="margin-left: 60px">Templates</div>
<div style="margin-left: 60px">Styling (CSS)</div>
<div style="margin-left: 60px">Client-side JavaScript</div>
<div style="margin-left: 30px">Controllers</div>
<div style="margin-left: 60px">Route handlers</div>
<div style="margin-left: 60px">Request validation</div>
<div style="margin-left: 60px">Response formatting</div>
<div style="margin-left: 60px">Error handling</div>
<div style="margin-left: 30px">Infrastructure</div>
<div style="margin-left: 60px">Database connection</div>
<div style="margin-left: 60px">Authentication/Authorization</div>
<div style="margin-left: 60px">Session management</div>
<div style="margin-left: 60px">Logging</div>
<div style="margin-left: 30px">Testing</div>
<div style="margin-left: 60px">Unit tests (models)</div>
<div style="margin-left: 60px">Integration tests (controllers)</div>
<div style="margin-left: 60px">UI tests (views)</div>`,
		},
		{
			name:        "REST API Design",
			description: "Complete REST API planning and implementation",
			category:    CategoryAPI,
			content: `<div>API: [API Name]</div>
<div style="margin-left: 30px">Resources</div>
<div style="margin-left: 60px">Identify entities</div>
<div style="margin-left: 60px">Define relationships</div>
<div style="margin-left: 60px">Design URL structure</div>
<div style="margin-left: 30px">Endpoints</div>
<div style="margin-left: 60px">GET /resource - List all</div>
<div style="margin-left: 60px">GET /resource/:id - Get single</div>
<div style="margin-left: 60px">POST /resource - Create new</div>
<div style="margin-left: 60px">PUT /resource/:id - Update</div>
<div style="margin-left: 60px">DELETE /resource/:id - Delete</div>
<div style="margin-left: 30px">Authentication</div>
<div style="margin-left: 60px">Auth strategy (JWT, OAuth, API keys)</div>
<div style="margin-left: 60px">Login/Register endpoints</div>
<div style="margin-left: 60px">Token refresh mechanism</div>
<div style="margin-left: 60px">Permission model</div>
<div style="margin-left: 30px">Request/Response</div>
<div style="margin-left: 60px">Input validation</div>
<div style="margin-left: 60px">Response format (JSON schema)</div>
<div style="margin-left: 60px">Pagination</div>
<div style="margin-left: 60px">Filtering and sorting</div>
<div style="margin-left: 30px">Error Handling</div>
<div style="margin-left: 60px">HTTP status codes</div>
<div style="margin-left: 60px">Error response format</div>
<div style="margin-left: 60px">Validation errors</div>
<div style="margin-left: 60px">Rate limiting</div>
<div style="margin-left: 30px">Documentation</div>
<div style="margin-left: 60px">OpenAPI/Swagger spec</div>
<div style="margin-left: 60px">Endpoint descriptions</div>
<div style="margin-left: 60px">Example requests/responses</div>
<div style="margin-left: 60px">Authentication guide</div>
<div style="margin-left: 30px">Testing</div>
<div style="margin-left: 60px">Unit tests (business logic)</div>
<div style="margin-left: 60px">Integration tests (endpoints)</div>
<div style="margin-left: 60px">Load testing</div>`,
		},
		{
			name:        "Microservice Architecture",
			description: "Microservice design and decomposition",
			category:    CategoryMicroservice,
			content: `<div>System: [System Name]</div>
<div style="margin-left: 30px">Service Boundaries</div>
<div style="margin-left: 60px">Identify bounded contexts</div>
<div style="margin-left: 60px">Define service responsibilities</div>
<div style="margin-left: 60px">Data ownership per service</div>
<div style="margin-left: 30px">Services</div>
<div style="margin-left: 60px">[Service 1 Name]</div>
<div style="margin-left: 90px">API endpoints</div>
<div style="margin-left: 90px">Data model</div>
<div style="margin-left: 90px">Dependencies</div>
<div style="margin-left: 60px">[Service 2 Name]</div>
<div style="margin-left: 90px">API endpoints</div>
<div style="margin-left: 90px">Data model</div>
<div style="margin-left: 90px">Dependencies</div>
<div style="margin-left: 30px">Communication</div>
<div style="margin-left: 60px">Synchronous (REST/gRPC)</div>
<div style="margin-left: 60px">Asynchronous (message queue)</div>
<div style="margin-left: 60px">Service discovery</div>
<div style="margin-left: 60px">API gateway</div>
<div style="margin-left: 30px">Data Management</div>
<div style="margin-left: 60px">Database per service</div>
<div style="margin-left: 60px">Data consistency strategy</div>
<div style="margin-left: 60px">Event sourcing (if needed)</div>
<div style="margin-left: 60px">CQRS pattern (if needed)</div>
<div style="margin-left: 30px">Deployment</div>
<div style="margin-left: 60px">Containerization (Docker)</div>
<div style="margin-left: 60px">Orchestration (Kubernetes)</div>
<div style="margin-left: 60px">CI/CD pipeline</div>
<div style="margin-left: 60px">Service configuration</div>
<div style="margin-left: 30px">Observability</div>
<div style="margin-left: 60px">Centralized logging</div>
<div style="margin-left: 60px">Distributed tracing</div>
<div style="margin-left: 60px">Metrics and monitoring</div>
<div style="margin-left: 60px">Health checks</div>
<div style="margin-left: 30px">Resilience</div>
<div style="margin-left: 60px">Circuit breakers</div>
<div style="margin-left: 60px">Retry policies</div>
<div style="margin-left: 60px">Timeout handling</div>
<div style="margin-left: 60px">Fallback strategies</div>`,
		},
		{
			name:        "Data Pipeline",
			description: "ETL/ELT data pipeline design",
			category:    CategoryDataPipeline,
			content: `<div>Pipeline: [Pipeline Name]</div>
<div style="margin-left: 30px">Data Sources</div>
<div style="margin-left: 60px">Source 1: [Type/Location]</div>
<div style="margin-left: 90px">Connection details</div>
<div style="margin-left: 90px">Data format</div>
<div style="margin-left: 90px">Update frequency</div>
<div style="margin-left: 60px">Source 2: [Type/Location]</div>
<div style="margin-left: 30px">Ingestion</div>
<div style="margin-left: 60px">Ingestion method (batch/stream)</div>
<div style="margin-left: 60px">Schedule/triggers</div>
<div style="margin-left: 60px">Error handling</div>
<div style="margin-left: 60px">Data validation on ingestion</div>
<div style="margin-left: 30px">Transformation</div>
<div style="margin-left: 60px">Data cleaning</div>
<div style="margin-left: 90px">Remove duplicates</div>
<div style="margin-left: 90px">Handle missing values</div>
<div style="margin-left: 90px">Fix data types</div>
<div style="margin-left: 60px">Data enrichment</div>
<div style="margin-left: 90px">Join with reference data</div>
<div style="margin-left: 90px">Calculate derived fields</div>
<div style="margin-left: 60px">Data aggregation</div>
<div style="margin-left: 60px">Business rules</div>
<div style="margin-left: 30px">Validation</div>
<div style="margin-left: 60px">Schema validation</div>
<div style="margin-left: 60px">Data quality checks</div>
<div style="margin-left: 60px">Business rule validation</div>
<div style="margin-left: 60px">Anomaly detection</div>
<div style="margin-left: 30px">Storage</div>
<div style="margin-left: 60px">Target destination</div>
<div style="margin-left: 60px">Data partitioning strategy</div>
<div style="margin-left: 60px">Retention policy</div>
<div style="margin-left: 60px">Backup strategy</div>
<div style="margin-left: 30px">Monitoring</div>
<div style="margin-left: 60px">Pipeline execution metrics</div>
<div style="margin-left: 60px">Data quality metrics</div>
<div style="margin-left: 60px">Alerting on failures</div>
<div style="margin-left: 60px">Performance monitoring</div>
<div style="margin-left: 30px">Testing</div>
<div style="margin-left: 60px">Unit tests (transformations)</div>
<div style="margin-left: 60px">Integration tests (end-to-end)</div>
<div style="margin-left: 60px">Data validation tests</div>`,
		},
		{
			name:        "Feature Development",
			description: "Complete feature implementation workflow",
			category:    CategoryFeature,
			content: `<div>Feature: [Feature Name]</div>
<div style="margin-left: 30px">Requirements</div>
<div style="margin-left: 60px">User stories</div>
<div style="margin-left: 60px">Acceptance criteria</div>
<div style="margin-left: 60px">Edge cases</div>
<div style="margin-left: 60px">Non-functional requirements</div>
<div style="margin-left: 30px">Design</div>
<div style="margin-left: 60px">Architecture changes</div>
<div style="margin-left: 60px">Data model changes</div>
<div style="margin-left: 60px">API design</div>
<div style="margin-left: 60px">UI/UX mockups</div>
<div style="margin-left: 30px">Implementation</div>
<div style="margin-left: 60px">Backend</div>
<div style="margin-left: 90px">Database migrations</div>
<div style="margin-left: 90px">Business logic</div>
<div style="margin-left: 90px">API endpoints</div>
<div style="margin-left: 90px">Error handling</div>
<div style="margin-left: 60px">Frontend</div>
<div style="margin-left: 90px">UI components</div>
<div style="margin-left: 90px">State management</div>
<div style="margin-left: 90px">API integration</div>
<div style="margin-left: 90px">Form validation</div>
<div style="margin-left: 30px">Testing</div>
<div style="margin-left: 60px">Unit tests</div>
<div style="margin-left: 60px">Integration tests</div>
<div style="margin-left: 60px">E2E tests</div>
<div style="margin-left: 60px">Manual testing checklist</div>
<div style="margin-left: 30px">Documentation</div>
<div style="margin-left: 60px">Code comments</div>
<div style="margin-left: 60px">API documentation</div>
<div style="margin-left: 60px">User documentation</div>
<div style="margin-left: 60px">Release notes</div>
<div style="margin-left: 30px">Deployment</div>
<div style="margin-left: 60px">Feature flags (if applicable)</div>
<div style="margin-left: 60px">Staging deployment</div>
<div style="margin-left: 60px">Production deployment</div>
<div style="margin-left: 60px">Monitoring and rollback plan</div>`,
		},
		{
			name:        "Bug Fix Process",
			description: "Systematic bug investigation and resolution",
			category:    CategoryBugFix,
			content: `<div>Bug: [Bug Description]</div>
<div style="margin-left: 30px">Reproduce</div>
<div style="margin-left: 60px">Steps to reproduce</div>
<div style="margin-left: 60px">Expected behavior</div>
<div style="margin-left: 60px">Actual behavior</div>
<div style="margin-left: 60px">Environment details</div>
<div style="margin-left: 30px">Diagnose</div>
<div style="margin-left: 60px">Review error logs</div>
<div style="margin-left: 60px">Check recent changes</div>
<div style="margin-left: 60px">Isolate the problem</div>
<div style="margin-left: 90px">Frontend vs backend</div>
<div style="margin-left: 90px">Specific component/function</div>
<div style="margin-left: 90px">Data issue vs code issue</div>
<div style="margin-left: 60px">Identify root cause</div>
<div style="margin-left: 30px">Fix</div>
<div style="margin-left: 60px">Develop solution</div>
<div style="margin-left: 60px">Consider side effects</div>
<div style="margin-left: 60px">Update related code</div>
<div style="margin-left: 60px">Add defensive checks</div>
<div style="margin-left: 30px">Test</div>
<div style="margin-left: 60px">Verify fix resolves issue</div>
<div style="margin-left: 60px">Test edge cases</div>
<div style="margin-left: 60px">Regression testing</div>
<div style="margin-left: 60px">Add test to prevent recurrence</div>
<div style="margin-left: 30px">Deploy</div>
<div style="margin-left: 60px">Code review</div>
<div style="margin-left: 60px">Staging verification</div>
<div style="margin-left: 60px">Production deployment</div>
<div style="margin-left: 60px">Monitor for issues</div>
<div style="margin-left: 30px">Document</div>
<div style="margin-left: 60px">Update issue tracker</div>
<div style="margin-left: 60px">Document root cause</div>
<div style="margin-left: 60px">Update documentation if needed</div>`,
		},
		{
			name:        "Word Guess Game",
			description: "Terminal-based word guessing game project structure",
			category:    CategoryBeginner,
			content: `<div>Project: Word Guess Game</div>
<div style="margin-left: 30px">Setup</div>
<div style="margin-left: 60px">Initialize project</div>
<div style="margin-left: 60px">Choose programming language</div>
<div style="margin-left: 60px">Set up development environment</div>
<div style="margin-left: 30px">Core Features</div>
<div style="margin-left: 60px">Word list management</div>
<div style="margin-left: 90px">Load words from file or array</div>
<div style="margin-left: 90px">Select random word</div>
<div style="margin-left: 60px">Game state</div>
<div style="margin-left: 90px">Track guessed letters</div>
<div style="margin-left: 90px">Track remaining attempts</div>
<div style="margin-left: 90px">Display masked word (e.g., _ _ _ _)</div>
<div style="margin-left: 60px">User input</div>
<div style="margin-left: 90px">Read letter from terminal</div>
<div style="margin-left: 90px">Validate input (single letter)</div>
<div style="margin-left: 90px">Check if already guessed</div>
<div style="margin-left: 60px">Game logic</div>
<div style="margin-left: 90px">Check if letter is in word</div>
<div style="margin-left: 90px">Update display</div>
<div style="margin-left: 90px">Decrease attempts if wrong</div>
<div style="margin-left: 90px">Check win/lose conditions</div>
<div style="margin-left: 30px">Display</div>
<div style="margin-left: 60px">Show current word state</div>
<div style="margin-left: 60px">Show guessed letters</div>
<div style="margin-left: 60px">Show remaining attempts</div>
<div style="margin-left: 60px">Draw hangman figure (optional)</div>
<div style="margin-left: 30px">Game Loop</div>
<div style="margin-left: 60px">Initialize game</div>
<div style="margin-left: 60px">Loop until win or lose</div>
<div style="margin-left: 60px">Display end message</div>
<div style="margin-left: 60px">Ask to play again</div>
<div style="margin-left: 30px">Testing</div>
<div style="margin-left: 60px">Test word selection</div>
<div style="margin-left: 60px">Test input validation</div>
<div style="margin-left: 60px">Test game logic</div>
<div style="margin-left: 60px">Play through complete game</div>`,
		},
		{
			name:        "CLI Text Processor",
			description: "Command-line tool for processing text files",
			category:    CategoryBeginner,
			content: `<div>Project: CLI Text Processor</div>
<div style="margin-left: 30px">Setup</div>
<div style="margin-left: 60px">Initialize project</div>
<div style="margin-left: 60px">Set up argument parsing library</div>
<div style="margin-left: 60px">Create project structure</div>
<div style="margin-left: 30px">Command-Line Interface</div>
<div style="margin-left: 60px">Define flags and options</div>
<div style="margin-left: 90px">--input/-i: input file path</div>
<div style="margin-left: 90px">--output/-o: output file path</div>
<div style="margin-left: 90px">--operation: type of processing</div>
<div style="margin-left: 60px">Parse arguments</div>
<div style="margin-left: 60px">Validate input parameters</div>
<div style="margin-left: 60px">Display help message</div>
<div style="margin-left: 30px">File Operations</div>
<div style="margin-left: 60px">Read input file</div>
<div style="margin-left: 90px">Handle file not found</div>
<div style="margin-left: 90px">Handle read errors</div>
<div style="margin-left: 60px">Write output file</div>
<div style="margin-left: 90px">Handle write errors</div>
<div style="margin-left: 90px">Create parent directories if needed</div>
<div style="margin-left: 30px">Text Processing Functions</div>
<div style="margin-left: 60px">Word count</div>
<div style="margin-left: 90px">Count total words</div>
<div style="margin-left: 90px">Count unique words</div>
<div style="margin-left: 60px">Find and replace</div>
<div style="margin-left: 90px">Simple text replacement</div>
<div style="margin-left: 90px">Regex-based replacement</div>
<div style="margin-left: 60px">Case conversion</div>
<div style="margin-left: 90px">Uppercase</div>
<div style="margin-left: 90px">Lowercase</div>
<div style="margin-left: 90px">Title case</div>
<div style="margin-left: 60px">Remove duplicates</div>
<div style="margin-left: 90px">Remove duplicate lines</div>
<div style="margin-left: 90px">Preserve order</div>
<div style="margin-left: 60px">Sort lines</div>
<div style="margin-left: 90px">Alphabetically</div>
<div style="margin-left: 90px">Numerically</div>
<div style="margin-left: 90px">Reverse order</div>
<div style="margin-left: 30px">Output Formatting</div>
<div style="margin-left: 60px">Display results to stdout</div>
<div style="margin-left: 60px">Write to file</div>
<div style="margin-left: 60px">Show statistics</div>
<div style="margin-left: 30px">Error Handling</div>
<div style="margin-left: 60px">Invalid file paths</div>
<div style="margin-left: 60px">Permission errors</div>
<div style="margin-left: 60px">Invalid operations</div>
<div style="margin-left: 60px">Provide helpful error messages</div>
<div style="margin-left: 30px">Testing</div>
<div style="margin-left: 60px">Test each processing function</div>
<div style="margin-left: 60px">Test CLI argument parsing</div>
<div style="margin-left: 60px">Test file I/O operations</div>
<div style="margin-left: 60px">Test error handling</div>`,
		},
		{
			name:        "Command-Line Notes App",
			description: "Simple note-taking application for the terminal",
			category:    CategoryBeginner,
			content: `<div>Project: Command-Line Notes</div>
<div style="margin-left: 30px">Setup</div>
<div style="margin-left: 60px">Initialize project</div>
<div style="margin-left: 60px">Choose data storage format (JSON, SQLite, etc.)</div>
<div style="margin-left: 60px">Set up project structure</div>
<div style="margin-left: 30px">Data Model</div>
<div style="margin-left: 60px">Note structure</div>
<div style="margin-left: 90px">ID (unique identifier)</div>
<div style="margin-left: 90px">Title</div>
<div style="margin-left: 90px">Content/body</div>
<div style="margin-left: 90px">Created timestamp</div>
<div style="margin-left: 90px">Modified timestamp</div>
<div style="margin-left: 90px">Tags (optional)</div>
<div style="margin-left: 30px">Commands</div>
<div style="margin-left: 60px">add - Create new note</div>
<div style="margin-left: 90px">Prompt for title</div>
<div style="margin-left: 90px">Prompt for content (multiline)</div>
<div style="margin-left: 90px">Save note</div>
<div style="margin-left: 60px">list - Display all notes</div>
<div style="margin-left: 90px">Show ID, title, date</div>
<div style="margin-left: 90px">Format as table</div>
<div style="margin-left: 60px">view - Show note details</div>
<div style="margin-left: 90px">Accept note ID</div>
<div style="margin-left: 90px">Display full content</div>
<div style="margin-left: 60px">edit - Modify existing note</div>
<div style="margin-left: 90px">Find note by ID</div>
<div style="margin-left: 90px">Edit title and/or content</div>
<div style="margin-left: 90px">Update modified timestamp</div>
<div style="margin-left: 60px">delete - Remove note</div>
<div style="margin-left: 90px">Accept note ID</div>
<div style="margin-left: 90px">Confirm deletion</div>
<div style="margin-left: 60px">search - Find notes</div>
<div style="margin-left: 90px">Search by title</div>
<div style="margin-left: 90px">Search by content</div>
<div style="margin-left: 90px">Search by tag (if implemented)</div>
<div style="margin-left: 30px">Storage</div>
<div style="margin-left: 60px">Load notes from storage</div>
<div style="margin-left: 60px">Save notes to storage</div>
<div style="margin-left: 60px">Handle storage errors</div>
<div style="margin-left: 60px">Data persistence</div>
<div style="margin-left: 30px">User Interface</div>
<div style="margin-left: 60px">Command menu</div>
<div style="margin-left: 60px">Input prompts</div>
<div style="margin-left: 60px">Display formatting</div>
<div style="margin-left: 60px">Error messages</div>
<div style="margin-left: 30px">Features (Optional)</div>
<div style="margin-left: 60px">Tag support</div>
<div style="margin-left: 60px">Export notes</div>
<div style="margin-left: 60px">Import notes</div>
<div style="margin-left: 60px">Note categories</div>
<div style="margin-left: 30px">Testing</div>
<div style="margin-left: 60px">Test CRUD operations</div>
<div style="margin-left: 60px">Test search functionality</div>
<div style="margin-left: 60px">Test data persistence</div>
<div style="margin-left: 60px">Test edge cases</div>`,
		},
		{
			name:        "Text-Based Dungeon Game",
			description: "Interactive dungeon exploration game for the terminal",
			category:    CategoryBeginner,
			content: `<div>Project: Text Dungeon Game</div>
<div style="margin-left: 30px">Setup</div>
<div style="margin-left: 60px">Initialize project</div>
<div style="margin-left: 60px">Choose programming language</div>
<div style="margin-left: 60px">Set up game structure</div>
<div style="margin-left: 30px">Game Data Models</div>
<div style="margin-left: 60px">Player</div>
<div style="margin-left: 90px">Health points</div>
<div style="margin-left: 90px">Inventory</div>
<div style="margin-left: 90px">Current location</div>
<div style="margin-left: 90px">Stats (strength, defense, etc.)</div>
<div style="margin-left: 60px">Room</div>
<div style="margin-left: 90px">Description</div>
<div style="margin-left: 90px">Connected rooms (north, south, east, west)</div>
<div style="margin-left: 90px">Items in room</div>
<div style="margin-left: 90px">Monsters in room</div>
<div style="margin-left: 60px">Item</div>
<div style="margin-left: 90px">Name</div>
<div style="margin-left: 90px">Description</div>
<div style="margin-left: 90px">Type (weapon, potion, key, etc.)</div>
<div style="margin-left: 90px">Properties (damage, healing, etc.)</div>
<div style="margin-left: 60px">Monster</div>
<div style="margin-left: 90px">Name</div>
<div style="margin-left: 90px">Health</div>
<div style="margin-left: 90px">Attack damage</div>
<div style="margin-left: 90px">Loot drops</div>
<div style="margin-left: 30px">Game World</div>
<div style="margin-left: 60px">Create dungeon layout</div>
<div style="margin-left: 60px">Define rooms and connections</div>
<div style="margin-left: 60px">Place items</div>
<div style="margin-left: 60px">Place monsters</div>
<div style="margin-left: 60px">Set win condition</div>
<div style="margin-left: 30px">Commands</div>
<div style="margin-left: 60px">Movement (go north/south/east/west)</div>
<div style="margin-left: 60px">Look (examine room)</div>
<div style="margin-left: 60px">Inventory (check items)</div>
<div style="margin-left: 60px">Take (pick up item)</div>
<div style="margin-left: 60px">Use (use item)</div>
<div style="margin-left: 60px">Attack (fight monster)</div>
<div style="margin-left: 60px">Help (show commands)</div>
<div style="margin-left: 60px">Quit (exit game)</div>
<div style="margin-left: 30px">Game Mechanics</div>
<div style="margin-left: 60px">Movement between rooms</div>
<div style="margin-left: 60px">Item interaction</div>
<div style="margin-left: 90px">Pick up items</div>
<div style="margin-left: 90px">Use items (potions, keys)</div>
<div style="margin-left: 90px">Equip weapons</div>
<div style="margin-left: 60px">Combat system</div>
<div style="margin-left: 90px">Turn-based fighting</div>
<div style="margin-left: 90px">Damage calculation</div>
<div style="margin-left: 90px">Monster AI (basic)</div>
<div style="margin-left: 90px">Death handling</div>
<div style="margin-left: 60px">Puzzle elements (locked doors, keys)</div>
<div style="margin-left: 30px">User Interface</div>
<div style="margin-left: 60px">Display room description</div>
<div style="margin-left: 60px">Show available exits</div>
<div style="margin-left: 60px">Show player status (health, inventory)</div>
<div style="margin-left: 60px">Parse user commands</div>
<div style="margin-left: 60px">Provide feedback messages</div>
<div style="margin-left: 30px">Game Loop</div>
<div style="margin-left: 60px">Initialize game state</div>
<div style="margin-left: 60px">Display current situation</div>
<div style="margin-left: 60px">Get player input</div>
<div style="margin-left: 60px">Process command</div>
<div style="margin-left: 60px">Update game state</div>
<div style="margin-left: 60px">Check win/lose conditions</div>
<div style="margin-left: 30px">Testing</div>
<div style="margin-left: 60px">Test movement system</div>
<div style="margin-left: 60px">Test combat mechanics</div>
<div style="margin-left: 60px">Test item interactions</div>
<div style="margin-left: 60px">Playtest complete game</div>`,
		},
		{
			name:        "LLM Chat Terminal",
			description: "Terminal-based chat interface with LLM API",
			category:    CategoryBeginner,
			content: `<div>Project: LLM Chat Terminal</div>
<div style="margin-left: 30px">Setup</div>
<div style="margin-left: 60px">Initialize project</div>
<div style="margin-left: 60px">Choose LLM API (OpenAI, Anthropic, etc.)</div>
<div style="margin-left: 60px">Install HTTP client library</div>
<div style="margin-left: 60px">Set up environment variables</div>
<div style="margin-left: 30px">Configuration</div>
<div style="margin-left: 60px">API key management</div>
<div style="margin-left: 90px">Load from environment variable</div>
<div style="margin-left: 90px">Load from config file</div>
<div style="margin-left: 90px">Secure storage</div>
<div style="margin-left: 60px">API settings</div>
<div style="margin-left: 90px">Model selection</div>
<div style="margin-left: 90px">Temperature setting</div>
<div style="margin-left: 90px">Max tokens</div>
<div style="margin-left: 90px">Other parameters</div>
<div style="margin-left: 30px">API Integration</div>
<div style="margin-left: 60px">Build API request</div>
<div style="margin-left: 90px">Format message payload</div>
<div style="margin-left: 90px">Set headers (authorization, content-type)</div>
<div style="margin-left: 90px">Handle conversation history</div>
<div style="margin-left: 60px">Send HTTP request</div>
<div style="margin-left: 90px">POST to API endpoint</div>
<div style="margin-left: 90px">Handle timeout</div>
<div style="margin-left: 60px">Parse API response</div>
<div style="margin-left: 90px">Extract message content</div>
<div style="margin-left: 90px">Handle errors</div>
<div style="margin-left: 90px">Parse JSON response</div>
<div style="margin-left: 30px">Conversation Management</div>
<div style="margin-left: 60px">Message history</div>
<div style="margin-left: 90px">Store user messages</div>
<div style="margin-left: 90px">Store assistant responses</div>
<div style="margin-left: 90px">Maintain context window</div>
<div style="margin-left: 60px">Session handling</div>
<div style="margin-left: 90px">Start new conversation</div>
<div style="margin-left: 90px">Continue existing conversation</div>
<div style="margin-left: 90px">Save conversation to file</div>
<div style="margin-left: 90px">Load conversation from file</div>
<div style="margin-left: 30px">User Interface</div>
<div style="margin-left: 60px">Display welcome message</div>
<div style="margin-left: 60px">Show prompt for user input</div>
<div style="margin-left: 60px">Display messages</div>
<div style="margin-left: 90px">Format user messages</div>
<div style="margin-left: 90px">Format assistant messages</div>
<div style="margin-left: 90px">Add visual distinction</div>
<div style="margin-left: 60px">Show loading indicator</div>
<div style="margin-left: 60px">Command handling</div>
<div style="margin-left: 90px">/help - show commands</div>
<div style="margin-left: 90px">/new - start new conversation</div>
<div style="margin-left: 90px">/save - save conversation</div>
<div style="margin-left: 90px">/load - load conversation</div>
<div style="margin-left: 90px">/quit - exit application</div>
<div style="margin-left: 30px">Error Handling</div>
<div style="margin-left: 60px">API errors</div>
<div style="margin-left: 90px">Invalid API key</div>
<div style="margin-left: 90px">Rate limiting</div>
<div style="margin-left: 90px">Network errors</div>
<div style="margin-left: 60px">Input validation</div>
<div style="margin-left: 60px">Handle empty messages</div>
<div style="margin-left: 60px">Provide user-friendly error messages</div>
<div style="margin-left: 30px">Features (Optional)</div>
<div style="margin-left: 60px">Streaming responses</div>
<div style="margin-left: 60px">Multiple conversations</div>
<div style="margin-left: 60px">System prompts/personas</div>
<div style="margin-left: 60px">Token usage tracking</div>
<div style="margin-left: 60px">Cost estimation</div>
<div style="margin-left: 30px">Testing</div>
<div style="margin-left: 60px">Test API integration (with mock)</div>
<div style="margin-left: 60px">Test conversation history</div>
<div style="margin-left: 60px">Test command parsing</div>
<div style="margin-left: 60px">Manual testing with real API</div>`,
		},
	}

	for _, tmpl := range templates {
		// Check if template already exists
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM templates WHERE name = ? AND is_system = 1", tmpl.name).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check template %s: %w", tmpl.name, err)
		}
		
		if count > 0 {
			// Template already exists, skip it
			continue
		}
		
		_, err = db.CreateTemplate(tmpl.name, tmpl.description, tmpl.content, tmpl.category, true, 0)
		if err != nil {
			return fmt.Errorf("failed to seed template %s: %w", tmpl.name, err)
		}
	}

	fmt.Println("System templates seeded successfully")
	return nil
}
