package main

import (
	"bytes"
	"os"
	"text/template"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"
)

var (
	err error
	log zerolog.Logger
	t   *template.Template
)

func init() {
	log = zerolog.New(os.Stdout).With().Logger()

	t, err = template.ParseFiles("index.tmpl")
	if err != nil {
		log.Error().Err(err).Send()
		os.Exit(1)
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

	log.Info().
		Str("true-client-ip", data.Ip).
		Str("user-agent", request.Headers["user-agent"]).
		Send()

	var buf bytes.Buffer

	err = t.Execute(&buf, data)
	if err != nil {
		log.Error().Err(err).Msg("error executing template")
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       "Internal server error",
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
