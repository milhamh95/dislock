.PHONY: redis-up
redis-up: ## Start redis
	@docker-compose up -d redis

.PHONY: redis-down
redis-down: ## Shutdown redis
	@docker stop local_redis
