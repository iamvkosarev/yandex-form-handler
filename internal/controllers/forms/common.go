package forms

import (
	"fmt"
	"forms-handler/internal/models"
)

const (
	clientEmailQUI = "clientEmail"
	couchEmailQUI  = "couchEmail"
)

type FormHandler interface {
	Handle(HandlerInput) (FormResult, error)
}

type FormHandleFunc func(HandlerInput) (FormResult, error)

func (f FormHandleFunc) Handle(handlerInput HandlerInput) (FormResult, error) {
	return f(handlerInput)
}

type EntryHandler struct {
	handlers map[string]FormHandler
}

func NewEntryHandler() *EntryHandler {
	return &EntryHandler{
		handlers: make(map[string]FormHandler),
	}
}

func (f EntryHandler) AddHandler(testType string, handler FormHandleFunc) {
	f.handlers[testType] = handler
}

func (f EntryHandler) Handle(body models.Request) (models.FormResult, error) {
	op := "forms.FormsHandler"

	if _, ok := f.handlers[body.TestType]; !ok {
		return models.FormResult{}, fmt.Errorf(
			"\"%s\" not supported or not set",
			body.TestType,
		)
	}

	clientEmail, err := findStringByQUI(body, clientEmailQUI)
	if err != nil {
		return models.FormResult{}, fmt.Errorf("%s: %w", op, err)
	}
	couchEmail, err := findStringByQUI(body, couchEmailQUI)
	if err != nil {
		return models.FormResult{}, fmt.Errorf("%s: %w", op, err)
	}

	formResult, err := f.handlers[body.TestType].Handle(HandlerInput{body, couchEmail, clientEmail})
	if err != nil {
		return models.FormResult{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.FormResult{
		ClientResult: models.ResultEmailData{
			Destination: clientEmail,
			BodyText:    formResult.ClientResult.BodyText,
			BodyHTML:    formResult.ClientResult.BodyHTML,
		},
		CouchResult: models.ResultEmailData{
			Destination: couchEmail,
			BodyText:    formResult.CouchResult.BodyText,
			BodyHTML:    formResult.CouchResult.BodyHTML,
		},
	}, nil
}

func findStringByQUI(req models.Request, QUI string) (string, error) {
	for qui, data := range req.Answer.Data {
		if qui != QUI {
			continue
		}
		value, ok := data.Value.(string)
		if !ok {
			return "", fmt.Errorf("invalid value format in qui \"%s\", expacting string", QUI)
		}
		return value, nil
	}
	return "", fmt.Errorf("there is no expected qui \"%s\"", QUI)
}
