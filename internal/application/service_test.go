package application

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/tests"
)

func TestApplicationService(t *testing.T) {
	ctx := context.Background()
	pgxpool := &pgxpool.Pool{}
	queries := &database.Queries{}
	application_repository := NewRepository(ctx, queries)
	project_service := tests.StubProjectService{}

	application_service := NewService(
		ctx,
		pgxpool,
		application_repository,
		&project_service,
	)

	fmt.Println("AS", application_service)
}