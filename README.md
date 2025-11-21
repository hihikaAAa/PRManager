# PR Reviewer Assignment Service

Микросервис для автоматического назначения ревьюеров на Pull Request'ы и управления командами/пользователями. Задание для стажировки Avito Backend (осенняя волна 2025).  

---

## Содержание

- [Описание](#описание)
- [Архитектура](#архитектура)
- [API](#api)
  - [Teams](#teams)
  - [Users](#users)
  - [PullRequests](#pullrequests)
  - [Health](#health)
- [Запуск](#запуск)
  - [Через docker-compose](#через-docker-compose)
  - [Локальный запуск](#локальный-запуск)
- [Конфигурация](#конфигурация)
- [Доменные правила](#доменные-правила)
- [Тесты](#тесты)
- [Вопросы и допущения](#вопросы-и-допущения)
- [Доп. задания](#доп.задания)

---

## Описание

Сервис:

- создаёт команды и пользователей;
- автоматически назначает до двух ревьюеров на PR из команды автора;
- позволяет переназначать ревьюеров;
- возвращает список PR'ов, где пользователь назначен ревьювером;
- управляет флагом активности пользователя;
- после `MERGED` запрещает менять ревьюверов;
- операция `merge` идемпотентна.

Взаимодействие только по HTTP, контракт описан в `openapi.yaml`.

---

## Архитектура

Проект организован по слоям:

- `cmd/pr-reviewer-service/main.go` — точка входа, роутинг, запуск сервера
- `config` - конфигурация сервера + dsn для БД
- `internal/domain`
  - `user.User`
  - `team.Team`
  - `pull-request.PullRequest`
  - `pull-request.PullRequestShort`
- `internal/repository/postgres`
  - `UserRepository`
  - `TeamRepository`
  - `PRRepository`
  - `repo_-_errors/repo_errors`
- `internal/services`
  - `prservice.PRService`
  - `teamservice.TeamService`
  - `userservice.UserService`
  - `serviceErrors.serverErrors`
- `internal/http-server/handlers`
  - `/team/add`, `/team/get`
  - `/users/setIsActive`, `/users/getReview`
  - `/pullRequest/create`, `/pullRequest/merge`, `/pullRequest/reassign`
- `internal/lib/logger` — логгер на базе `slog` + pretty handler
- `internal/storage` — создание `*sql.DB`

БД — PostgreSQL.

---

## API

Контракт полностью описан и соответствует `openapi.yaml`. Далее идут примеры curl запросов, тестируемые в Git Bash.

### Teams

#### Создание команды `/team/add`

```bash 
    curl -X POST http://localhost:8080/team/add \
    -H "Content-Type: application/json" \
    -d '{
    "team_name": "backend",
    "members": [
        { "user_id": "u1", "username": "Alice", "is_active": true },
        { "user_id": "u2", "username": "Bob", "is_active": true },
        { "user_id": "u3", "username": "Sergei", "is_active": true },
        { "user_id": "u4", "username": "Clare", "is_active": true },
        { "user_id": "u5", "username": "Max", "is_active": true },
        { "user_id": "u6", "username": "Nikita", "is_active": true }
    ]
    }' 
```

#### Ответ:

```bash
    {
        "team": {
        "team_name": "backend",
        "members": [
      { "user_id": "u1", "username": "Alice", "is_active": true },
      { "user_id": "u2", "username": "Bob", "is_active": true }
        ]
        }
    }
```

#### Получение команды `/team/get?team_name=backend`

```bash
    curl -X GET "http://localhost:8080/team/get?team_name=backend"
```

#### Ответ:

```bash
    {
    "team_name": "backend",
    "members": [
        { "user_id": "u1", "username": "Alice", "is_active": true },
        { "user_id": "u2", "username": "Bob", "is_active": true }
    ]
    }
```
---

### Users

#### Установка активности пользователя `/users/setIsActive`

```bash
    curl -X POST http://localhost:8080/users/setIsActive \
    -H "Content-Type: application/json" \
    -d '{
    "user_id": "u2",
    "is_active": false
    }'
```

#### Ответ:

```bash
    {
    "user": {
    "user_id": "u2",
    "username": "Bob",
    "team_name": "backend",
    "is_active": false
    }
    }
```
---

### PullRequests

#### Получение PR, где пользователь — ревьювер /users/getReview

```bash
curl -X GET "http://localhost:8080/users/getReview?user_id=u2"
```

#### Ответ:

```bash
    {
    "user_id": "u2",
    "pull_requests": [
    {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search",
    "author_id": "u1",
    "status": "OPEN"
    }
    ]
    }
```

#### Создание PR /pullRequest/create

```bash
    curl -i -X POST "http://localhost:8080/pullRequest/create" \
    -H "Content-Type: application/json" \
    --data-raw '{"pull_request_id":"pr-1001","pull_request_name":"Add search","author_id":"u1"}'
```

#### Ответ:

```bash
    {
    "pr": {
        "pull_request_id": "pr-1001",
        "pull_request_name": "Add search",
        "author_id": "u1",
        "status": "OPEN",
        "assigned_reviewers": ["u2", "u3"],
        "createdAt": "2025-01-01T10:00:00Z",
        "mergedAt": null
    }
    }
```

#### Merge PR (идемпотентно) /pullRequest/merge

```bash
    curl -X POST http://localhost:8080/pullRequest/merge \
    -H "Content-Type: application/json" \
    -d '{
    "pull_request_id": "pr-1001"
    }'
```

#### Ответ:

``` bash
    {
    "pr": {
        "pull_request_id": "pr-1001",
        "pull_request_name": "Add search",
        "author_id": "u1",
        "status": "MERGED",
        "assigned_reviewers": ["u2", "u3"],
        "mergedAt": "2025-10-24T12:34:56Z"
        }
        }
```

#### Переназначение ревьювера /pullRequest/reassign

```bash
    curl -X POST http://localhost:8080/pullRequest/reassign \
    -H "Content-Type: application/json" \
    -d '{
    "pull_request_id": "pr-1001",
    "old_user_id": "u6"
    }'
```

#### Ответ:

```bash
    {
    "pr": {
        "pull_request_id": "pr-1001",
        "pull_request_name": "Add search",
        "author_id": "u1",
        "status": "OPEN",
        "assigned_reviewers": ["u3", "u5"]
    },
    "replaced_by": "u5"
    }
```
---

### Health 

#### Проверка работоспособности /health

```bash
curl http://localhost:8080/health
```

#### Ответ:

```bash
{"status":"ok"}
```

---

## Запуск

### Через docker-compose
1. Собрать и поднять сервис:

    ```bash
    docker-compose up --build
    ```

2. Сервис будет доступен на:

    ```bash
    http://localhost:8080
    ```

3. PostgreSQL:

    ```bash
    host=localhost, port=5432, user=postgres, password=postgres, dbname=prmanager
    ```

Миграции автоматически применяются контейнером migrator из папки migrations/

### Локальный запуск

Требования:
- Go 1.22+
- PostgreSQL локально
- config , адаптированный под local

---

## Конфигурация 

Загрузка конфига: 
- CONFIG_PATH (env) — путь к YAML-файлу
- Структура — internal/config.Config

Поля структуры:
- env — "local" | "dev" | "prod" — влияет на формат и уровень логов
- http_server.address — адрес HTTP-сервера (по умолчанию: 8080)
- http_server.read_timeout, write_timeout, idle_timeout - необходимые таймауты
- db.dsn — строка подключения к PostgreSQL

---

## Доменные правила

- User.is_active = false — пользователь никогда не назначается ревьювером
- При создании PR:
  - ищутся активные пользователи из команды автора, кроме самого автора;
  - выбираются до двух случайных ревьюверов;
  - если доступен только один — назначается один;
  - если нет ни одного — список ревьюверов пустой.
- При reassign:
  - сначала проверяется, что PR не MERGED;
  - проверяется, что old_user_id действительно один из ревьюверов;
  - ищутся активные пользователи команды старого ревьювера, исключая автора и всех текущих ревьюверов;
  - случайно выбирается один кандидат;
  - если кандидатов нет — ошибка NO_CANDIDATE.
- merge: 
  - идемпотентен: повторный вызов возвращает актуальное состояние PR;
  - внутри репозитория используется UPDATE ... WHERE status = 'OPEN' + SELECT, чтобы корректно обрабатывать повтор.

---

## Тесты

Юнит-тесты домена/сервисов/репозиториев можно запускать командой:

```bash
make test
```

--- 

## Вопросы и допущения

- В качестве источника случайности при выборе ревьюверов используется math/rand.
- Уникальность pull_request_id и user_id обеспечивается на уровне БД (PRIMARY KEY по текстовому идентификатору).
- Сервис не выполняет дополнительной валидации форматов user_id/pull_request_id/team_name, кроме обязательности полей.
- Большое ограничение: управление пользователями и составом команд.
  - В текущей версии сервиса нет отдельных методов для создания/изменения пользователей и обновления состава команды. Пользователи создаются и обновляются только через эндпоинт: `POST /team/add`. После успешного создания команды её состав нельзя изменить через публичное HTTP-API. Это ограничение следует из изначальной OpenAPI-спецификации: в ней отсутствуют методы для явного создания/редактирования пользователей и изменения состава команды, поэтому сервис реализует только описанные там операции.

--- 

## Доп.Задания 

### 1. Добавить простой эндпоинт статистики (например, количество назначений по пользователям и/или по PR)

#### Описание эндпоинта

- total_pr — общее количество PR в системе
- open_pr — количество PR в статусе OPEN
- merged_pr — количество PR в статусе MERGED
- reviewers — массив объектов вида { user_id, count }, где
  - user_id — идентификатор пользователя;
  - count — сколько раз этот пользователь был назначен ревьювером (во всех PR).

#### Пример запроса

```bash
curl -X GET "http://localhost:8080/stats"
```

#### Ответ:

```bash
{
  "status": "OK",
  "data": {
    "total_pr": 5,
    "open_pr": 3,
    "merged_pr": 2,
    "reviewers": [
      { "user_id": "u1", "count": 3 },
      { "user_id": "u2", "count": 2 }
    ]
  }
}
```




