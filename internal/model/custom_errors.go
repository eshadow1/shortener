package model

import "fmt"

// CustomPostgresError представляет кастомную ошибку для обработки
// специфичных ошибок базы данных PostgreSQL.
type CustomPostgresError struct {
	// Message содержит текстовое описание ошибки.
	Message string
	// Err содержит оригинальную ошибку, которая была обернута.
	Err error
}

// Error реализует интерфейс error, возвращая отформатированное строковое представление ошибки.
func (e *CustomPostgresError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}
