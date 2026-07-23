// Package panicpkg — тестовый пакет для проверки диагностики вызова panic.
package panicpkg

// FuncPanic вызывает panic — анализатор должен сообщить о нарушении.
func FuncPanic() {
	panic("panic") // want "use of built-in panic is forbidden"
}
