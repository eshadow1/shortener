// Package fatalpkg — тестовый пакет для проверки диагностики log.Fatal вне main.
package fatalpkg

import "log"

// FuncFatal вызывает log.Fatal — анализатор должен сообщить о нарушении.
func FuncFatal() {
	log.Fatal("err") // want "call to log.Fatal is only allowed in the main function of the main package"
}
