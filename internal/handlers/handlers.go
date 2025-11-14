package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/kristofer/composter/internal/database"
	"github.com/kristofer/composter/internal/middleware"
)

type Handler struct {
	DB    *database.DB
	Store *middleware.SessionStore
	Tmpl  *template.Template
}

func New(db *database.DB, store *middleware.SessionStore) *Handler {
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	return &Handler{
		DB:    db,
		Store: store,
		Tmpl:  tmpl,
	}
}

func generateSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Login handlers
func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// Check if already logged in
	if cookie, err := r.Cookie("session"); err == nil {
		if _, ok := h.Store.Get(cookie.Value); ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	h.Tmpl.ExecuteTemplate(w, "login.html", nil)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := h.DB.VerifyPassword(username, password)
	if err != nil {
		h.Tmpl.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "Invalid username or password",
		})
		return
	}

	sessionID, err := generateSessionID()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.Store.Set(sessionID, user)

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400, // 24 hours
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil {
		h.Store.Delete(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Outline handlers
func (h *Handler) ListOutlines(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUser(r)

	outlines, err := h.DB.GetUserOutlines(user.ID)
	if err != nil {
		http.Error(w, "Error retrieving outlines", http.StatusInternalServerError)
		return
	}

	h.Tmpl.ExecuteTemplate(w, "outlines.html", map[string]interface{}{
		"User":     user,
		"Outlines": outlines,
	})
}

func (h *Handler) ViewOutline(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUser(r)

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		// New outline
		h.Tmpl.ExecuteTemplate(w, "editor.html", map[string]interface{}{
			"User":    user,
			"Outline": nil,
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid outline ID", http.StatusBadRequest)
		return
	}

	outline, err := h.DB.GetOutline(id, user.ID)
	if err != nil {
		http.Error(w, "Outline not found", http.StatusNotFound)
		return
	}

	h.Tmpl.ExecuteTemplate(w, "editor.html", map[string]interface{}{
		"User":    user,
		"Outline": outline,
	})
}

func (h *Handler) SaveOutline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := middleware.GetUser(r)

	var data struct {
		ID      int    `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if data.ID == 0 {
		// Create new outline
		id, err := h.DB.CreateOutline(user.ID, data.Title, data.Content)
		if err != nil {
			http.Error(w, "Error creating outline", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      id,
		})
	} else {
		// Update existing outline
		err := h.DB.UpdateOutline(data.ID, user.ID, data.Title, data.Content)
		if err != nil {
			http.Error(w, "Error updating outline", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      data.ID,
		})
	}
}

func (h *Handler) DeleteOutline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := middleware.GetUser(r)

	var data struct {
		ID int `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.DB.DeleteOutline(data.ID, user.ID)
	if err != nil {
		http.Error(w, "Error deleting outline", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Admin handlers
func (h *Handler) AdminPage(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUser(r)

	users, err := h.DB.GetAllUsers()
	if err != nil {
		http.Error(w, "Error retrieving users", http.StatusInternalServerError)
		return
	}

	h.Tmpl.ExecuteTemplate(w, "admin.html", map[string]interface{}{
		"User":  user,
		"Users": users,
	})
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"is_admin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.DB.CreateUser(data.Username, data.Password, data.IsAdmin)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"is_admin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.DB.UpdateUser(data.ID, data.Username, data.Password, data.IsAdmin)
	if err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		ID int `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.DB.DeleteUser(data.ID)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Template handlers
func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUser(r)

	systemTemplates, err := h.DB.GetSystemTemplates()
	if err != nil {
		http.Error(w, "Error retrieving system templates", http.StatusInternalServerError)
		return
	}

	userTemplates, err := h.DB.GetUserTemplates(user.ID)
	if err != nil {
		http.Error(w, "Error retrieving user templates", http.StatusInternalServerError)
		return
	}

	h.Tmpl.ExecuteTemplate(w, "templates.html", map[string]interface{}{
		"User":            user,
		"SystemTemplates": systemTemplates,
		"UserTemplates":   userTemplates,
	})
}

func (h *Handler) InstantiateTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := middleware.GetUser(r)

	var data struct {
		TemplateID int `json:"template_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get the template
	template, err := h.DB.GetTemplate(data.TemplateID)
	if err != nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	// Create a new outline from the template
	id, err := h.DB.CreateOutline(user.ID, template.Name, template.Content)
	if err != nil {
		http.Error(w, "Error creating outline from template", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
	})
}

func (h *Handler) CreateTemplateFromOutline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := middleware.GetUser(r)

	var data struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
		Category    string `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, err := h.DB.CreateTemplate(data.Name, data.Description, data.Content, data.Category, false, user.ID)
	if err != nil {
		http.Error(w, "Error creating template", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
	})
}

func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := middleware.GetUser(r)

	var data struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
		Category    string `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.DB.UpdateTemplate(data.ID, data.Name, data.Description, data.Content, data.Category, user.ID)
	if err != nil {
		http.Error(w, "Error updating template", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := middleware.GetUser(r)

	var data struct {
		ID int `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.DB.DeleteTemplate(data.ID, user.ID)
	if err != nil {
		http.Error(w, "Error deleting template", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *Handler) ExportTemplate(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUser(r)

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Template ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	// Get the template
	template, err := h.DB.GetTemplate(id)
	if err != nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	// Check permissions: system templates can be exported by anyone, user templates only by owner
	if !template.IsSystem && template.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Create export structure
	export := map[string]interface{}{
		"name":        template.Name,
		"description": template.Description,
		"content":     template.Content,
		"category":    template.Category,
		"version":     "1.0",
		"exported_at": template.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+template.Name+".json\"")

	json.NewEncoder(w).Encode(export)
}

func (h *Handler) ImportTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := middleware.GetUser(r)

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("template")
	if err != nil {
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Parse JSON
	var importData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
		Category    string `json:"category"`
		Version     string `json:"version"`
	}

	if err := json.NewDecoder(file).Decode(&importData); err != nil {
		http.Error(w, "Invalid template file", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if importData.Name == "" || importData.Description == "" || importData.Content == "" || importData.Category == "" {
		http.Error(w, "Invalid template: missing required fields", http.StatusBadRequest)
		return
	}

	// Create template as user template (not system)
	id, err := h.DB.CreateTemplate(importData.Name, importData.Description, importData.Content, importData.Category, false, user.ID)
	if err != nil {
		http.Error(w, "Error importing template", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
	})
}
