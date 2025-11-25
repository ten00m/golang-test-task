.PHONY: help build up down restart logs clean

help: ## Показать справку
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать Docker образы
	docker-compose build

up: ## Запустить приложение
	docker-compose up -d

down: ## Остановить приложение
	docker-compose down

restart: down up ## Перезапустить приложение

logs: ## Показать логи
	docker-compose logs -f

logs-app: ## Показать логи приложения
	docker-compose logs -f app

logs-db: ## Показать логи базы данных
	docker-compose logs -f db

clean: ## Остановить приложение и удалить volumes
	docker-compose down -v

ps: ## Показать статус контейнеров
	docker-compose ps

rebuild: ## Пересобрать и запустить приложение
	docker-compose up -d --build
