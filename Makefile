.PHONY: all up migrate generate

all: up migrate

up:
	docker-compose up -d

down:
	docker-compose down

migrate:
	docker-compose exec database sh -c 'until pg_isready -U casino -h database; do sleep 2; done && psql -U casino -h database -d casino -f /db/migrations/00001.create_base.sql'

generator:
	docker-compose run --rm generator

generator-logs:
	docker-compose logs -f generator