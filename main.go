package main

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var t *template.Template
var err error

func init() {
	t, err = template.ParseFiles("index.tmpl")
	if err != nil {
		panic(err)
	}
}

func handler(request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	data := struct {
		Ip       string
		Hostname string
		Date     string
	}{
		request.Headers["true-client-ip"],
		request.Headers["x-cdn-host"],
		time.Now().Format("Jan 2, 2006 15:04:05 MST"),
	}

	var buf bytes.Buffer

	err = t.Execute(&buf, data)
	if err != nil {
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("error executing template: %v", err.Error()),
		}, err
	}

	response := events.LambdaFunctionURLResponse{
		Headers: map[string]string{
			"Content-Type":              "text/html",
			"Content-Security-Policy":   "script-src 'self'; frame-ancestors 'none'",
			"Permissions-Policy":        "geolocation=()",
			"Referrer-Policy":           "no-referrer",
			"Strict-Transport-Security": "max-age=63072000; includeSubDomains",
			"X-Content-Type-Options":    "nosniff",
			"X-XSS-Protection":          "1; mode=block",
		},
		StatusCode: 200,
		Body:       buf.String(),
	}

	return response, nil
}

func main() {
	lambda.Start(handler)
}
