# Task Service

Сервис для управления задачами с HTTP API на Go, включая периодические расписания и фоновую материализацию задач.

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

Причина в том, что SQL-файлы из `migrations/*.up.sql` монтируются в `docker-entrypoint-initdb.d` и применяются только при инициализации пустого data volume.

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
- `POST /api/v1/schedules`
- `GET /api/v1/schedules`
- `GET /api/v1/schedules/{id}`
- `PUT /api/v1/schedules/{id}`
- `DELETE /api/v1/schedules/{id}` (деактивация)

## Поддерживаемые типы периодичности

- `daily` — каждый `n`-й день (`schedule_payload.interval`).
- `monthly_day` — в конкретный день месяца `1..30` (`schedule_payload.day`).
- `specific_dates` — в конкретные даты (`schedule_payload.dates`).
- `odd_even_days` — только четные или только нечетные дни месяца (`schedule_payload.mode = odd|even`).

## Как работает генератор

1. При старте приложения выполняется один цикл генерации (`startup`).
2. Далее генерация запускается периодически по `ticker`.
3. Генератор выбирает активные расписания, пересекающиеся с окном `today..today+N`.
4. Для каждой целевой даты пытается вставить задачу с `schedule_id` и `planned_for`.
5. Повторные циклы безопасны: дубликаты не создаются за счет уникального индекса `(schedule_id, planned_for)` и `ON CONFLICT DO NOTHING`.

## Конфигурация генератора

Все параметры задаются через env:

- `SCHEDULE_GENERATOR_ENABLED` (default: `true`)
- `SCHEDULE_GENERATOR_INTERVAL` (default: `1h`)
- `SCHEDULE_GENERATOR_TIMEOUT` (default: `15s`)
- `SCHEDULE_GENERATOR_WINDOW_DAYS` (default: `30`)

## Бизнес-допущения и политика

- **Timezone-политика:** use case и генератор работают в `UTC`, правила интерпретируются как `date-only`.
- **Короткие месяцы:** для `monthly_day` задача не создается в месяце, где такого дня нет (например, 30-е в феврале).
- **Редактирование расписания:** уже сгенерированные задачи не пересоздаются и не модифицируются; новые циклы учитывают обновленные правила только для будущих дат.
- **Удаление расписания:** `DELETE /schedules/{id}` выполняет деактивацию (`is_active=false`), а не физическое удаление.
- **Обратная совместимость:** `schedule_id` и `planned_for` в ответе задачи опциональны.

## Фильтры списка задач

`GET /api/v1/tasks` поддерживает query-параметры:

- `planned_for_from=YYYY-MM-DD`
- `planned_for_to=YYYY-MM-DD`
- `schedule_id=<int64>`

При невалидных фильтрах возвращается `400`.

## Тестирование

Запуск всех тестов:

```bash
go test ./...
```

Ключевые группы тестов:

- валидация расписаний (`internal/domain/task/schedule_test.go`);
- расчет дат генерации, включая февраль/високосность (`internal/usecase/task/schedule_generator_test.go`);
- идемпотентность генератора при повторных и конкурентных запусках (`internal/usecase/task/schedule_service_generate_test.go`).
