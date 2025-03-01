package email

import (
	"fmt"
)

type SendFunc func(body interface{}) error

type EmailSender struct {
	sendFunc SendFunc
}

func NewEmailSender(sender SendFunc) *EmailSender {
	return &EmailSender{sendFunc: sender}
}
func (e *EmailSender) Send(messages ...Message) error {
	if len(messages) == 0 {
		return fmt.Errorf("no messages provided")
	}

	requestBody := RequestBody{
		Messages: messages,
	}

	return e.sendFunc(requestBody)
}
