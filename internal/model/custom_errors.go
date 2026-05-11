package model

import "fmt"

type CustomPostgresError struct {
	Message string
	Err     error
}

func (e *CustomPostgresError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}
