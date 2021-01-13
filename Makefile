.PHONY: build clean deploy gomodgen

build: gomodgen
	@printf "\n"
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/create_repo         src/create_repo/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/create_user         src/create_user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/delete_repo         src/delete_repo/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/delete_user         src/delete_user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/list_repos          src/list_repos/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/list_users          src/list_users/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/login_user          src/login_user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/release             src/release/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/reset_user_password src/reset_user_password/main.go

test:
	@printf "\n"
	go test \
	./src/create_repo \
	./src/create_user \
	./src/delete_repo \
	./src/delete_user \
	./src/list_repos \
	./src/list_users \
	./src/login_user \
	./src/release \
	./src/reset_user_password \

clean:
	rm -rf ./bin ./vendor go.sum

deploy: clean build test
	@printf "\n"
	sls deploy --verbose  --aws-profile sean

gomodgen:
	chmod u+x gomod.sh
	./gomod.sh

destroy:
	serverless remove --verbose --aws-profile sean
