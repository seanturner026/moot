.PHONY: build clean deploy gomodgen

build: gomodgen
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/create_user         src/functions/create_user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/delete_user         src/functions/delete_user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/list_users          src/functions/list_users/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/login_user          src/functions/login_user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/release             src/functions/release/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/reset_user_password src/functions/reset_user_password/main.go

clean:
	rm -rf ./bin ./vendor go.sum

deploy: clean build
	sls deploy --verbose  --aws-profile sean

gomodgen:
	chmod u+x gomod.sh
	./gomod.sh

destroy:
	serverless remove --verbose --aws-profile sean
