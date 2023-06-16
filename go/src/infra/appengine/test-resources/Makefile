.PHONY: help build clean

help:
	@echo "Available commands:"
	@sed -n '/^[a-zA-Z0-9_]*:/s/:.*//p' <Makefile

build:
	cd frontend && npm run build && cp -R build ../static

clean:
	rm -rf static/* && cd frontend && rm -rf build/*
