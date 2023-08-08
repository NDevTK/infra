.PHONY: help build clean

help:
	@echo "Available commands:"
	@sed -n '/^[a-zA-Z0-9_]*:/s/:.*//p' <Makefile

build: clean
	cd frontend && npm ci && npm run build && cp -R build ../static

clean:
	rm -rf static && rm -rf frontend/build
