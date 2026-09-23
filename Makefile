.PHONY: dev build build-frontend build-backend run clean install watch tray stop

APP_NAME=app
FRONTEND_DIR=frontend

install:
	cd $(FRONTEND_DIR) && npm install

build-frontend:
	cd $(FRONTEND_DIR) && npm run build

build: build-frontend
	go build -o $(APP_NAME).exe ./...
	go build -o stop.exe ./cmd/stop

build-backend:
	go build -o $(APP_NAME).exe ./...
	go build -o stop.exe ./cmd/stop

run: build
	./$(APP_NAME).exe

tray: build
	./$(APP_NAME).exe --hide

stop:
	./stop.exe 2>nul || taskkill /IM app.exe /F 2>nul || echo "nothing to stop"

watch:
	powershell -ExecutionPolicy Bypass -File ./watch.ps1

dev:
	@echo "Jalankan 2 terminal:"
	@echo "  Terminal 1: cd frontend && npm run dev   (Vite HMR di :5173, proxy /api -> :1067)"
	@echo "  Terminal 2: go run .                     (API di :1067)"

clean:
	rm -f $(APP_NAME).exe stop.exe
	rm -rf $(FRONTEND_DIR)/dist/assets $(FRONTEND_DIR)/dist/index.html
