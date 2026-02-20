-include ../.env .env
export

.PHONY: run test build clean migrate nuke

run:
	go run ./cmd/server

test:
	go test ./... -v -count=1

build:
	go build -o bin/server ./cmd/server

clean:
	rm -rf bin/

migrate:
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f migrations/001_create_guests.sql
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f migrations/002_create_users.sql

nuke:
	@if [ "$(APP_ENV)" != "test" ]; then \
		echo "Refusing to run nuke outside test environment (APP_ENV must be 'test')."; \
		exit 1; \
	fi
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -c "DROP TABLE IF EXISTS users, guests CASCADE;"
	$(MAKE) migrate
