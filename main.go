package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kristofer/composter/internal/database"
	"github.com/kristofer/composter/internal/handlers"
	"github.com/kristofer/composter/internal/middleware"
)

func main() {
	// Initialize database
	db, err := database.New("composter.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		log.Fatal("Error initializing database:", err)
	}

	// Create session store
	store := middleware.NewSessionStore()

	// Create handlers
	h := handlers.New(db, store)

	// Setup routes
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.LoginPage(w, r)
		} else {
			h.Login(w, r)
		}
	})

	// Protected routes
	authMux := http.NewServeMux()
	authMux.HandleFunc("/", h.ListOutlines)
	authMux.HandleFunc("/logout", h.Logout)
	authMux.HandleFunc("/editor", h.ViewOutline)
	authMux.HandleFunc("/templates", h.ListTemplates)
	authMux.HandleFunc("/api/outline/save", h.SaveOutline)
	authMux.HandleFunc("/api/outline/delete", h.DeleteOutline)
	authMux.HandleFunc("/api/template/instantiate", h.InstantiateTemplate)
	authMux.HandleFunc("/api/template/create", h.CreateTemplateFromOutline)
	authMux.HandleFunc("/api/template/update", h.UpdateTemplate)
	authMux.HandleFunc("/api/template/delete", h.DeleteTemplate)
	authMux.HandleFunc("/api/template/export", h.ExportTemplate)
	authMux.HandleFunc("/api/template/import", h.ImportTemplate)

	// Admin routes
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/admin", h.AdminPage)
	adminMux.HandleFunc("/api/admin/user/create", h.CreateUser)
	adminMux.HandleFunc("/api/admin/user/update", h.UpdateUser)
	adminMux.HandleFunc("/api/admin/user/delete", h.DeleteUser)

	// Apply middleware
	mux.Handle("/", middleware.AuthRequired(store)(authMux))
	mux.Handle("/logout", middleware.AuthRequired(store)(authMux))
	mux.Handle("/editor", middleware.AuthRequired(store)(authMux))
	mux.Handle("/templates", middleware.AuthRequired(store)(authMux))
	mux.Handle("/api/outline/", middleware.AuthRequired(store)(authMux))
	mux.Handle("/api/template/", middleware.AuthRequired(store)(authMux))
	mux.Handle("/admin", middleware.AdminRequired(store)(adminMux))
	mux.Handle("/api/admin/", middleware.AdminRequired(store)(adminMux))

	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start server
	port := ":8080"
	fmt.Printf("Starting server on http://localhost%s\n", port)
	fmt.Println("Default admin login: admin / admin")
	log.Fatal(http.ListenAndServe(port, mux))
}
