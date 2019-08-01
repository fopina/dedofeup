package function

import (
	"encoding/json"
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

	json.Unmarshal(req.Body, &input)

	status := http.StatusOK
	data, err := handleAux(input)

	if err != nil {
		status = http.StatusBadRequest
		out := jsonOutput{}
		out.Error = err.Error()
		data = &out
	}
	res, err := json.Marshal(data)

	return handler.Response{
		Body:       res,
		StatusCode: status,
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
