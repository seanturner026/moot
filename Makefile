.PHONY: build test deploy destroy

build:
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/create_repo         cmd/create_repo/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/create_user         cmd/create_user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/delete_repo         cmd/delete_repo/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/delete_user         cmd/delete_user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/list_repos          cmd/list_repos/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/list_users          cmd/list_users/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/login_user          cmd/login_user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/release             cmd/release/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/reset_user_password cmd/reset_user_password/main.go

test:
	@printf "\n"
	go test \
	./cmd/create_repo \
	./cmd/create_user \
	./cmd/delete_repo \
	./cmd/delete_user \
	./cmd/list_repos \
	./cmd/list_users \
	./cmd/login_user \
	./cmd/release \
	./cmd/reset_user_password \

deploy: build test
	@printf "\n"
	serverless deploy --verbose  --aws-profile ${AWS_DEFAULT_PROFILE}

destroy:
	serverless remove --verbose --aws-profile ${AWS_DEFAULT_PROFILE}
