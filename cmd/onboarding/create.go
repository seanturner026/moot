package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/google/uuid"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
	log "github.com/sirupsen/logrus"
)

type onboardingEvent struct {
	PK             string `dynamodbav:"PK" json:",omitempty"`
	TenantName     string `dynamodbav:"SK" json:"organization_name"`
	TenantID       string `dynamodbav:"ID" json:"tenant_id,omitempty"`
	AvailableSeats string `dynamodbav:"AvailableSeats" json:"available_seats"`
	ContactEmail   string `dynamodbav:"Contact" json:"contact_email"`
	OnboardingDate string `dynamodbav:"OnboardingDate" json:",omitempty"`
	Status         string `dynamodbav:"Status" json:"status,omitempty"`
}

type onboardingChannel struct {
	Error error
	Type  string
}

func generatePutItemInputExpression(e onboardingEvent) (map[string]*dynamodb.AttributeValue, error) {
	itemInput, err := dynamodbattribute.MarshalMap(e)
	if err != nil {
		return map[string]*dynamodb.AttributeValue{}, err
	}
	return itemInput, nil
}

func (app application) writeOrgToDynamoDB(e onboardingEvent, itemInput map[string]*dynamodb.AttributeValue) error {
	input := &dynamodb.PutItemInput{
		Item: itemInput,
		ConditionExpression: aws.String(
			"attribute_not_exists(PK) AND attribute_not_exists(SK) AND attribute_not_exists(Contact)",
		),
		ReturnConsumedCapacity:      aws.String("NONE"),
		ReturnItemCollectionMetrics: aws.String("NONE"),
		ReturnValues:                aws.String("NONE"),
		TableName:                   aws.String(app.Config.TableName),
	}

	_, err := app.Config.DB.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(aerr.Error())
		} else {
			log.Error(err.Error())
		}
		return err
	}
	return nil
}

var cognitoUsername string

func (app application) createCognitoUser(e onboardingEvent, wg *sync.WaitGroup, errChan chan<- onboardingChannel) {
	defer wg.Done()
	input := &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:             aws.String(app.Config.UserPoolID),
		Username:               aws.String(e.ContactEmail),
		DesiredDeliveryMediums: aws.StringSlice([]string{"EMAIL"}),
		ForceAliasCreation:     aws.Bool(true),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(e.ContactEmail),
			},
			{
				Name:  aws.String("custom:tenant_id"),
				Value: aws.String(e.TenantID),
			},
			{
				Name:  aws.String("custom:tenant_name"),
				Value: aws.String(e.TenantName),
			},
		},
	}
	resp, err := app.Config.IDP.AdminCreateUser(input)
	if err != nil {
		res := onboardingChannel{Error: err, Type: "cognito"}
		if awsErr, ok := err.(awserr.Error); ok {
			res.Error = fmt.Errorf("%v", awsErr)
		} else {
			res.Error = fmt.Errorf("%v", err)
		}
		errChan <- res
	} else {
		if *resp.User.Username != "" {
			cognitoUsername = *resp.User.Username
		} else {
			res := onboardingChannel{Error: errors.New("unable to get username"), Type: "cognito"}
			errChan <- res
		}
	}
}

//go:embed assumeRolePolicy.json
var assumeRolePolicy string

func (app application) createIAMRoleAndCognitoGroup(
	tenantID string,
	groupName string,
	roleName string,
	wg *sync.WaitGroup,
	errChan chan<- onboardingChannel,
) {
	defer wg.Done()
	for key, value := range map[string]string{
		"$LAMBDA_AUTH_ROLE_ARN":         app.Config.LambdaAuthRoleARN,
		"$LAMBDA_RELEASES_ROLE_ARN":     app.Config.LambdaReleasesRoleARN,
		"$LAMBDA_REPOSITORIES_ROLE_ARN": app.Config.LambdaRepositoriesRoleArn,
		"$LAMBDA_USERS_ROLE_ARN":        app.Config.LambdaUsersRoleArn,
	} {
		assumeRolePolicy = strings.ReplaceAll(assumeRolePolicy, key, value)
	}

	iamInput := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
		Description:              aws.String(fmt.Sprintf("Lambda execution role for tenant %s", tenantID)),
		RoleName:                 aws.String(roleName),
		Tags: []*iam.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("dev_release_dashboard"),
			},
			{
				Key:   aws.String("managed_by"),
				Value: aws.String("dev_release_dashboard"),
			},
		},
	}

	iamResp, err := app.Config.IAM.CreateRole(iamInput)
	if err != nil {
		res := onboardingChannel{Error: err, Type: "iam"}
		if awsErr, ok := err.(awserr.Error); ok {
			res.Error = fmt.Errorf("%v", awsErr)
		} else {
			res.Error = fmt.Errorf("%v", err)
		}
		errChan <- res
		return
	}

	cognitoInput := &cognitoidentityprovider.CreateGroupInput{
		Description: aws.String("Default group for any tenant."),
		GroupName:   aws.String(groupName),
		RoleArn:     aws.String(*iamResp.Role.Arn),
		UserPoolId:  aws.String(app.Config.UserPoolID),
	}

	_, err = app.Config.IDP.CreateGroup(cognitoInput)
	if err != nil {
		res := onboardingChannel{Error: err, Type: "cognito"}
		if awsErr, ok := err.(awserr.Error); ok {
			res.Error = fmt.Errorf("%v", awsErr)
		} else {
			res.Error = fmt.Errorf("%v", err)
		}
		errChan <- res
	}
}

func (app application) createParameters(tenantID, provider string, wg *sync.WaitGroup, errChan chan<- onboardingChannel) {
	defer wg.Done()
	input := &ssm.PutParameterInput{
		DataType:    aws.String("text"),
		Description: aws.String(fmt.Sprintf("Token for %s access", strings.Title(provider))),
		Name:        aws.String(fmt.Sprintf("/deploy-tower/%s/%s/token", tenantID, provider)),
		Overwrite:   aws.Bool(false),
		Tags: []*ssm.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("dev_release_dashboard"),
			},
			{
				Key:   aws.String("managed_by"),
				Value: aws.String("dev_release_dashboard"),
			},
		},
		Tier:  aws.String("Standard"),
		Type:  aws.String("SecureString"),
		Value: aws.String("42"),
	}
	_, err := app.Config.SSM.PutParameter(input)
	if err != nil {
		res := onboardingChannel{Error: err, Type: "ssm"}
		if awsErr, ok := err.(awserr.Error); ok {
			res.Error = fmt.Errorf("%v", awsErr)
		} else {
			res.Error = fmt.Errorf("%v", err)
		}
		errChan <- res
	}
}

func (app application) addUserToCognitoGroup(groupName string, wg *sync.WaitGroup, errChan chan<- onboardingChannel) {
	defer wg.Done()
	input := &cognitoidentityprovider.AdminAddUserToGroupInput{
		GroupName:  aws.String(groupName),
		UserPoolId: aws.String(app.Config.UserPoolID),
		Username:   aws.String(cognitoUsername),
	}
	_, err := app.Config.IDP.AdminAddUserToGroup(input)
	if err != nil {
		res := onboardingChannel{Error: err, Type: "cognito"}
		if awsErr, ok := err.(awserr.Error); ok {
			res.Error = fmt.Errorf("%v", awsErr)
		} else {
			res.Error = fmt.Errorf("%v", err)
		}
		errChan <- res
	}
}

func (app application) writeUserToDynamoDB(e onboardingEvent, wg *sync.WaitGroup, errChan chan<- onboardingChannel) {
	defer wg.Done()
	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(fmt.Sprintf("org#%s#user", e.TenantID)),
			},
			"SK": {
				S: aws.String(e.ContactEmail),
			},
			"Enabled": {
				BOOL: aws.Bool(true),
			},
			"ID": {
				S: aws.String(cognitoUsername),
			},
		},
		ReturnConsumedCapacity:      aws.String("NONE"),
		ReturnItemCollectionMetrics: aws.String("NONE"),
		ReturnValues:                aws.String("NONE"),
		TableName:                   aws.String(app.Config.TableName),
	}
	_, err := app.Config.DB.PutItem(input)
	if err != nil {
		res := onboardingChannel{Error: err, Type: "dynamodb"}
		if awsErr, ok := err.(awserr.Error); ok {
			res.Error = fmt.Errorf("%v", awsErr)
		} else {
			res.Error = fmt.Errorf("%v", err)
		}
		errChan <- res
	}
}

//go:embed executionRolePolicy.json
var executionRolePolicy string

func (app application) addPermissionsToRole(tenantID string, roleName string, wg *sync.WaitGroup, errChan chan<- onboardingChannel) {
	defer wg.Done()
	for key, value := range map[string]string{
		"$TABLE_ARN":     app.Config.TableARN,
		"$TENANT_ID":     tenantID,
		"$ACCOUNT_ID":    app.Config.AccountID,
		"$REGION":        app.Config.Region,
		"$USER_POOL_ARN": app.Config.UserPoolARN,
	} {
		executionRolePolicy = strings.ReplaceAll(executionRolePolicy, key, value)
	}

	input := &iam.PutRolePolicyInput{
		PolicyDocument: aws.String(executionRolePolicy),
		PolicyName:     aws.String("TenantExecutionPolicy"),
		RoleName:       aws.String(roleName),
	}
	_, err := app.Config.IAM.PutRolePolicy(input)
	if err != nil {
		res := onboardingChannel{Error: err, Type: "iam"}
		if awsErr, ok := err.(awserr.Error); ok {
			res.Error = fmt.Errorf("%v", awsErr)
		} else {
			res.Error = fmt.Errorf("%v", err)
		}
		errChan <- res
	}
}

func (app application) updateOrgStatusToOnboarded(e onboardingEvent) error {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#s": aws.String("Status"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":s": {
				S: aws.String("AwaitingPaymentDetails"),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String("organization"),
			},
			"SK": {
				S: aws.String(e.TenantName),
			},
		},
		TableName:        aws.String(app.Config.TableName),
		UpdateExpression: aws.String("SET #s = :s"),
	}

	_, err := app.Config.DB.UpdateItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(aerr.Error())
		} else {
			log.Error(err.Error())
		}
		return err
	}
	return nil
}

func (app application) onboardingCreateHandler(event events.APIGatewayV2HTTPRequest) (string, int) {
	e := onboardingEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
	}

	e.PK = "organization"
	e.TenantID = uuid.NewString()
	e.TenantName = strings.ReplaceAll(strings.Title(e.TenantName), " ", "")
	e.OnboardingDate = time.Now().String()

	groupName := fmt.Sprintf("UserGroup-%s", e.TenantID)
	roleName := fmt.Sprintf("LambdaExecutionRole-%s", e.TenantID)

	itemInput, err := generatePutItemInputExpression(e)
	if err != nil {
		message := "Something's gone wrong and there's been an issue onboarding your oganization. We will be in touch."
		statusCode := 400
		return message, statusCode
	}

	err = app.writeOrgToDynamoDB(e, itemInput)
	if err != nil {
		message := fmt.Sprintf("Organization %s with contact email %s has already been onboarded.", e.TenantName, e.ContactEmail)
		statusCode := 400
		return message, statusCode
	}

	ssmParameterProviders := []string{"github", "gitlab", "slack"}
	apiCalls := len(ssmParameterProviders) + 2
	errChan := make(chan onboardingChannel, apiCalls)
	wg := new(sync.WaitGroup)
	wg.Add(apiCalls)

	go app.createCognitoUser(e, wg, errChan)
	go app.createIAMRoleAndCognitoGroup(e.TenantID, groupName, roleName, wg, errChan)
	for _, provider := range ssmParameterProviders {
		go app.createParameters(e.TenantID, provider, wg, errChan)
	}

	wg.Wait()
	close(errChan)
	if len(errChan) > 0 {
		slackMessage := fmt.Sprintf("*Onboarding Error*\n```TenantName: %s\nTenantID:   %s\n\n", e.TenantName, e.TenantID)
		for issue := range errChan {
			slackMessage += fmt.Sprintf("- %s %s\n", issue.Type, issue.Error)
		}
		slackMessage += "```"
		slackErr := util.PostToSlack(app.Config.SlackWebhookURL, slackMessage)
		if slackErr != nil {
			log.Error(fmt.Sprintf("unable to send onboarding failure notification for %s ID %s", e.TenantName, e.TenantID))
		}
		message := "Something's gone wrong and there's been an issue onboarding your oganization. We will be in touch."
		statusCode := 400
		return message, statusCode
	}

	apiCalls = 3
	errChan = make(chan onboardingChannel, apiCalls)
	wg.Add(apiCalls)
	go app.addUserToCognitoGroup(groupName, wg, errChan)
	go app.addPermissionsToRole(e.TenantID, roleName, wg, errChan)
	go app.writeUserToDynamoDB(e, wg, errChan)

	wg.Wait()
	close(errChan)
	if len(errChan) > 0 {
		slackMessage := fmt.Sprintf("*Onboarding Error*\n```TenantName: %s\nTenantID:   %s\n\n", e.TenantName, e.TenantID)
		for issue := range errChan {
			slackMessage += fmt.Sprintf("- %s %s\n", issue.Type, issue.Error)
		}
		slackMessage += "```"
		slackErr := util.PostToSlack(app.Config.SlackWebhookURL, slackMessage)
		if slackErr != nil {
			log.Error(fmt.Sprintf("unable to send onboarding failure notification for %s ID %s", e.TenantName, e.TenantID))
		}
		message := "Something's gone wrong and there's been an issue onboarding your oganization. We will be in touch."
		statusCode := 400
		return message, statusCode
	}

	e.Status = "AwaitingPaymentDetails"
	err = app.updateOrgStatusToOnboarded(e)
	if err != nil {
		slackMessage := fmt.Sprintf("*Onboarding Error*\n```TenantName: %s\nTenantID:   %s\n\nUnable to update org status```", e.TenantName, e.TenantID)
		slackErr := util.PostToSlack(app.Config.SlackWebhookURL, slackMessage)
		if slackErr != nil {
			log.Error(fmt.Sprintf("unable to send onboarding failure notification for %s ID %s", e.TenantName, e.TenantID))
		}
		message := fmt.Sprintf("Successfully onboarded organization %s", e.TenantName)
		statusCode := 400
		return message, statusCode
	}

	message := fmt.Sprintf("Successfully onboarded organization %s", e.TenantName)
	statusCode := 200
	return message, statusCode
}
