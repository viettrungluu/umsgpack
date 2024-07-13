build:
	go build -v ./...

test:
	go test -v -coverprofile cover.out ./...

viewcoverage: test
	go tool cover -html=cover.out

format:
	gofmt -w .

checkformat:
	@test -z $(shell gofmt -l . | tee /dev/stderr)

vet:
	go vet .

fuzz:
	go test -fuzz . -fuzztime 60s

benchmark:
	go test -bench . -benchmem -benchtime 10s 
