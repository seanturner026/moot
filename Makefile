.PHONY: build clean deploy gomodgen

define NEWLINE

endef

build: gomodgen
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/create_user         src/create_user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/delete_user         src/delete_user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/list_users          src/list_users/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/login_user          src/login_user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/release             src/release/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/reset_user_password src/reset_user_password/main.go \
	$(NEWLINE)

test:
	go test ./src/create_user \
	./src/delete_user \
	./src/list_users \
	./src/login_user \
	./src/release \
	./src/reset_user_password \
	$(NEWLINE)

clean:
	rm -rf ./bin ./vendor go.sum \
	$(NEWLINE)

deploy: clean build test
	sls deploy --verbose  --aws-profile sean \
	$(NEWLINE)

gomodgen:
	chmod u+x gomod.sh
	./gomod.sh \
	$(NEWLINE)

destroy:
	serverless remove --verbose --aws-profile sean
