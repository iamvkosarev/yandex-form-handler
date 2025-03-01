package email

type RequestBody struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Destination string `json:"destination"`
	Subject     string `json:"subject"`
	BodyHTML    string `json:"html_body,omitempty"`
	BodyText    string `json:"text_body,omitempty"`
}
