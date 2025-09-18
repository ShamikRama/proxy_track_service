# Proxy Track Service

## Требования
- Docker и Docker Compose
- Go 1.25+ (для локальной разработки)


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
curl http://localhost:8080/track/LK517880262CN

```
**Ответ сервиса:**
```json
{
  "status": true,
  "data": {
    "countries": ["CH - RU"],
    "events": [
      {
        "status": "Domestic Air Cargo Termina / Depart from facility to service provider.",
        "date": "2025-09-18T03:53:58Z"
      },
      {
        "status": "SYSTEM / Shipment arrived at facility and measured.",
        "date": "2025-09-18T03:53:58Z"
      },
      {
        "status": "Parcel information received",
        "date": "2025-09-18T03:53:58Z"
      }
    ]
  }
}
```

## 🧪 Тестирование

### Запуск тестов

```bash
go test ./internal/test/... -v 
