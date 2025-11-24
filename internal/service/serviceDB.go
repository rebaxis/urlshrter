package service

import (
	"context"
	"fmt"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rebaxis/urlshrter/internal/repository"
)

type PingDB interface {
	Ping() error
}

type DBChecker interface {
	PingDB
}

type DBService struct {
	repo repository.URLRepository
}

func NewDBService(repo *repository.URLRepository) DBService {
	return DBService{
		repo: *repo,
	}
}

func (s DBService) Ping() error {
	if s.repo.UseDB {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.repo.StorageDB.PingContext(ctx); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("we don't use DB")
	}
	return nil
}
