package db

import (
	"context"
	"time"

	"github.com/Xacor/go-metrics/internal/logger"
	"github.com/Xacor/go-metrics/internal/server/config"
	"github.com/Xacor/go-metrics/internal/server/storage"
	"go.uber.org/zap"
)

func InitDB(cfg *config.Config) storage.Storage {
	l := logger.Get()

	var repo storage.Storage

	if cfg.DatabaseDSN != "" {
		ctx, cancelfunc := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancelfunc()
		postgre, err := storage.NewPostgreStorage(ctx, cfg.DatabaseDSN, l)
		if err != nil {
			l.Fatal("can't init db connection", zap.Error(err))
		}
		repo = postgre
	} else {
		fs, err := storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			l.Error("cant'init file storage", zap.Error(err))
		}

		repo = storage.NewMemStorage(fs, cfg.StoreInterval, l)

		if cfg.Restore {
			if err := fs.Load(repo); err != nil {
				l.Error("can't restore data from file", zap.Error(err))
			}
		}
	}

	return repo
}
