.PHONY: build test deploy destroy

build:
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/create_repo         cmd/create_repo/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/create_user         cmd/create_user/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/delete_repo         cmd/delete_repo/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/delete_user         cmd/delete_user/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/list_repos          cmd/list_repos/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/list_users          cmd/list_users/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/login_user          cmd/login_user/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/release             cmd/release/main.go &
	env GOOS=linux go build -ldflags="-s -w" -o bin/functions/reset_user_password cmd/reset_user_password/main.go &


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

compress:
	@printf "\n"
	zip -q archive/functions/CreateRepo.zip        bin/functions/create_repo  &
	zip -q archive/functions/CreateUser.zip        bin/functions/create_user &
	zip -q archive/functions/DeleteRepo.zip        bin/functions/delete_repo &
	zip -q archive/functions/DeleteUser.zip        bin/functions/delete_user &
	zip -q archive/functions/ListRepos.zip         bin/functions/list_repos &
	zip -q archive/functions/ListUsers.zip         bin/functions/list_users &
	zip -q archive/functions/LoginUser.zip         bin/functions/login_user &
	zip -q archive/functions/Release.zip           bin/functions/release &
	zip -q archive/functions/ResetUserPassword.zip bin/functions/reset_user_password &

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
