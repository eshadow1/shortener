# shortener

## Бенчмарки и оптимизация памяти

### Запуск бенчмарков

```bash
go test -bench=Benchmark -benchmem -memprofile='profiles/result.pprof' -run=^$ .\internal\handler
```

### Профилирование памяти

```bash

go test -bench=Benchmark -benchmem -memprofile=profiles/base.pprof -run=^$ .\internal\handler

go test -bench=Benchmark -benchmem -memprofile=profiles/result.pprof -run=^$ .\internal\handler

go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

### Проведенные действия

В `handler` произведены замены:
   - io.ReadAll + json.Unmarshal на json.NewDecoder().Decode()
   - json.Marshal + w.Write на json.NewEncoder().Encode()
   - w.Write([]byte(str)) на io.WriteString(w, str)
   - strings.TrimSpace(string(b)) на string(bytes.TrimSpace(b))

В  `AuthMiddleware` создается `worker` один раз при инициализации


### Вывод `pprof -diff_base`
```
Type: alloc_space
Time: 2026-06-28 17:32:55 MSK
Showing nodes accounting for -6MB, 5.99% of 100.14MB total
Dropped 6 nodes (cum <= 0.50MB)
      flat  flat%   sum%        cum   cum%
      -3MB  3.00%  3.00%       -4MB  4.00%  github.com/golang-jwt/jwt/v5.(*Token).SignedString
      -3MB  3.00%  5.99%       -3MB  3.00%  context.(*cancelCtx).propagateCancel
    2.50MB  2.50%  3.50%        3MB  3.00%  net/http.readRequest
    2.50MB  2.50%     1%     2.50MB  2.50%  crypto/internal/fips140/sha256.New (inline)
      -2MB  2.00%  3.00%       -2MB  2.00%  github.com/golang-jwt/jwt/v5.NewWithClaims (inline)
       2MB  2.00%     1%     2.50MB  2.50%  context.WithDeadlineCause
      -2MB  2.00%  3.00%     0.50MB   0.5%  crypto/internal/fips140/hmac.New[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }]
    1.50MB  1.50%  1.50%     1.50MB  1.50%  github.com/jackc/pgx/v5.(*Conn).getRows
   -1.50MB  1.50%  3.00%    -1.50MB  1.50%  net/http.(*Request).SetPathValue (inline)
   -1.50MB  1.50%  4.49%       -1MB     1%  github.com/golang-jwt/jwt/v5.(*SigningMethodHMAC).Sign
    1.50MB  1.50%  3.00%        3MB  3.00%  github.com/jackc/pgx/v5/stdlib.(*Conn).QueryContext
   -1.50MB  1.50%  4.49%    -1.50MB  1.50%  net/http/httptest.NewRecorder (inline)
   ...
```


# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

