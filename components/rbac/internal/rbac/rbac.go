package rbac

import (
	"context"
	"errors"
	"time"

	"github.com/davidliyutong/idekube-rbac/internal/opa"
	"github.com/davidliyutong/idekube-rbac/pkg/database"
	"github.com/davidliyutong/idekube-rbac/pkg/k8s"
	"github.com/davidliyutong/idekube-rbac/pkg/logger"

	"github.com/davidliyutong/idekube-rbac/internal/api"
	"github.com/davidliyutong/idekube-rbac/internal/config"
	"github.com/davidliyutong/idekube-rbac/internal/permission"
)

type RBACService struct {
	cfg       *config.Config
	k8sClient *k8s.Client
	db        *database.PostgresClient
	log       *logger.Logger

	enforcer  *opa.Enforcer
	perm      *permission.PermissionService
	apiServer *api.Server
}

func NewRBACService(
	cfg *config.Config,
	k8sClient *k8s.Client,
	db *database.PostgresClient,
	log *logger.Logger,
) (*RBACService, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if db == nil {
		return nil, errors.New("database client is required")
	}
	if log == nil {
		return nil, errors.New("logger is required")
	}

	enforcer, err := opa.NewEnforcer(db.DB(), cfg.OPA.PolicyPath, cfg.OPA.DataPath, log)
	if err != nil {
		return nil, err
	}

	perm := permission.NewPermissionService(enforcer, log)
	apiServer := api.NewServer(cfg.HTTPPort, perm, log)

	return &RBACService{
		cfg:       cfg,
		k8sClient: k8sClient,
		db:        db,
		log:       log,
		enforcer:  enforcer,
		perm:      perm,
		apiServer: apiServer,
	}, nil
}

func (r *RBACService) Start(ctx context.Context) error {
	r.log.Info("RBAC service started")

	// Background heartbeat for liveness metrics
	go r.runHeartbeat(ctx)

	// Start HTTP API server (blocks until context cancel or error)
	return r.apiServer.Start(ctx)
}

func (r *RBACService) runHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.log.Info("RBAC heartbeat stopped")
			return
		case <-ticker.C:
			r.log.Debug("RBAC service heartbeat")
		}
	}
}
