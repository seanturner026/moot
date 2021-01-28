.PHONY: build test deploy destroy

build:
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/releases_create      cmd/releases_create/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/repositories_list    cmd/repositories_list/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/repositories_create  cmd/repositories_create/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/repositories_delete  cmd/repositories_delete/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/users_create         cmd/users_create/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/users_delete         cmd/users_delete/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/users_list           cmd/users_list/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/users_login          cmd/users_login/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/users_reset_password cmd/users_reset_password/main.go


test:
	@printf "\n"
	go test \
	./cmd/releases_create \
	./cmd/repositories_create \
	./cmd/repositories_delete \
	./cmd/repositories_list \
	./cmd/users_create \
	./cmd/users_delete \
	./cmd/users_list \
	./cmd/users_login \
	./cmd/users_reset_password \

compress:
	@printf "\n"
	zip -q archive/functions/ReleasesCreate.zip     bin/functions/releases_create &
	zip -q archive/functions/RepositoriesCreate.zip bin/functions/repositories_create &
	zip -q archive/functions/RepositoriesDelete.zip bin/functions/repositories_delete &
	zip -q archive/functions/RepositoriesList.zip   bin/functions/repositories_list &
	zip -q archive/functions/UsersCreate.zip        bin/functions/users_create &
	zip -q archive/functions/UsersDelete.zip        bin/functions/users_delete &
	zip -q archive/functions/UsersList.zip          bin/functions/users_list &
	zip -q archive/functions/UsersLogin.zip         bin/functions/users_login &
	zip -q archive/functions/UsersResetPassword.zip bin/functions/users_reset_password &

	@printf "\n"
	zip -q -j archive/custom_resources/CFNGetCognitoClientSecret.zip \
		deployments/custom_resources/get_cognito_client_secret.py \
		deployments/custom_resources/cfnresponse.py

deploy: build test compress
	@set -e
	@printf "\n"
	serverless deploy --verbose  --aws-profile ${AWS_DEFAULT_PROFILE}

destroy:
	serverless remove --verbose --aws-profile ${AWS_DEFAULT_PROFILE}
