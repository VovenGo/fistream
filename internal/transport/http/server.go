package httptransport

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/fistream/fistream/internal/config"
	"github.com/fistream/fistream/internal/domain"
	"github.com/fistream/fistream/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	cfg     config.Config
	rooms   *service.RoomService
	webRoot string
	buildID string
}

func NewServer(cfg config.Config, rooms *service.RoomService) *Server {
	return &Server{
		cfg:     cfg,
		rooms:   rooms,
		webRoot: "web",
		buildID: cfg.AppBuildID,
	}
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Client-Fingerprint"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/config", s.clientConfig)
		r.Post("/rooms/create", s.createRoom)
		r.Post("/rooms/join", s.joinRoom)
		r.Post("/rooms/{code}/close", s.closeRoom)
	})

	r.Get("/room/{code}", s.serveIndex)
	r.Get("/", s.serveIndex)
	r.Handle("/assets/*", http.HandlerFunc(s.serveAssets))
	return r
}

func (s *Server) clientConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"build": map[string]any{"id": s.buildID},
		"jitsi": map[string]any{"domain": strings.TrimSpace(s.cfg.JitsiDomain)},
		"features": map[string]any{
			"room_password": false,
			"registration":  false,
		},
	})
}

func (s *Server) createRoom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DisplayName     string `json:"display_name"`
		ServicePassword string `json:"service_password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	access, err := s.rooms.CreateRoom(r.Context(), req.DisplayName, req.ServicePassword, clientFingerprint(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, access)
}

func (s *Server) joinRoom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DisplayName     string `json:"display_name"`
		RoomCode        string `json:"room_code"`
		ServicePassword string `json:"service_password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	access, err := s.rooms.JoinRoom(r.Context(), req.DisplayName, req.RoomCode, req.ServicePassword, clientFingerprint(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, access)
}

func (s *Server) closeRoom(w http.ResponseWriter, r *http.Request) {
	err := s.rooms.CloseRoom(r.Context(), chi.URLParam(r, "code"), bearerToken(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "closed", "room_code": strings.ToUpper(chi.URLParam(r, "code"))})
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.webRoot, "index.html"))
}

func (s *Server) serveAssets(w http.ResponseWriter, r *http.Request) {
	file := strings.TrimPrefix(r.URL.Path, "/assets/")
	if file == "" || strings.Contains(file, "..") {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(s.webRoot, file))
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad_request", "message": "invalid json body"})
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad_request", "message": "invalid json body"})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, err error) {
	if mapped, ok := mapDomainError(err); ok {
		writeJSON(w, mapped.status, map[string]string{"error": mapped.code, "message": mapped.message})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error", "message": "internal server error"})
}

type domainHTTPError struct {
	status  int
	code    string
	message string
}

func mapDomainError(err error) (domainHTTPError, bool) {
	catalog := []struct {
		target error
		http   domainHTTPError
	}{
		{
			target: domain.ErrInvalidServicePassword,
			http:   domainHTTPError{status: http.StatusUnauthorized, code: "invalid_service_password", message: "service password is invalid"},
		},
		{
			target: domain.ErrInvalidDisplayName,
			http:   domainHTTPError{status: http.StatusBadRequest, code: "invalid_display_name", message: "display_name is required"},
		},
		{
			target: domain.ErrInvalidRoomCode,
			http:   domainHTTPError{status: http.StatusBadRequest, code: "invalid_room_code", message: "room_code is invalid"},
		},
		{
			target: domain.ErrRoomNotFound,
			http:   domainHTTPError{status: http.StatusNotFound, code: "room_not_found", message: "room not found"},
		},
		{
			target: domain.ErrRoomClosed,
			http:   domainHTTPError{status: http.StatusGone, code: "room_closed", message: "room is closed"},
		},
		{
			target: domain.ErrRoomExpired,
			http:   domainHTTPError{status: http.StatusGone, code: "room_expired", message: "room expired"},
		},
		{
			target: domain.ErrInvalidToken,
			http:   domainHTTPError{status: http.StatusUnauthorized, code: "invalid_token", message: "authorization token is invalid"},
		},
		{
			target: domain.ErrNotHost,
			http:   domainHTTPError{status: http.StatusForbidden, code: "not_room_host", message: "only room host can close the room"},
		},
	}

	for _, item := range catalog {
		if errors.Is(err, item.target) {
			return item.http, true
		}
	}
	return domainHTTPError{}, false
}

func bearerToken(r *http.Request) string {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorization == "" {
		return ""
	}
	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func clientFingerprint(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("X-Client-Fingerprint"))
	if len(value) > 255 {
		value = value[:255]
	}
	return value
}
