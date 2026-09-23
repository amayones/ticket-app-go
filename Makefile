.PHONY: dev build build-frontend build-backend run clean install

APP_NAME=app
FRONTEND_DIR=frontend

install:
	cd $(FRONTEND_DIR) && npm install

build-frontend:
	cd $(FRONTEND_DIR) && npm run build

build: build-frontend
	go build -o $(APP_NAME).exe ./...

build-backend:
	go build -o $(APP_NAME).exe ./...

run: build
	./$(APP_NAME).exe

dev:
	@echo "Jalankan 2 terminal:"
	@echo "  Terminal 1: cd frontend && npm run dev   (Vite HMR di :5173, proxy /api -> :8080)"
	@echo "  Terminal 2: go run .                     (API di :8080, fallback jika dist belum ada)"

clean:
	rm -f $(APP_NAME).exe
	rm -rf $(FRONTEND_DIR)/dist/assets $(FRONTEND_DIR)/dist/index.html
