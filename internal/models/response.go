package models

type Response struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
}
