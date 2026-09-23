.PHONY: dev build build-frontend build-backend run clean install watch tray stop

ROOT := ..
APP := $(ROOT)/app.exe
STOP := $(ROOT)/stop.exe
FRONTEND_DIR := $(ROOT)/frontend

install:
	cd $(FRONTEND_DIR) && npm install

build-frontend:
	cd $(FRONTEND_DIR) && npm run build

build: build-frontend
	cd $(ROOT) && go vet ./...
	cd $(ROOT) && go build -o app.exe .
	cd $(ROOT) && go build -o stop.exe ./cmd/stop

build-backend:
	cd $(ROOT) && go vet ./...
	cd $(ROOT) && go build -o app.exe .
	cd $(ROOT) && go build -o stop.exe ./cmd/stop

run: build
	cd $(ROOT) && ./app.exe

tray: build
	cd $(ROOT) && ./app.exe --hide

stop:
	cd $(ROOT) && ./stop.exe 2>nul || taskkill /IM app.exe /F 2>nul || echo "nothing to stop"

watch:
	powershell -ExecutionPolicy Bypass -File $(CURDIR)/watch.ps1

dev:
	@echo "Jalankan 2 terminal:"
	@echo "  Terminal 1: cd frontend && npm run dev   (Vite HMR di :5173, proxy /api -> :1067)"
	@echo "  Terminal 2: go run .                     (API di :1067)"

clean:
	rm -f $(APP) $(STOP)
	rm -rf $(FRONTEND_DIR)/dist/assets $(FRONTEND_DIR)/dist/index.html
