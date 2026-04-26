-include ../.env .env
export

.PHONY: run test test-integration test-all build clean migrate nuke

run:
	go run ./cmd/server

test:
	go test ./... -v -count=1

test-integration:
	go test ./... -v -count=1 -tags=integration -run '^TestIntegration'

test-all: test test-integration

build:
	go build -o bin/server ./cmd/server

clean:
	rm -rf bin/

migrate:
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f migrations/001_create_guests.sql
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f migrations/002_create_users.sql
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f migrations/003_create_otp_codes.sql
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f migrations/004_create_audit_log.sql
	@PGPASSWORD=${DB_PASSWORD} psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f migrations/005_single_groom_bride.sql
	@PGPASSWORD=${DB_PASSWORD} psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f migrations/006_create_gifts.sql
	@PGPASSWORD=${DB_PASSWORD} psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f migrations/007_payment_idempotency.sql

nuke:
	@if [ "$(APP_ENV)" != "test" ]; then \
		echo "Refusing to run nuke outside test environment (APP_ENV must be 'test')."; \
		exit 1; \
	fi
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -c "DROP TABLE IF EXISTS gift_transactions, gifts, audit_log, otp_codes, users, guests CASCADE;"
	$(MAKE) migrate
