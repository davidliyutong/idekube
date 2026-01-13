package rbac

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	"github.com/davidliyutong/idekube-rbac/pkg/logger"
)

func newEnforcer(db *gorm.DB, modelPath, policyPath string, log *logger.Logger) (*casbin.Enforcer, error) {
	if db == nil {
		return nil, errors.New("gorm db is required")
	}
	if strings.TrimSpace(modelPath) == "" {
		return nil, errors.New("casbin model path is required")
	}

	cleanedModel := filepath.Clean(modelPath)
	m, err := model.NewModelFromFile(cleanedModel)
	if err != nil {
		return nil, fmt.Errorf("load casbin model: %w", err)
	}

	adapter, err := gormadapter.NewAdapterByDBUseTableName(db, "", "casbin_rule")
	if err != nil {
		return nil, fmt.Errorf("create casbin adapter: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("load casbin policy: %w", err)
	}

	if len(enforcer.GetPolicy()) == 0 && strings.TrimSpace(policyPath) != "" {
		if err := seedPolicyFromFile(enforcer, filepath.Clean(policyPath), log); err != nil {
			return nil, err
		}
	}

	return enforcer, nil
}

func seedPolicyFromFile(enforcer *casbin.Enforcer, policyPath string, log *logger.Logger) error {
	file, err := os.Open(policyPath)
	if err != nil {
		return fmt.Errorf("open casbin policy file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("read casbin policy file: %w", err)
	}

	added := 0
	for _, record := range records {
		if len(record) == 0 {
			continue
		}
		// Support comments in the policy seed file
		if strings.HasPrefix(strings.TrimSpace(record[0]), "#") {
			continue
		}
		if len(record) < 3 {
			continue
		}
		// Prepare params as []interface{} for Casbin v2 API
		params := make([]interface{}, 3)
		params[0] = strings.TrimSpace(record[0])
		params[1] = strings.TrimSpace(record[1])
		params[2] = strings.TrimSpace(record[2])
		
		if enforcer.HasPolicy(params) {
			continue
		}
		if _, err := enforcer.AddPolicy(params); err != nil {
			return fmt.Errorf("seed casbin policy: %w", err)
		}
		added++
	}

	if added > 0 {
		if err := enforcer.SavePolicy(); err != nil {
			return fmt.Errorf("persist casbin policy: %w", err)
		}
		if log != nil {
			log.Infof("Seeded %d casbin policies from %s", added, policyPath)
		}
	}

	return nil
}
