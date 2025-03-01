package models

type RequestBody struct {
	Body string `json:"body"`
}

type Request struct {
	TestName string `json:"testName"`
	TestType string `json:"testType"`
	Answer   Answer `json:"answer"`
}

type Answer struct {
	Data map[string]Data `json:"data"`
}

type Data struct {
	Value    interface{} `json:"value"`
	Question Question    `json:"question"`
}

type Question struct {
	Slug string `json:"slug"`
}
