package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/salmanrf/capybara-cloud/internal/database"
	"github.com/salmanrf/capybara-cloud/pkg/auth"
	"github.com/salmanrf/capybara-cloud/pkg/dto"
	"github.com/salmanrf/capybara-cloud/pkg/utils"
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

func create_router(ctx context.Context, queries *database.Queries) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/auth/me", func(w http.ResponseWriter, r *http.Request) {
		sid_cookie, err := r.Cookie("sid")

		if err != nil {
			fmt.Println("Error extracting session cookie", err.Error())
			utils.ResponseWithError(w, http.StatusUnauthorized, nil, "Unauthorized")
			return
		}

		sub, err := auth.ValidateJWT(sid_cookie.Value, os.Getenv("AUTH_JWT_SECRET"))

		if err != nil {
			fmt.Println("Error validating session JWT", err.Error())	
			utils.ResponseWithError(w, http.StatusUnauthorized, nil, "Unauthorized")
			return
		}

		utils.ResponseWithSuccess(
			w, 
			http.StatusOK, 
			&map[string]any{
				"user_id": sub,
			}, 
			"Session retrieved successfully",
		)
	})

	mux.HandleFunc("POST /api/auth/signup", func(w http.ResponseWriter, r *http.Request) {
		var body dto.SignupDto

		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&body); err != nil {
			utils.ResponseWithError(w, http.StatusUnprocessableEntity, nil, "Unprocessable Entity")
			return
		}

		_, err := body.Validate()
		if err != nil {
			utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		existing, err := queries.GetUsersByEmail(ctx, body.Email)

		if err != nil {
			err_msg := err.Error()
			if !strings.Contains(err_msg, "no rows") {
				fmt.Println("Error finding user", err)
				utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
				return
			}
		}

		if existing.UserID.Valid {
			utils.ResponseWithError(w, http.StatusBadRequest, nil, "This user already exists")
			return
		}

		hashed_password, err := auth.Hash(body.Password)

		if err != nil {
			fmt.Println("Signup error when hashing user password: ", err.Error())
			utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
			return
		}

		_, err = queries.CreateOneUser(ctx, database.CreateOneUserParams{
			Email: body.Email,
			Username: body.Username,
			FullName: body.FullName,
			HashedPassword: hashed_password,
		})

		if err != nil {
			errr_msg := err.Error()
			fmt.Println("Signup error when hashing user password: ", errr_msg)

			utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
			return
		}

		utils.ResponseWithSuccess[any](w, http.StatusOK, nil, "Signed up successfully")
	})

	mux.HandleFunc("POST /api/auth/signin", func(w http.ResponseWriter, r *http.Request) {
		var body dto.SigninDto 

		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&body); err != nil {
			fmt.Println("Failed to decode request body", err)

			utils.ResponseWithError(w, http.StatusUnprocessableEntity, nil, "Unprocessable Entity")
			return
		}

		_, err := body.Validate()
		if err != nil {
			utils.ResponseWithError(w, http.StatusBadRequest, nil, err.Error())
			return
		}

		user, err := queries.GetUsersByEmail(ctx, body.Email)

		if err != nil {
			fmt.Println("Error finding user", err)
			
			utils.ResponseWithError(w, http.StatusBadRequest, nil, "Incorrect username/email")
			return
		}

		password_match, err := auth.HashCompare(body.Password, user.HashedPassword)

		if err != nil {
			fmt.Println("Error password compare", err)
			
			utils.ResponseWithError(w, http.StatusBadRequest, nil, "Incorrect username/email")
			return
		}

		if !password_match {
			utils.ResponseWithError(w, http.StatusBadRequest, nil, "Incorrect username/email")
			return
		}

		jwt_string, err := auth.MakeJWT(user.UserID, os.Getenv("AUTH_JWT_SECRET"), time.Hour * 24) 

		if err != nil {
			fmt.Println("Error building jwt for signin", err.Error())
			
			utils.ResponseWithError(w, http.StatusInternalServerError, nil, "Internal server error")
			return
		}

		sid_cookie := http.Cookie{
			Name: "sid",
			Value: jwt_string,
			Path: "/",
			SameSite: http.SameSiteStrictMode,
			MaxAge: 3600 * 24,
			HttpOnly: true,
			Secure: os.Getenv("STAGE") != "local",
		}

		http.SetCookie(w, &sid_cookie)

		utils.ResponseWithSuccess[any](
			w, 
			http.StatusOK, 
			nil,
			"Signed in successfully",
		)
	})

	return mux
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

	mux := create_router(ctx, queries)
	server := http.Server{
		Addr: ":" + "8080",
		Handler: mux,
	}
	fmt.Println("Setup completed successfully")

	server.ListenAndServe()
}