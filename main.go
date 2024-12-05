package main

import (
	"bytes"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"
)

var (
	err              error
	log              zerolog.Logger
	tmpl             *template.Template
	unwantedPrefixes = []string{
		"/wp-",
		"/admin",
		"/cgi-bin",
		"/.git",
		"/.env",
		"//",
	}
)

func init() {
	log = zerolog.New(os.Stdout).With().Logger()

	tmpl, err = template.ParseFiles("index.tmpl")
	if err != nil {
		log.Error().Err(err).Send()
		os.Exit(1)
	}
}

func matchUnwantedPaths(path string) bool {
	if strings.HasSuffix(path, ".php") {
		return true
	}

	for _, prefix := range unwantedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

func handler(request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	if matchUnwantedPaths(request.RawPath) {
		log.Info().
			Str("true-client-ip", request.Headers["true-client-ip"]).
			Str("user-agent", request.Headers["user-agent"]).
			Msg("blocked")

		return events.LambdaFunctionURLResponse{
			StatusCode: 403,
			Body:       "Forbidden",
		}, nil
	}

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

	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.Error().Err(err).Msg("error")
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

	log.Info().
		Str("true-client-ip", request.Headers["true-client-ip"]).
		Str("user-agent", request.Headers["user-agent"]).
		Msg("ok")

	return response, nil
}

func main() {
	lambda.Start(handler)
}
