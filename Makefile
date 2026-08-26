.PHONY: test test-race integ vet fmt clean cover

# Unit tests (default; no credentials, no integration build tag).
test:
	go test -count=1 ./...

# Unit tests with race detector and coverage report.
test-race:
	go test -race -count=1 -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out | tail -1

# Integration tests against the real gateway. Requires credentials via
# TIGEROPEN_TIGER_ID / PRIVATE_KEY / ACCOUNT env vars (or TIGEROPEN_PROPS_PATH).
# Export TIGER_RUN_INTEG=true to actually exercise the endpoints; otherwise
# the tests compile but are skipped.
integ:
	TIGER_RUN_INTEG=true go test -tags=integration -count=1 -v ./quote/ ./trade/

vet:
	go vet ./...

fmt:
	gofmt -w .

cover:
	go test -count=1 -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out | tail -1

clean:
	rm -f coverage.out
