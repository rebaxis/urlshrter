// Package service содержит сервисы для работы с базой данных.
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

// PingDB определяет интерфейс для проверки доступности базы данных.
type PingDB interface {
	Ping() error
}

// DBChecker объединяет интерфейсы для проверки состояния БД.
// Используется для health check эндпоинта.
type DBChecker interface {
	PingDB
}

// DBService предоставляет операции для проверки состояния базы данных.
// Используется в health check эндпоинтах для мониторинга.
type DBService struct {
	repo repository.URLRepository
}

// NewDBService создает новый экземпляр DBService с указанным репозиторием.
func NewDBService(repo *repository.URLRepository) DBService {
	return DBService{
		repo: *repo,
	}
}

// Ping проверяет доступность базы данных с таймаутом 3 секунды.
// Возвращает ошибку если БД не используется или недоступна.
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
