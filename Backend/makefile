#! loading .env
include .env

fix :
	go mod tidy

help :
	@echo MAKEFILE COMMANDS LIST 
	@echo fix - fix package not found issues
	@echo build - builds go project 
	@echo run - builds & run go project
	@echo stop - stops the api server

# ! building & running {-s flag for silent output}
build :
	@echo FOUND AND DELETING OLD DEPRICIATED "$(BINARY)"
	if exist "$(BINARY)" del "$(BINARY)"
	@echo started building new binary...
	go build -o $(BINARY) cmd/server/main.go
	@echo "binary created in backend root."

run : build
	$(BINARY)
	@echo "blog api has started🚀..."

stop : 
	@echo "stopping whole backend servers❌..."
	 taskkill /IM $(BINARY) /F 2>nul || echo not running
	@echo "Server stopped successfully🪧."



