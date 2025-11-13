package middleware

import (
	"context"
	"net/http"

	"github.com/kristofer/composter/internal/database"
)

type contextKey string

const UserKey contextKey = "user"

type SessionStore struct {
	sessions map[string]*database.User
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*database.User),
	}
}

func (s *SessionStore) Set(sessionID string, user *database.User) {
	s.sessions[sessionID] = user
}

func (s *SessionStore) Get(sessionID string) (*database.User, bool) {
	user, ok := s.sessions[sessionID]
	return user, ok
}

func (s *SessionStore) Delete(sessionID string) {
	delete(s.sessions, sessionID)
}

func AuthRequired(store *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			user, ok := store.Get(cookie.Value)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminRequired(store *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			user, ok := store.Get(cookie.Value)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			if !user.IsAdmin {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUser(r *http.Request) (*database.User, bool) {
	user, ok := r.Context().Value(UserKey).(*database.User)
	return user, ok
}
