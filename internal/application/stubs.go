package application

import (
	"github.com/salmanrf/capybara-cloud/internal/database"
)

type stub_repository struct {
	queries *database.Queries
}

type StubApplicationRepository interface {
	CreateConfig(database.CreateApplicationConfigParams) (*database.ApplicationConfig, error) 
}