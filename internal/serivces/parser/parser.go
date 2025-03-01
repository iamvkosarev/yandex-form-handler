package parser

import (
	"encoding/json"
	"errors"
	"forms-handler/internal/models"
)

func ParseRequest(request []byte) (models.Request, error) {
	requestBody := &models.RequestBody{}
	if err := json.Unmarshal(request, requestBody); err != nil {
		return models.Request{}, errors.New("error parsing request body: " + err.Error())
	}

	req := models.Request{}
	if err := json.Unmarshal([]byte(requestBody.Body), &req); err != nil {
		return models.Request{}, errors.New("error parsing body content: " + err.Error())
	}

	return req, nil
}
