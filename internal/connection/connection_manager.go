package connection

import (
	"sync"
	"time"

	"github.com/CaioMtho/pinoql-mcp/internal/adapters"
	"github.com/CaioMtho/pinoql-mcp/internal/adapters/postgres"
	"github.com/CaioMtho/pinoql-mcp/internal/errors"
)

type Manager struct {
	mu       sync.Mutex
	adapters map[string]adapters.Adapter
}

func NewConnectionManager() *Manager {
	return &Manager{
		adapters: make(map[string]adapters.Adapter),
	}
}

func (cm *Manager) GetAdapter(cfg Config) (adapters.Adapter, error) {
	key := string(cfg.Dialect) + "|" + cfg.DSN

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if adapter, ok := cm.adapters[key]; ok {
		return adapter, nil
	}

	var adapter adapters.Adapter
	var err error
	switch cfg.Dialect {
	case PostgreSQL:
		adapter, err = postgres.NewPostgresAdapter(cfg.DSN)

	default:
		return nil, errors.InvalidDialectError{DialectInput: cfg.DSN, ValidDialects: GetDialects()}
	}

	if err != nil {
		return nil, err
	}

	db := adapter.GetDB()

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(1 * time.Hour)

	cm.adapters[key] = adapter
	return adapter, nil
}

func (cm *Manager) CloseAll() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	for _, adapter := range cm.adapters {
		err := adapter.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
