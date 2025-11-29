package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/salmanrf/capybara-cloud/api"
	"github.com/salmanrf/capybara-cloud/internal/auth"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/internal/organization"
	"github.com/salmanrf/capybara-cloud/internal/user"
	auth_utils "github.com/salmanrf/capybara-cloud/pkg/auth"
)

func setup() (context.Context, *pgx.Conn, error) {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		return ctx, nil, err
	}

	postgres_uri := os.Getenv("POSTGRES_URI")

	conn, err := pgx.Connect(ctx, postgres_uri)
	if err != nil {
		return ctx, nil, err
	}

	err = conn.Ping(ctx)

	if err != nil {
		return ctx, nil, err
	}

	fmt.Println("Database connection established")

	return ctx, conn, nil
}

func main() {
	ctx, db_conn, err := setup()
	defer db_conn.Close(ctx)
	defer func() {
		fmt.Println("Server is stopped")
	}()
	if err != nil {
		log.Fatal(err)
	}

	queries := database.New(db_conn)
	user_service := user.NewService(ctx, queries)
	auth_service := auth.NewService(ctx, user_service)
	org_service := organization.NewService(ctx, queries, user_service)
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