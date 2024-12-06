package main

import (
	"bytes"
	"os"
	"regexp"
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
	unwantedPathRegex *regexp.Regexp
	phpFileRegex      *regexp.Regexp
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

	unwantedPathRegex = regexp.MustCompile(`^(/wp-|/admin|/cgi-bin|/\.git|/\.env|//)`)
	phpFileRegex = regexp.MustCompile(`\.php$`)
}

func logRequest(request events.LambdaFunctionURLRequest, message string) {
	log.Info().
		Str("ip", request.Headers["true-client-ip"]).
		Str("path", request.RawPath).
		Str("user-agent", request.Headers["user-agent"]).
		Msg(message)
}

func matchUnwantedPaths(path string) bool {
	return phpFileRegex.MatchString(path) || unwantedPathRegex.MatchString(path)
}

func handler(request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	if matchUnwantedPaths(request.RawPath) {
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
		StatusCode: 200,
		Body:       buf.String(),
	}

	logRequest(request, "ok")

	return response, nil
}

func main() {
	lambda.Start(handler)
}
