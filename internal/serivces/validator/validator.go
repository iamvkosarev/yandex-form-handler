package validator

import (
	"errors"
	"forms-handler/internal/models"
)

func Validate(req models.Request) error {
	if req.TestName == "" {
		return errors.New("test name is required")
	}
	if req.TestType == "" {
		return errors.New("test type is required")
	}
	return nil
}
