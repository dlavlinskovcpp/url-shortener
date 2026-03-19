# URL Shortener

HTTP-сервис на Go для сокращения ссылок.

Сервис:
- создает короткую ссылку для исходного URL;
- возвращает исходный URL по сокращению;
- поддерживает два хранилища: `memory` и `postgres`;
- запускается через Docker.

## Возможности

- сокращение генерируется случайным образом;
- длина сокращения составляет `10` символов;
- используются символы `a-z`, `A-Z`, `0-9` и `_`;
- для одного и того же исходного URL возвращается одно и то же сокращение;
- сервис умеет работать либо в памяти, либо через PostgreSQL;
- добавлены unit-тесты для usecase и обеих реализаций хранилища.

## Запуск через Docker

### Docker Compose

Запуск сервиса с in-memory хранилищем:

```bash
docker compose up --build app-memory
```
Сервис будет доступен на `http://localhost:8080`.

Запуск сервиса с PostgreSQL:

```bash
docker compose up --build app-postgres db
```
Сервис будет доступен на `http://localhost:8081`.

Остановка контейнеров:

```bash
docker compose down -v
```

## Make-команды

```bash
make run-memory
make run-postgres
make test
make build
make down
make clean
```

## HTTP API

### 1. Создать короткую ссылку

`POST /api/shorten`

Пример запроса:

```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://alpha.dev/some/long/path"}'
```

Пример ответа:

```json
{
  "short_url": "gbB7x9Z_ao"
}
```

Пример ошибки:

```json
{
  "error": "invalid original url"
}
```

### 2. Получить исходную ссылку

`GET /api/original/{short}`

Пример запроса:

```bash
curl http://localhost:8080/api/original/gbB7x9Z_ao
```

Пример ответа:

```json
{
  "original_url": "https://alpha.dev/some/long/path"
}
```

Если ссылка не найдена:

```json
{
  "error": "url not found"
}
```

## Тесты

Запуск всех тестов:

```bash
go test ./...
```

Запуск с покрытием:

```bash
go test ./... -cover
```

Проверка race condition:

```bash
go test -race ./...
```

## Как работает генерация коротких ссылок

- сервис генерирует случайную строку заданной длины;
- строка собирается из допустимого набора символов;
- если сокращение уже занято, выполняется повторная попытка;
- если исходный URL уже был сохранен ранее, возвращается существующее сокращение.
