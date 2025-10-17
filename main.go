package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/salmanrf/capybara-cloud/internal/database"
)

func setup() (*pgx.Conn, error) {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	postgres_uri := os.Getenv("POSTGRES_URI")

	conn, err := pgx.Connect(ctx, postgres_uri)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(ctx)

	if err != nil {
		return nil, err
	}

	fmt.Println("Database connection established")

	return conn, nil
}

func main() {
	db_conn, err := setup()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Setup completed successfully")
	
	ctx := context.Background()

	defer db_conn.Close(ctx)

	queries := database.New(db_conn)

	user, err := queries.CreateOneUser(ctx, database.CreateOneUserParams{
		Username: "frnamlas",
		Email: "salmanrf2@gmail.com",
		FullName: "Salman RF",
	})

	if err != nil {
		log.Fatal("Error Creating user ", err.Error())
	}

	fmt.Println("Created user", user)
}