package forms

import (
	"forms-handler/internal/models"
)

type HandlerInput struct {
	models.Request
	CouchEmail, ClientEmail string
}

type FormResult struct {
	ClientResult, CouchResult PersonalFormResult
}

type PersonalFormResult struct {
	BodyText, BodyHTML string
}
