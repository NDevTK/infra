.PHONY: shivas
shivas:
	@$(RM) shivas dev-shivas
	go test ./...
	go build -o shivas main.go
	go build -o dev-shivas -tags=dev main.go

# Build a version of shivas targeted specifically
# at a version of UFS running locally.
.PHONY: dev
dev:
	go test ./...
	go build -tags=dev -o dev-shivas main.go

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	$(RM) shivas
	$(RM) dev-shivas
