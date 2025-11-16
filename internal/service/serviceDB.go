package service

import (
	"context"
	"fmt"
	"time"

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

func (s DBService) CloseDB() error {
	if s.repo.UseDB {
		if err := s.repo.StorageDB.Close(); err != nil {
			return err
		}
	}
	return nil
}
