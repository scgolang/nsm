coverage:
	@go test -coverprofile cover.out && go tool cover -html cover.out

test:
	@go test

.PHONY: coverage test
