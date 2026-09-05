

FRONTEND_DIR := frontend
BACKEND_DIR := backend
WEB_DIST := $(BACKEND_DIR)/internal/web/dist
APP_BIN := $(BACKEND_DIR)/tmp/main

.PHONY: help install dev dev-frontend dev-backend build build-frontend build-backend check-frontend vet test clean

help:
    @echo "Targets:"
    @echo "  install          pnpm install for frontend/"
    @echo "  dev              run frontend + backend together (Ctrl-C stops both)"
    @echo "  dev-frontend     vite dev on :5173, proxies /api/v1 to APP_PORT"
    @echo "  dev-backend      air live-reload for backend/"
    @echo "  build            frontend + backend → single binary at $(APP_BIN)"
    @echo "  build-frontend   vite build + copy dist into Go embed dir"
    @echo "  build-backend    go build ./cmd only (needs build-frontend for UI)"
    @echo "  check-frontend   eslint + typecheck for frontend/"
    @echo "  vet              go vet ./... for backend/"
    @echo "  test             go test ./... for backend/"
    @echo "  clean            remove build outputs (keeps web/dist/.gitkeep)"

install:
    pnpm --dir $(FRONTEND_DIR) install

dev:
    $(MAKE) -j2 dev-frontend dev-backend

dev-frontend:
    pnpm --dir $(FRONTEND_DIR) dev

dev-backend:
    cd $(BACKEND_DIR) && air

build: build-frontend build-backend

build-frontend:
	pnpm --dir $(FRONTEND_DIR) build
	rm -rf $(WEB_DIST)
	mkdir -p $(WEB_DIST)
	cp -a $(FRONTEND_DIR)/dist/. $(WEB_DIST)/
	touch $(WEB_DIST)/.gitkeep

build-backend:
    cd $(BACKEND_DIR) && go build -o tmp/main ./cmd

check-frontend:
    pnpm --dir $(FRONTEND_DIR) lint
    pnpm --dir $(FRONTEND_DIR) typecheck

vet:
    cd $(BACKEND_DIR) && go vet ./...

test:
    cd $(BACKEND_DIR) && go test ./...

clean:
    rm -rf $(FRONTEND_DIR)/dist
    find $(WEB_DIST) -mindepth 1 ! -name '.gitkeep' -delete
    rm -f $(APP_BIN)