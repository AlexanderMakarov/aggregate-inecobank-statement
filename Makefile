run:
	go run .

test:
	go test -coverprofile coverage.out

coverage:
	go tool cover -html=coverage.out

build:
	go build
