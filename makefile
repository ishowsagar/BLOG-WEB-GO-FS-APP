# shortcuts for better dev environment & less complexity
local_build :
	docker compose -f ./docker-compose.local.yml up --build
local_backend :
	docker compose -f ./docker-compose.local.yml build backend
local_restart :
	docker compose -f ./docker-compose.local.yml restart
local_down :
	docker compose -f ./docker-compose.local.yml down
