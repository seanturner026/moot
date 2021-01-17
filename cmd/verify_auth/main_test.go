package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestHandler(t *testing.T) {
	t.Run("Sucessfully authorized user", func(t *testing.T) {
		expectedBody, err := json.Marshal(map[string]string{"message": "Authorized"})
		if err != nil {
			t.Fatal("json marshal error")
		}

		var buf bytes.Buffer
		json.HTMLEscape(&buf, expectedBody)

		resp, err := handler()
		fmt.Println(resp.Body)
		if err != nil || resp.Body != buf.String() {
			t.Fatal("User should have been authorized")
		}
	})
}
