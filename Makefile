run:
	go run cmd/main.go

migrate_up:
	migrate -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -path build/migrations up


migrate_down:
	migrate -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -path build/migrations down


get_coverage:
	go test ./... -coverprofile cover.out
	go tool cover -html=cover.out