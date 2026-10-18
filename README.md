# Task Manager

Учебный HTTP API для управления задачами на Go: создание, получение, обновление и удаление. Приложение хранит данные в PostgreSQL; HTTP-сервер запускается локально на `localhost:8080`, база — через Docker Compose.

## Стек и архитектура

Go (`net/http`, `httptest`), PostgreSQL 17, `pgx/v5` и пул соединений `pgxpool`, Docker Compose.

Путь запроса: `HTTP → Handler → Service → Repository → PostgreSQL`.

- **Handler** разбирает HTTP-запросы, проверяет JSON и формирует ответы.
- **Service** проверяет и нормализует имя задачи, проверяет параметры выборки.
- **Repository** выполняет SQL-запросы через `pgxpool`. В `cmd/app` подключён PostgreSQL-репозиторий; реализация в памяти также сохранена в проекте.

Middleware создаёт идентификатор запроса, возвращает его в `X-Request-ID` и записывает сведения о запросе в лог. Ошибки обработчиков имеют единый JSON-формат с `error.code`, `error.message` и `request_id`.

Таймауты HTTP-сервера: чтение заголовков — 5 с, чтение запроса — 10 с, запись ответа — 15 с, ожидание следующего запроса — 60 с. При `SIGINT` или `SIGTERM` выполняется graceful shutdown с ограничением 10 с.

## Требования

- Go 1.26.1 или совместимая более новая версия — см. [go.mod](go.mod).
- Docker с работающим Docker Engine и Docker Compose v2.
- Git для клонирования и `curl` для проверки API.
- Свободные локальные порты `15432` и `8080`.

Устанавливать PostgreSQL и `psql` на компьютер не требуется: команды ниже используют утилиты внутри контейнера.

## Локальный запуск

Примеры предназначены для Unix/macOS (`sh`, `bash`, `zsh`). Отличия для PowerShell приведены ниже. Все команды после клонирования выполняются из корня репозитория.

### 1. Получить проект

```sh
git clone https://github.com/Snoksik/task-manager.git
cd task-manager
```

Если проект уже скачан, перейдите в его каталог.

### 2. Запустить PostgreSQL

```sh
docker compose up -d postgres
docker compose exec -T postgres pg_isready -U task_user -d task_manager
```

`-d` запускает контейнер в фоне, `-T` отключает выделение терминала. При первом запуске дождитесь ответа `accepting connections`; если база ещё запускается, повторите проверку.

[compose.yaml](compose.yaml) использует образ `postgres:17`, базу `task_manager`, пользователя `task_user` и пароль `task_password`. Порт `15432` компьютера направлен на порт `5432` контейнера. Данные хранятся в volume `postgres_data` и сохраняются при обычном перезапуске контейнера.

**Dev-only:** эти учётные данные и все строки подключения ниже — примеры исключительно для локальной разработки. `sslmode=disable` отключает TLS. Для развёртывания вне локальной среды нужны отдельные учётные данные и настройки доступа; реальные секреты нельзя добавлять в Git.

### 3. Однократно применить миграцию

Для новой базы выполните:

```sh
docker compose cp migrations/000001_create_tasks.up.sql postgres:/tmp/task-manager-schema.sql
docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U task_user -d task_manager -f /tmp/task-manager-schema.sql
```

Первая команда копирует SQL-файл в контейнер, вторая выполняет его. `-U` задаёт пользователя, `-d` — базу, `-f` — файл; `ON_ERROR_STOP=1` останавливает выполнение при SQL-ошибке.

Миграция создаёт таблицу `tasks` и не рассчитана на повторное применение: в уже подготовленной базе этот шаг пропустите. Приложение не применяет миграции автоматически.

### 4. Настроить подключение и запустить API

```sh
export DATABASE_URL='postgres://task_user:task_password@localhost:15432/task_manager?sslmode=disable'
go run ./cmd/app
```

`DATABASE_URL` обязательна; перед запуском HTTP-сервера приложение проверяет соединение с БД. Переменная задаётся в текущем терминале. Простого создания `.env` недостаточно: `main.go` читает окружение процесса.

API доступен по адресу `http://localhost:8080`. Для остановки нажмите `Ctrl+C`. Чтобы отдельно остановить PostgreSQL:

```sh
docker compose stop postgres
```

### 5. Проверить API

В другом терминале:

```sh
curl -i http://localhost:8080/tasks

curl -i -X POST http://localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"name":"Изучить PostgreSQL","user_id":1}'
```

`-i` показывает заголовки ответа, `-X` задаёт метод, `-H` — заголовок запроса, `-d` — тело. Для обращения к созданной задаче используйте `id` из ответа POST, например `/tasks/1`, если сервер вернул `id: 1`.

## HTTP API

| Метод | Путь | Назначение | Успешный статус |
|---|---|---|---|
| GET | `/tasks` | Список задач с параметрами выборки | `200 OK` |
| GET | `/tasks/{id}` | Получение задачи по ID | `200 OK` |
| POST | `/tasks` | Создание задачи | `201 Created` |
| PUT | `/tasks/{id}` | Обновление имени и `user_id` задачи | `200 OK` |
| DELETE | `/tasks/{id}` | Удаление задачи, ответ без тела | `204 No Content` |

`GET /tasks` поддерживает `search` (поиск по имени через SQL `LIKE`), `limit` (от 1 до 100, по умолчанию 10) и `offset` (от 0, по умолчанию 0). Задачи возвращаются по возрастанию ID:

```sh
curl -i 'http://localhost:8080/tasks?limit=10&offset=0'
```

POST и PUT принимают JSON с полями `name` и `user_id` и требуют `Content-Type: application/json`. Имя обрезается по краям и не может быть пустым; `user_id` обязателен, но существование пользователя не проверяется. Неизвестные поля, несколько JSON-значений и тело больше 4096 байт отклоняются.

Описание API находится в [api/openapi.yaml](api/openapi.yaml). Его полноту и соответствие реализации следует проверять отдельно; README не заменяет спецификацию.

## Тестирование

Обычный запуск и проверка с детектором гонок:

```sh
go test ./...
go test -race ./...
```

В проекте есть тесты обработчиков, middleware, сервиса, репозиториев и worker pool. Интеграционный `TestTaskCRUDIntegration` проводит CRUD-цикл через `httptest` и настоящий PostgreSQL-репозиторий; отдельно запускать HTTP-сервер для него не нужно.

**Без `TEST_DATABASE_URL` интеграционный тест пропускается.** Успех обычного `go test ./...` сам по себе не подтверждает проверку PostgreSQL. Детектор гонок требует поддерживаемой Go платформы и доступного C-компилятора.

### Подготовить отдельную тестовую БД

PostgreSQL должен быть запущен. Следующие команды создания базы и таблицы выполняются один раз для новой тестовой БД:

```sh
docker compose exec -T postgres createdb -U task_user task_manager_test
docker compose cp migrations/000001_create_tasks.up.sql postgres:/tmp/task-manager-schema.sql
docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U task_user -d task_manager_test -f /tmp/task-manager-schema.sql
```

Если база и таблица уже подготовлены, повторять эти команды не нужно. Тест не создаёт схему автоматически и требует имя базы именно `task_manager_test`.

Задайте подключение в том терминале, где запускаете тесты:

```sh
export TEST_DATABASE_URL='postgres://task_user:task_password@localhost:15432/task_manager_test?sslmode=disable'
go test -v -count=1 ./...
go test -race -count=1 ./...
```

`-count=1` отключает использование кешированных результатов, `-v` показывает результаты отдельных тестов, включая пропуски. Для запуска только интеграционного CRUD-теста:

```sh
go test -v -count=1 -run '^TestTaskCRUDIntegration$' ./internal/Handler
```

### Отличия для PowerShell

Команды Docker, миграции через `docker compose cp`, `go run` и `go test` остаются теми же. Вместо `export` задавайте переменные так:

```powershell
$env:DATABASE_URL = 'postgres://task_user:task_password@localhost:15432/task_manager?sslmode=disable'
$env:TEST_DATABASE_URL = 'postgres://task_user:task_password@localhost:15432/task_manager_test?sslmode=disable'
```

Для простой проверки в Windows используйте `curl.exe -i http://localhost:8080/tasks`, чтобы не вызвать псевдоним PowerShell. Многострочные примеры с `\` выше предназначены для Unix-оболочки.

## Структура проекта

```text
api/openapi.yaml        — спецификация API
cmd/app/                — подключение БД и запуск HTTP-сервера
cmd/dbcheck/            — вспомогательная утилита проверки БД
internal/Handler/       — HTTP-обработчики, middleware и HTTP-тесты
internal/service/       — правила обработки задач
internal/Repository/    — PostgreSQL-репозиторий и реализация в памяти
internal/model/         — модели и ошибки
internal/workerpool/    — отдельный учебный worker pool
migrations/             — SQL-миграции up/down
sql/                    — учебные SQL-запросы и результаты анализа
compose.yaml            — локальный PostgreSQL
```

Регистр каталогов `Handler` и `Repository` соответствует репозиторию. SQL-упражнения из `sql/` не заменяют миграцию и не применяются при запуске приложения.
