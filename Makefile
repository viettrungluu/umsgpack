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

profile_unmarshal:
	go test -bench UnmarshalBytes -benchmem -benchtime 10s -cpuprofile profile_unmarshal.out -run ^$$ .
	go tool pprof -http=localhost:8080 profile_unmarshal.out

memprofile_unmarshal:
	go test -bench UnmarshalBytes -benchmem -benchtime 10s -memprofile memprofile_unmarshal.out -run ^$$ .
	go tool pprof -http=localhost:8080 memprofile_unmarshal.out

profile_marshal:
	go test -bench MarshalToBytes -benchmem -benchtime 10s -cpuprofile profile_marshal.out -run ^$$ .
	go tool pprof -http=localhost:8080 profile_marshal.out

memprofile_marshal:
	go test -bench MarshalToBytes -benchmem -benchtime 10s -memprofile memprofile_marshal.out -run ^$$ .
	go tool pprof -http=localhost:8080 memprofile_marshal.out
