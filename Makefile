server: gen-docs
	nodemon --watch './**/*.go' --signal SIGTERM --exec APP_ENV=dev 'go' run cmd/api/*.go

migratecreate:
	# Create a new migration file
	migrate create -ext sql -dir db/migrations $(name)

makepostgres:
	docker compose up -d

droppostgres:
	docker compose down

createdb:
	docker exec -it goflow_postgres createdb --username=root --owner=root goflow

dropdb:
	docker exec -it goflow_postgres dropdb goflow
migrateup:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/goflow?sslmode=disable" -verbose up
	#migrate -path db/migrations -database "postgresql://root:secret@localhost:5434/goflow?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/goflow?sslmode=disable" -verbose down
	#migrate -path db/migrations -database "postgresql://root:secret@localhost:5434/goflow?sslmode=disable" -verbose down

gen-docs:
	swag init -g ./api/main.go -d cmd,internal && swag fmt


sqlc:
	sqlc generate