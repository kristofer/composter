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
	// Check if system templates already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM templates WHERE is_system = 1").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Templates already seeded
	}

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
	}

	for _, tmpl := range templates {
		_, err := db.CreateTemplate(tmpl.name, tmpl.description, tmpl.content, tmpl.category, true, 0)
		if err != nil {
			return fmt.Errorf("failed to seed template %s: %w", tmpl.name, err)
		}
	}

	fmt.Println("System templates seeded successfully")
	return nil
}
