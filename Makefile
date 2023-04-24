DB_URL=postgres://root:secret@localhost:5433/key_keeper?sslmode=disable

migration_file:
	migrate create -ext sql -dir db/migrations -seq $(file_name)

sqlc:
	sqlc generate

server:
	go run main.go

postgres:
	docker run --name key_keeper_db -p 5433:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:15.2-alpine

createdb:
	docker exec -it key_keeper_db createdb --username=root --owner=root key_keeper

dropdb:
	docker exec -it key_keeper_db dropdb key_keeper

migrateup:
	migrate -path db/migrations -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migrations -database "$(DB_URL)" -verbose down

.PHONY: migration_file sqlc postgres createdb dropdb migrateup migratedown