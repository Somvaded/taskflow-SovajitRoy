db_url := env_var_or_default("DB_CONNECTION_STR", "postgresql://postgres:postgres@localhost:5432/taskflow?sslmode=disable")

# List available recipes
default:
    @just --list

# Start the local database container
db:
    docker compose up -d db

# Run all SQL migrations
migrate:
    @for f in migrations/*.up.sql; do \
        echo "Applying $f..."; \
        psql "{{db_url}}" -f "$f"; \
    done

# Load seed data (test@example.com / password123)
seed:
    psql "{{db_url}}" -f migrations/seed.sql

# Start the API server
run:
    go run ./cmd/server

# Start the database, apply migrations, then run the server
dev: db
    @echo "Waiting for database to be ready..."
    @until psql "{{db_url}}" -c '\q' 2>/dev/null; do sleep 1; done
    @just migrate
    @just run
