package function

import (
	"encoding/json"
	"log"
	"net/http"

	handler "github.com/openfaas-incubator/go-function-sdk"
)

type jsonInput struct {
	Username,
	Password,
	Token string
	Logout bool
}

type jsonOutput struct {
	Token,
	Error string
	Days []Day
}

// Handle a function invocation
func Handle(req handler.Request) (handler.Response, error) {
	var err error
	var input jsonInput

	if req.Method == "OPTIONS" {
		// allow any origin for this function
		return handler.Response{
			Body:       nil,
			StatusCode: http.StatusOK,
			Header: map[string][]string{
				"Access-Control-Allow-Origin":  {"*"},
				"Access-Control-Allow-Headers": {"Content-Type"},
			},
		}, err
	}

	json.Unmarshal(req.Body, &input)

	status := http.StatusOK
	data, err := handleAux(input)

	if err != nil {
		switch err.(type) {
		case *UnauthorizedError:
			status = http.StatusUnauthorized
		default:
			status = http.StatusBadRequest
		}

		out := jsonOutput{}
		out.Error = err.Error()
		data = &out
		log.Println(err)
	}
	res, err := json.Marshal(data)

	return handler.Response{
		Body:       res,
		StatusCode: status,
		Header: map[string][]string{
			"Access-Control-Allow-Origin":  {"*"},
			"Access-Control-Allow-Headers": {"Content-Type"},
		},
	}, err
}

func handleAux(input jsonInput) (*jsonOutput, error) {
	var err error
	var token string
	var output jsonOutput

	if input.Token == "" {
		token, err = Login(input.Username, input.Password)

		if err != nil {
			return nil, err
		}
		output.Token = token
	}

	if input.Token != "" {
		token = input.Token
	}

	if input.Logout {
		err = Logout(token)
		if err != nil {
			return nil, err
		}
		return &output, nil
	}

	days, err := GetData(token)

	if err != nil {
		return nil, err
	}
	output.Days = days
	return &output, nil
}
