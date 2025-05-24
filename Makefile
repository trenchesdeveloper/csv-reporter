API_FILES := $(filter-out %_test.go,$(wildcard cmd/api/*.go))
.PHONY: postgres makepostgres droppostgres createdb dropdb migrateup migratedown migratedown migratecreate sqlc server startWorker test tf-init tf-plan tf-apply gen-docs

# Docker commands
makepostgres:
	docker compose up -d

droppostgres:
	docker compose down

postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=password -d postgres:14-alpine

# Database commands
createdb:
	docker exec -it goflow_postgres createdb --username=root --owner=root goflow

dropdb:
	docker exec -it goflow_postgres dropdb goflow

# Migration commands
migrateup:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/goflow?sslmode=disable" -verbose up
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5433/goflow_test?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/goflow?sslmode=disable" -verbose down
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5433/goflow_test?sslmode=disable" -verbose down

migratecreate:
	# Create a new migration file
	migrate create -ext sql -dir db/migrations $(name)

# Code generation
sqlc:
	sqlc generate

gen-docs:
	swag init -g ./api/main.go -d cmd && swag fmt

# Run commands
server: gen-docs
	nodemon \
	  --watch './**/*.go' \
	  --signal SIGTERM \
	  --exec "APP_ENV=dev go run $(API_FILES)"

startWorker:
	go run cmd/worker/main.go

# Testing
test:
	go test -v -cover ./...

# Terraform commands
tf-init:
	cd terraform && terraform init

tf-plan:
	cd terraform && terraform plan

tf-apply:
	cd terraform && terraform apply