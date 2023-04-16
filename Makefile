test:
	go test . -coverprofile=./coverage.out
	go tool cover -html=./coverage.out -o ./coverage.html

test-detail:
	go test

open-coverage:
	start $(shell pwd)/coverage.html