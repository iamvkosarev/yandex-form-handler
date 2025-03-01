package response

import (
	"forms-handler/internal/models"
)

func BadResponse(code int, message string) *models.Response {
	return &models.Response{
		StatusCode: code,
		Body:       message,
	}
}

func Ok(message string) *models.Response {
	return &models.Response{
		StatusCode: 200,
		Body:       message,
	}
}
