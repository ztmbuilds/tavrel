WEB_DIR := web
WEB_DIST := internal/web/dist
APP_BIN := tmp/main

.PHONY: help install dev dev-frontend dev-backend compose-dev build build-frontend build-backend check-frontend vet test clean

help:
	@echo "Targets:"
	@echo "  install          pnpm install for $(WEB_DIR)/"
	@echo "  dev              run frontend + backend together (Ctrl-C stops both)"
	@echo "  dev-frontend     vite dev on :5173, proxies /api/v1 to APP_PORT"
	@echo "  dev-backend      air live-reload for the Go app"
	@echo "  compose-dev      build tavrel:local and run the compose stack"
	@echo "  build            frontend + backend → single binary at $(APP_BIN)"
	@echo "  build-frontend   vite build into $(WEB_DIST)"
	@echo "  build-backend    go build ./cmd/tavrel only (needs build-frontend for UI)"
	@echo "  check-frontend   eslint + typecheck for $(WEB_DIR)/"
	@echo "  vet              go vet ./..."
	@echo "  test             go test ./..."
	@echo "  clean            remove build outputs (keeps $(WEB_DIST)/.gitkeep)"

install:
	pnpm --dir $(WEB_DIR) install

dev:
	$(MAKE) -j2 dev-frontend dev-backend

dev-frontend:
	pnpm --dir $(WEB_DIR) dev

dev-backend:
	air

compose-dev:
	docker compose -f compose.yaml -f compose.dev.yaml up -d --build

build: build-frontend build-backend

build-frontend:
	pnpm --dir $(WEB_DIR) build
	mkdir -p $(WEB_DIST)
	touch $(WEB_DIST)/.gitkeep

build-backend:
	go build -o $(APP_BIN) ./cmd/tavrel

check-frontend:
	pnpm --dir $(WEB_DIR) lint
	pnpm --dir $(WEB_DIR) typecheck

vet:
	go vet ./...

test:
	go test ./...

clean:
	find $(WEB_DIST) -mindepth 1 ! -name '.gitkeep' -delete
	rm -f $(APP_BIN)
