# Galaxy

Galaxy — микросервисная система для планирования задач и управления проектами.

Проект находится в активной разработке. В текущем состоянии репозиторий содержит базовую инфраструктуру, сервис авторизации, BFF-слой, сервис компаний, сервис логирования, Kafka, Redis, PostgreSQL и Kong API Gateway. Доменные сущности для проектов, задач, комментариев, файлов и уведомлений уже описаны в SQL-инициализации базы данных, но отдельные сервисы для этих областей пока не вынесены в самостоятельные модули.

## Назначение проекта

Galaxy разрабатывается как система для командной работы с проектами и задачами. Целевая модель включает регистрацию пользователей, управление компаниями, проектами, участниками, ролями, правами доступа, задачами, исполнителями, комментариями, файлами, уведомлениями и аудитом событий.

Текущая реализация покрывает базовый контур микросервисной архитектуры:

- авторизация и управление аккаунтами;
- управление компаниями;
- BFF-слой для взаимодействия клиентской части с backend-сервисами;
- централизованная маршрутизация через Kong;
- событийный обмен через Kafka;
- логирование событий в отдельную базу данных;
- хранение временных данных и кэша в Redis;
- основная PostgreSQL-база для бизнес-данных.

## Текущее состояние

Реализовано:

- `auth-service` — сервис авторизации на C#;
- `bff-service` — BFF-сервис на Go;
- `companies-service` — сервис компаний на Go;
- `logging-service` — сервис логирования на Python;
- `kong` — конфигурация API Gateway;
- `docker-compose.yml` — запуск основной инфраструктуры и сервисов;
- `init.sql` — ручная инициализация основной PostgreSQL-базы;
- `init_logs.sql` — ручная инициализация PostgreSQL-базы логирования;
- `architecture.excalidraw` — схема архитектуры.

В разработке / запланировано:

- `front` — клиентская часть на React + TypeScript;
- `projects-service` — сервис проектов;
- `tasks-service` — сервис задач;
- `notification-service` — сервис уведомлений;
- расширение API для полноценной работы с проектами, задачами, комментариями, файлами и уведомлениями.

## Архитектура

Проект построен как набор независимых сервисов, взаимодействующих через HTTP и Kafka.

Основные компоненты:

| Компонент           | Назначение                                                                       |
| ------------------- | -------------------------------------------------------------------------------- |
| `kong`              | API Gateway. Отвечает за маршрутизацию внешних запросов к внутренним сервисам.   |
| `bff-service`       | Backend for Frontend. Агрегация данных.                                          |
| `auth-service`      | Авторизация, регистрация, JWT, refresh token, reset password token, OTP-коды.    |
| `companies-service` | Управление компаниями, ролями, участниками и правами доступа на уровне компании. |
| `logging-service`   | Обработка событий из Kafka и запись логов в отдельную PostgreSQL-базу.           |
| `postgres`          | Основная база данных приложения.                                                 |
| `postgres-logs`     | Отдельная база данных для логов.                                                 |
| `auth-redis`        | Redis для данных авторизации.                                                    |
| `companies-redis`   | Redis для кэша сервиса компаний.                                                 |
| `kafka-service`     | Kafka-брокер для событийного обмена между сервисами.                             |
| `kafka-ui`          | Web-интерфейс для просмотра Kafka-кластерa и топиков.                            |

Схема взаимодействия:

![схема взаимодействия](docs/image.png)

## Технологический стек

| Область                | Технологии             |
| ---------------------- | ---------------------- |
| API Gateway            | Kong Gateway           |
| BFF                    | Go                     |
| Сервис компаний        | Go                     |
| Сервис авторизации     | C# / ASP.NET Core      |
| Сервис логирования     | Python                 |
| Основная база данных   | PostgreSQL             |
| База логов             | PostgreSQL             |
| Кэш / временные данные | Redis                  |
| Брокер сообщений       | Apache Kafka           |
| Kafka UI               | Provectus Kafka UI     |
| Контейнеризация        | Docker, Docker Compose |

## Структура репозитория

```text
.
├── auth-service/            # сервис авторизации
├── bff-service/             # BFF-сервис
├── companies-service/       # сервис компаний
├── kong/                    # конфигурация Kong API Gateway
├── logging-service/         # сервис логирования
├── architecture.excalidraw  # схема архитектуры
├── docker-compose.yml       # запуск приложения и инфраструктуры
├── docker-compose.infra.yml # отдельный compose-файл для CI
├── init.sql                 # SQL-инициализация основной БД
├── init_logs.sql            # SQL-инициализация БД логов
└── README.md
```

## Требования

Для запуска проекта необходимы:

- Docker;
- Docker Compose;
- клиент для подключения к PostgreSQL, например `psql`, DBeaver, DataGrip или pgAdmin.

## Переменные окружения

Перед запуском необходимо создать файл `.env` в корне проекта.

Пример `.env`:

```env
AUTH_SERVICE_URL=http://auth-service:8080
COMPANIES_SERVICE_URL=http://companies-service:8080

POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=postgres

LOGS_POSTGRES_USER=postgres
LOGS_POSTGRES_PASSWORD=postgres
LOGS_POSTGRES_DB=postgres_logs

AUTH_REDIS_PASSWORD=auth_redis_password
COMPANIES_REDIS_PASSWORD=companies_redis_password

JWT_KEY=change_me_to_a_long_secret_key
JWT_ISSUER=galaxy
JWT_AUDIENCE=galaxy
JWT_ACCESS_MINUTES=15
JWT_REFRESH_DAYS=30

PASSWORD_HASHER_MEMORY_SIZE=65536
PASSWORD_HASHER_ITERATIONS=3
PASSWORD_HASHER_DEGREE_OF_PARALLELISM=2
PASSWORD_HASHER_SALT_LENGTH=16
PASSWORD_HASHER_HASH_LENGTH=32

OTP_CODE_ACCESS_MINUTES=10
RESET_PASSWORD_TOKEN_ACCESS_MINUTES=15
INVITATION_LINK_ACCESS_MINUTES=1440

KAFKA_BOOTSTRAP_SERVERS=kafka-service:9092
KAFKA_AUTH_EVENTS_TOPIC=auth-events
KAFKA_COMPANY_EVENTS_TOPIC=company-events
```

Значения в примере предназначены для локального запуска. Для production-окружения необходимо заменить пароли, JWT-ключ и остальные чувствительные параметры.

## Запуск проекта

Склонировать репозиторий:

```bash
git clone https://github.com/nomadxx404/galaxy.git
cd galaxy
```

Создать `.env` в корне проекта и заполнить переменные окружения.

Запустить контейнеры:

```bash
docker compose up --build
```

## Доступные сервисы и порты

После запуска через `docker compose up --build` сервисы доступны на следующих портах:

Доступные маршруты документации через Scalar:

| Маршрут                                  | Назначение                           |
| ---------------------------------------- | ------------------------------------ |
| `http://localhost:8000/auth/scalar`      | Документация API сервиса авторизации |
| `http://localhost:8000/companies/scalar` | Документация API сервиса компаний    |
| `http://localhost:8000/ui/scalar`        | Документация API BFF/UI-слоя         |

Инфраструктурные порты:

| Сервис          |              URL / порт | Назначение                 |
| --------------- | ----------------------: | -------------------------- |
| Kong Gateway    | `http://localhost:8000` | Единая внешняя точка входа |
| Kafka           |        `localhost:9092` | Kafka broker               |
| Kafka UI        | `http://localhost:8085` | Web-интерфейс Kafka        |
| PostgreSQL      |        `localhost:5430` | Основная БД                |
| PostgreSQL Logs |        `localhost:5431` | БД логирования             |
| Redis Auth      |        `localhost:6379` | Redis для авторизации      |
| Redis Companies |        `localhost:6380` | Redis для сервиса компаний |

## База данных

Основная база данных инициализируется файлом `init.sql`.

В текущей схеме создаются следующие PostgreSQL-схемы:

- `auth` — аккаунты, авторизация, outbox-события авторизации;
- `company` — компании, роли, участники, права доступа, outbox-события компаний;
- `project` — проекты, роли, участники, права доступа проектов;
- `task` — задачи, статусы, теги, приоритеты, исполнители, согласование задач;
- `comment` — комментарии и связанные файлы;
- `file` — метаданные файлов;
- `notification` — уведомления.

База логирования инициализируется файлом `init_logs.sql`.

В ней создаётся схема:

- `log` — таблица логов сервисов.

## Событийная модель

В проекте используется Kafka для передачи событий между сервисами.

Сейчас в конфигурации используются топики:

- `auth-events` — события сервиса авторизации;
- `company-events` — события сервиса компаний.

Сервисы могут публиковать события в Kafka, а `logging-service` обрабатывает их и сохраняет данные в базу логирования.

В сервисах также заложен подход с outbox-таблицами. Это позволяет сначала сохранить событие в той же транзакции, что и бизнес-изменение, а затем передать его в Kafka.

## Работа с логами

Сервис `logging-service` получает события из Kafka и сохраняет их в отдельную базу данных `postgres-logs`.

Для просмотра данных можно подключиться к базе логов:

```text
Host: localhost
Port: 5431
Database: значение LOGS_POSTGRES_DB
User: значение LOGS_POSTGRES_USER
Password: значение LOGS_POSTGRES_PASSWORD
```

## Kafka UI

Kafka UI доступен после запуска контейнеров:

```text
http://localhost:8085
```

Через него можно просматривать Kafka-кластер, топики, сообщения и consumer groups.

## Остановка проекта

Остановить контейнеры:

```bash
docker compose down
```

Остановить контейнеры и удалить volumes с данными баз и Redis:

```bash
docker compose down -v
```

Команда с `-v` удалит сохранённые данные PostgreSQL и Redis. Использовать её нужно только если необходимо полностью очистить локальное окружение.

## Текущие ограничения

- Frontend-приложение пока отсутствует в репозитории.
- Сервисы проектов, задач и уведомлений пока не выделены в отдельные модули.
- Часть доменной модели уже описана на уровне базы данных, но ещё не полностью реализована на уровне сервисов.
- Конфигурация предназначена для локальной разработки и требует доработки перед production-использованием.

## Roadmap

Ближайшие планируемые доработки:

- реализовать `projects-service`;
- реализовать `tasks-service`;
- реализовать `notification-service`;
- добавить frontend на React + TypeScript;
- расширить OpenAPI-документацию;
- добавить полноценные сценарии интеграционного тестирования;
- настроить CI/CD;
- подготовить production-конфигурацию окружения.

## Лицензия
