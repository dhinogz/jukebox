dev:
	docker compose up -d
	modd

prod:
	docker compose -f docker-compose.yml -f docker-compose.production.yml up --build

down:
	docker compose down --remove-orphans
