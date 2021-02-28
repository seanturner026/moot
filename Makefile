.PHONY: build test

build:
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth         cmd/auth/. &
	env GOOS=linux go build -ldflags="-s -w" -o bin/releases     cmd/releases/. &
	env GOOS=linux go build -ldflags="-s -w" -o bin/repositories cmd/repositories/. &
	env GOOS=linux go build -ldflags="-s -w" -o bin/users        cmd/users/. &


test:
	@printf "\n"
	go test \
	./cmd/auth \
	./cmd/releases \
	./cmd/repositories \
	./cmd/users
