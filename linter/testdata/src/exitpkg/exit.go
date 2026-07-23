// Package exitpkg — тестовый пакет для проверки диагностики вызова os.Exit вне main.
package exitpkg

import "os"

// FuncExit вызывает os.Exit(1) — анализатор должен сообщить о нарушении.
func FuncExit() {
	os.Exit(1) // want "call to os.Exit is only allowed in the main function of the main package"
}
