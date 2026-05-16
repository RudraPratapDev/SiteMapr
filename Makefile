.PHONY: build run clean tidy

APP_NAME=SiteMapr

build:
	@echo "\033[1;32mBuilding $(APP_NAME)...\033[0m"
	@go build -o $(APP_NAME) main.go

run: build
	@echo "\033[1;36mRunning $(APP_NAME)...\033[0m"
	@./$(APP_NAME)

tidy:
	@echo "\033[1;33mTidying modules...\033[0m"
	@go mod tidy

clean:
	@echo "\033[1;31mCleaning up...\033[0m"
	@rm -rf $(APP_NAME) results/ results.json results.csv
