package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

// slackRequestBody defines the schema for POSTs to Slack webhooks
type slackRequestBody struct {
	Text string `json:"text"`
}

// PostToSlack reads a webhookURL from the provided environment variable, and sends the message
// argument to the channel associated with the webhookURL.
func PostToSlack(webhookURL, message string) error {
	slackBody, _ := json.Marshal(slackRequestBody{Text: message})
	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		log.Printf("[ERROR] unable to form request, %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	clientSlack := &http.Client{Timeout: 10 * time.Second}
	log.Printf("[INFO] sending slack notification...")
	resp, err := clientSlack.Do(req)
	if err != nil {
		log.Printf("[ERROR] unable to send POST request, %v", err)
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		log.Println("[ERROR] non-ok response returned from Slack")
		return errors.New("non-ok response returned from Slack")
	}

	return nil
}
