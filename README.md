# Proxy Track Service

Сервис-прокси для отслеживания посылок с батчингом запросов к внешнему API 4px.

## 🚀 Возможности

- **Батчинг запросов**: Группирует запросы пользователей для оптимизации внешних API вызовов
- **Кэширование**: Использует Redis для кэширования результатов отслеживания
- **Web Scraping**: Получает данные о посылках с сайта 4px
- **Docker Support**: Полная контейнеризация с Docker Compose
- **Graceful Shutdown**: Корректное завершение работы сервиса
- **Health Checks**: Мониторинг состояния сервиса

## 📋 Требования

- Docker и Docker Compose
- Go 1.25+ (для локальной разработки)

## 🛠 Установка и запуск

### Быстрый запуск с Docker

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/ShamikRama/proxy_track_service.git
cd proxy_track_service
```

2. **Запустите сервис:**
```bash
docker-compose up -d --build
```

3. **Проверьте статус:**
```bash
docker-compose ps
```

### Локальная разработка

1. **Установите зависимости:**
```bash
go mod download
```

2. **Запустите Redis:**
```bash
docker-compose up -d redis
```

3. **Создайте .env файл:**
```bash
cp .env.example .env
```

4. **Запустите сервис:**
```bash
go run cmd/server/main.go
```

## 🔧 Конфигурация

Создайте файл `.env` в корне проекта:

```env
# Server Configuration
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Batcher Configuration
BATCHER_BATCH_SIZE=50
BATCHER_BATCH_TIMEOUT=2s
BATCHER_WORKERS=2

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# External API Configuration
EXTERNAL_API_URL=https://track.4px.com
EXTERNAL_API_TIMEOUT=30s
EXTERNAL_HASH_PATTERN=4px
```

## 📡 API Endpoints

### Основные endpoints

- `GET /` - Информация о сервисе
- `GET /health` - Проверка состояния сервиса
- `GET /track/{trackCode}` - Отслеживание посылки

### Примеры использования

**Проверка состояния сервиса:**
```bash
curl http://localhost:8080/health
```

**Отслеживание посылки:**
```bash
curl http://localhost:8080/track/TEST123
```

**Ответ сервиса:**
```json
{
  "status": true,
  "data": {
    "countries": ["TEST123 - In Transit"],
    "events": [
      {
        "status": "Package received",
        "date": "2025-09-18T03:53:58Z"
      }
    ]
  }
}
```

## 🧪 Тестирование

### Запуск тестов

```bash
# Все тесты
go test ./...

# Тесты батчера
go test ./internal/batcher/ -v

# Тесты с покрытием
go test ./... -cover
```

### Тесты батчинга

Сервис включает тесты для проверки логики батчинга:

- **TestBatcherBatchSize**: Проверяет накопление запросов до размера батча
- **TestBatcherTimeout**: Проверяет отправку батча по таймауту
- **TestBatcherMultipleBatches**: Проверяет обработку нескольких батчей
- **TestBatcherConcurrency**: Проверяет конкурентную обработку

## 🏗 Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │───▶│  Track Handler  │───▶│ Tracking Service│
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                                                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Redis Cache   │◀───│     Batcher     │───▶│  External API   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Компоненты

- **Track Handler**: HTTP обработчик запросов
- **Tracking Service**: Основная бизнес-логика
- **Batcher**: Группирует запросы по условиям:
  - Накопилось 50 трек-кодов
  - Прошло более 2 секунд
- **Cache**: Redis для кэширования результатов
- **External Client**: Web scraping клиент для 4px

## 📊 Мониторинг

### Health Check

```bash
curl http://localhost:8080/health
```

**Успешный ответ:**
```json
{
  "status": true,
  "message": "service is healthy",
  "service": "proxy_track_service"
}
```

### Логи

```bash
# Логи всех сервисов
docker-compose logs

# Логи только tracking-service
docker-compose logs tracking-service

# Логи в реальном времени
docker-compose logs -f tracking-service
```

## 🐳 Docker

### Команды Docker Compose

```bash
# Запуск
docker-compose up -d

# Пересборка и запуск
docker-compose up -d --build

# Остановка
docker-compose down

# Просмотр логов
docker-compose logs -f

# Статус сервисов
docker-compose ps
```

### Структура контейнеров

- **tracking-service**: Основной сервис (порт 8080)
- **tracking-redis**: Redis для кэширования (порт 6379)

## 🔧 Разработка

### Структура проекта

```
├── cmd/
│   └── server/           # Точка входа приложения
├── internal/
│   ├── batcher/          # Логика батчинга
│   ├── client/           # Клиенты внешних API
│   ├── config/           # Конфигурация
│   ├── handler/          # HTTP обработчики
│   ├── repository/       # Репозитории данных
│   └── service/          # Бизнес-логика
├── pkg/
│   └── models/           # Модели данных
├── docker-compose.yml    # Docker Compose конфигурация
├── Dockerfile           # Docker образ
└── README.md            # Документация
```

### Добавление новых тестов

1. Создайте тестовый файл: `*_test.go`
2. Используйте mock объекты для изоляции тестов
3. Проверяйте как успешные, так и ошибочные сценарии

### Линтинг и форматирование

```bash
# Форматирование кода
go fmt ./...

# Линтинг
golangci-lint run

# Проверка импортов
goimports -w .
```

## 🚨 Устранение неполадок

### Сервис не запускается

1. Проверьте логи:
```bash
docker-compose logs tracking-service
```

2. Проверьте порты:
```bash
netstat -an | grep 8080
```

3. Проверьте конфигурацию:
```bash
docker-compose config
```

### Ошибки Redis

1. Проверьте статус Redis:
```bash
docker-compose ps redis
```

2. Проверьте подключение:
```bash
docker exec tracking-redis redis-cli ping
```

### Ошибки внешнего API

1. Проверьте доступность сайта:
```bash
curl -I https://track.4px.com
```

2. Проверьте логи scraping:
```bash
docker-compose logs tracking-service | grep "scraping"
```

## 📝 Лицензия

MIT License

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature branch
3. Добавьте тесты для новой функциональности
4. Убедитесь, что все тесты проходят
5. Создайте Pull Request

## 📞 Поддержка

При возникновении проблем:

1. Проверьте раздел "Устранение неполадок"
2. Создайте Issue в GitHub
3. Приложите логи и описание проблемы