module github.com/salmanrf/capybara-cloud/masbro-manager

go 1.26.4

replace github.com/salmanrf/capybara-cloud/packages/shared-go => ../shared

require (
	github.com/redis/go-redis/v9 v9.21.0
	github.com/salmanrf/capybara-cloud/packages/shared-go v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/jaevor/go-nanoid v1.4.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
)
