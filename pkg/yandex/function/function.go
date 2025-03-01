package function

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/http"
)

// YandexFunctionClient — клиент для вызова Яндекс Функции через API-ключ
type YandexFunctionClient struct {
	FunctionURL string
	APIKey      string
}

func NewYandexFunctionClient(functionURL, apiKey string) *YandexFunctionClient {
	return &YandexFunctionClient{
		FunctionURL: functionURL,
		APIKey:      apiKey,
	}
}

// InvokeFunction отправляет запрос на Яндекс Функцию
func (c *YandexFunctionClient) InvokeFunction(body interface{}) error {
	client := resty.New()

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Api-Key "+c.APIKey).
		SetBody(body).
		Post(c.FunctionURL)

	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	return nil
}
