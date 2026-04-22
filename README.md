# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файл из `migrations/0001_create_tasks.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

## Recurring Tasks

В рамках доработки сервиса была добавлена поддержка повторяющихся задач по расписанию.

### Что было сделано

- **Расширена модель данных:** в таблицу `tasks` добавлены поля `repeat_rule` (строка cron-выражения) и `next_run_at` (время ближайшего запланированного запуска).
- **API для работы с периодичностью:**  
  - При создании задачи (`POST /api/v1/tasks`) можно опционально передать поле `repeat_rule`.  
  - При обновлении (`PUT /api/v1/tasks/{id}`) правило можно изменить или удалить (передав пустую строку).  
  - Ответы теперь содержат `repeat_rule` и `next_run_at` (при наличии).
- **Валидация cron-выражений:** используется библиотека `robfig/cron/v3`. Некорректный синтаксис приводит к ошибке `400 Bad Request`.
- **Автоматический расчёт `next_run_at`:** при задании или изменении правила вычисляется ближайшее будущее время запуска относительно текущего момента (UTC).
- **Миграции БД:** добавлен файл `0002_add_repeat_rule.up.sql` для эволюции схемы без потери данных.
- **Документация OpenAPI:** спецификация (`openapi.json`) актуализирована - схемы `CreateTaskRequest`, `UpdateTaskRequest` и `Task` включают поля периодичности.

### Принятые допущения

- Правила интерпретируются в часовом поясе **UTC**.
- `next_run_at` вычисляется один раз при создании/обновлении; фоновый исполнитель (scheduler) в текущей версии отсутствует.
- Пустая строка в `repeat_rule` при обновлении трактуется как удаление правила.
- Статус задачи по умолчанию - `"new"`.

### Затронутые компоненты

- `internal/domain/task/task.go` - расширение сущности `Task`
- `internal/usecase/task/` - бизнес-логика расчёта и валидации
- `internal/repository/postgres/task.go` - SQL-запросы и сканирование новых полей
- `internal/scheduler/cron.go` - обёртка над `robfig/cron/v3`
- `internal/transport/http/handlers/` - обработка `repeat_rule` в DTO
- `migrations/` - новый файл миграции
- `internal/transport/http/docs/openapi.json` - актуализация схем

## Настройка периодических задач (Cron)

Сервис использует стандартные cron-выражения для определения расписания повторяющихся задач.

### Формат cron-выражения

Выражение состоит из пяти полей, разделённых пробелами:

cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow

