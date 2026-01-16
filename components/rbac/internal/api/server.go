package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/davidliyutong/idekube-rbac/internal/permission"
	"github.com/davidliyutong/idekube-rbac/pkg/logger"
)

// Server exposes HTTP endpoints for permission checks.
type Server struct {
	httpServer *http.Server
	perm       *permission.PermissionService
	log        *logger.Logger
}

// @title IDEKube RBAC API
// @version 1.0
// @description RBAC service for managing access control in the idekube platform
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@idekube.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
func NewServer(port int, perm *permission.PermissionService, log *logger.Logger) *Server {
	mux := http.NewServeMux()
	srv := &Server{perm: perm, log: log}

	mux.HandleFunc("/healthz", srv.handleHealth)
	mux.HandleFunc("/api/v1/rbac/check", srv.handleCheckPermission)

	srv.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return srv
}

func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Infof("RBAC HTTP server listening on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		return err
	}
}

// @Summary Health check
// @Description Check if the service is running
// @Tags health
// @Produce plain
// @Success 200 {string} string "ok"
// @Router /healthz [get]
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// @Summary Check permission
// @Description Check if a user has permission to perform an action on a resource
// @Tags rbac
// @Accept json
// @Produce json
// @Param request body permission.CheckPermissionRequest true "Permission check request"
// @Success 200 {object} map[string]bool "allowed: true/false"
// @Failure 400 {string} string "Invalid request or permission check failed"
// @Failure 405 {string} string "Method not allowed"
// @Router /rbac/check [post]
func (s *Server) handleCheckPermission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "METHOD_NOT_ALLOWED",
				"message": "only POST method is allowed",
			},
		})
		return
	}

	var req permission.CheckPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("invalid request body: %v", err),
			},
		})
		return
	}

	allowed, err := s.perm.CheckPermission(r.Context(), req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "PERMISSION_CHECK_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	s.writeJSON(w, map[string]any{"success": true, "allowed": allowed})
}

func (s *Server) writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(payload); err != nil {
		s.log.Errorf("failed to write JSON response: %v", err)
	}
}
