package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/salmanrf/capybara-cloud/api"
	"github.com/salmanrf/capybara-cloud/internal/auth"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/internal/organization"
	"github.com/salmanrf/capybara-cloud/internal/user"
	auth_utils "github.com/salmanrf/capybara-cloud/pkg/auth"
)

func create_db_conn(ctx context.Context, db_uri string) *pgxpool.Pool {
	dbpool, err := pgxpool.New(ctx, db_uri)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	return dbpool
}

func setup() (context.Context, *pgxpool.Pool, error) {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Unable to load env vars")
		os.Exit(1)
	}

	postgres_uri := os.Getenv("POSTGRES_URI")
	dbpool := create_db_conn(ctx, postgres_uri)

	err = dbpool.Ping(ctx)

	if err != nil {
		fmt.Println("Unable to ping database")
		os.Exit(1)
	}

	fmt.Println("Database connection established")

	return ctx, dbpool, nil
}

func main() {
	ctx, db_conn, err := setup()
	defer db_conn.Close()
	defer func() {
		fmt.Println("Server is stopped")
	}()
	if err != nil {
		log.Fatal(err)
	}

	queries := database.New(db_conn)
	user_service := user.NewService(ctx, queries)
	auth_service := auth.NewService(ctx, user_service)
	org_service := organization.NewService(ctx, db_conn, queries, user_service)
	jwt_utils := auth_utils.NewJWTUtils(os.Getenv("AUTH_JWT_SECRET"))
	
	api_server := api.NewAPIServer(
		ctx, 
		user_service, 
		auth_service, 
		org_service,
		jwt_utils,
	)
	
	api_port := os.Getenv("API_PORT")
	http.ListenAndServe(fmt.Sprintf(":%s", api_port), api_server)
}