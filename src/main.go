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
	err               error
	log               zerolog.Logger
	tmpl              *template.Template
	forbiddenResponse = events.LambdaFunctionURLResponse{
		StatusCode: 403,
		Body:       "Forbidden",
	}
	serverErrorResponse = events.LambdaFunctionURLResponse{
		StatusCode: 500,
		Body:       "Internal server error",
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

func logRequest(request events.LambdaFunctionURLRequest, message string) {
	log.Info().
		Str("ip", request.Headers["true-client-ip"]).
		Str("country", request.Headers["cloudfront-viewer-country"]).
		Str("city", request.Headers["cloudfront-viewer-city"]).
		Str("user-agent", request.Headers["user-agent"]).
		Msg(message)
}

func handler(request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	if request.RawPath != "/" {
		logRequest(request, "blocked")
		return forbiddenResponse, nil
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
		return serverErrorResponse, err
	}

	response := events.LambdaFunctionURLResponse{
		Body: buf.String(),
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
		StatusCode: 200,
	}

	logRequest(request, "ok")

	return response, nil
}

func main() {
	lambda.Start(handler)
}
