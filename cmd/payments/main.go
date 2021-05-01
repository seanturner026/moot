package main

// import (
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"strings"

// 	"github.com/aws/aws-lambda-go/events"
// 	"github.com/aws/aws-lambda-go/lambda"
// 	log "github.com/sirupsen/logrus"
// 	"github.com/stripe/stripe-go/v72"
// 	"github.com/stripe/stripe-go/v72/account"
// )

// type application struct {
// 	Config configuration
// 	Stripe stripeAccountAPI
// }

// type configuration struct {
// 	StripeKey string
// }

// type stripeAccountAPI interface {
// 	New(params *stripe.AccountParams) (*stripe.Account, error)
// }

// type paymentOnboardingEvent struct {
// 	TenantName   string
// 	ContactEmail string
// }

// func (app application) handler(event events.SQSEvent) error {
// 	e := paymentOnboardingEvent{}
// 	err := json.Unmarshal([]byte(event.Records[0].Body), &e)
// 	if err != nil {
// 		log.Error(fmt.Sprintf("%v", err))
// 	}

// 	e.TenantName = strings.ReplaceAll(strings.Title(e.TenantName), " ", "")

// 	stripe.Key = app.Config.StripeKey

// 	params := &stripe.AccountParams{
// 		Capabilities: &stripe.AccountCapabilitiesParams{
// 			CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
// 				Requested: stripe.Bool(true),
// 			},
// 			Transfers: &stripe.AccountCapabilitiesTransfersParams{
// 				Requested: stripe.Bool(true),
// 			},
// 		},
// 		Country: stripe.String("US"),
// 		Email:   stripe.String("jenny.rosen@example.com"),
// 		Type:    stripe.String("custom"),
// 	}
// 	a, _ := account.New(params)

// 	// customer, err := customer.New(params)

// 	// if event.RawPath == "/onboarding/create" {
// 	// 	log.Info(fmt.Sprintf("handling request on %s", event.RawPath))
// 	// 	message, statusCode := app.onboardingCreateHandler(event)
// 	// 	return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

// 	// } else {
// 	// 	log.Error(fmt.Sprintf("path %s does not exist", event.RawPath))
// 	// }
// 	return errors.New("unable t")
// }

// func main() {
// 	log.SetFormatter(&log.JSONFormatter{})

// 	config := configuration{
// 		StripeKey: "",
// 	}

// 	app := application{
// 		Config: config,
// 	}

// 	lambda.Start(app.handler)
// }
