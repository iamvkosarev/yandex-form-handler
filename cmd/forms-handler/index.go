package main

import (
	"context"
	"fmt"
	"forms-handler/internal/controllers/forms"
	"forms-handler/internal/models"
	"forms-handler/internal/response"
	"forms-handler/internal/serivces/email"
	"forms-handler/internal/serivces/parser"
	"forms-handler/internal/serivces/validator"
	"forms-handler/pkg/yandex/function"
	"net/http"
	"os"
)

const (
	yandexFunctionURLEnv = "YANDEX_FUNCTION_URL"
	yandexApiKeyEnv      = "YANDEX_API_KEY"
)

func Handler(ctx context.Context, request []byte) (*models.Response, error) {
	op := "FormsHandler"
	functionURL, apiKey, err := getFunctionURLAndApiKey()
	if err != nil {
		return response.BadResponse(http.StatusInternalServerError, "internal error"),
			fmt.Errorf("%s: %v\n", op, err)
	}

	req, err := parser.ParseRequest(request)
	if err != nil {
		return response.BadResponse(
			http.StatusBadRequest,
			fmt.Sprintf("parsing error: %v", err),
		), fmt.Errorf("%s: %v\n", op, err)
	}

	if err := validator.Validate(req); err != nil {
		return response.BadResponse(
			http.StatusBadRequest,
			fmt.Sprintf("validation error: %v", err),
		), fmt.Errorf("%s: %v\n", op, err)
	}

	handler := forms.NewEntryHandler()
	handler.AddHandler("testBelov", forms.HandleBelov)
	handler.AddHandler("bpnss", forms.HandleBPNSS)
	handler.AddHandler("reana", forms.HandleReana)
	handler.AddHandler("tsov4", forms.HandleTSOV4)
	handler.AddHandler("gse", forms.HandleGSE)

	testResult, err := handler.Handle(req)
	if err != nil {
		fmt.Printf("%s: %v\n", op, err)
		return response.BadResponse(http.StatusInternalServerError, "internal error"),
			fmt.Errorf("%s: %v\n", op, err)
	}

	functionClient := function.NewYandexFunctionClient(functionURL, apiKey)
	emailSender := email.NewEmailSender(functionClient.InvokeFunction)

	subject := fmt.Sprintf("Результаты тестирования \"%s\"", req.TestName)

	messages := make([]email.Message, 0)
	if testResult.ClientResult.Destination == testResult.CouchResult.Destination {
		messages = append(messages, prepareMessage(testResult.ClientResult, subject))
	} else {
		messages = append(
			messages,
			prepareMessage(testResult.ClientResult, subject),
			prepareMessage(testResult.CouchResult, subject),
		)
	}
	err = emailSender.Send(messages...)

	if err != nil {
		return response.BadResponse(
			http.StatusInternalServerError,
			fmt.Sprintf("yandex function error: %v", err),
		), fmt.Errorf("%s: %v\n", op, err)
	}

	return response.Ok("Тест обработан."), nil
}

func getFunctionURLAndApiKey() (string, string, error) {
	functionURL := os.Getenv(yandexFunctionURLEnv)

	if functionURL == "" {
		return "", "", fmt.Errorf("%s is empty", yandexFunctionURLEnv)
	}

	apiKey := os.Getenv(yandexApiKeyEnv)

	if apiKey == "" {
		return "", "", fmt.Errorf("%s is empty", yandexApiKeyEnv)
	}
	return functionURL, apiKey, nil
}

func prepareMessage(result models.ResultEmailData, subject string) email.Message {
	return email.Message{
		Destination: result.Destination,
		Subject:     subject,
		BodyHTML:    result.BodyHTML,
		BodyText:    result.BodyText,
	}
}
