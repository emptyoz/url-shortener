# URL Shortener Service

Микросервис для сокращения URL-адресов с использованием Go, PostgreSQL, Redis и Prometheus.

## Архитектура

- **Backend**: Go с использованием gorilla/mux
- **База данных**: PostgreSQL для постоянного хранения
- **Кэширование**: Redis для повышения производительности  
- **Мониторинг**: Prometheus + Grafana для сбора и визуализации метрик
- **Контейнеризация**: Docker Compose

## Функциональность

- Создание коротких ссылок
- Редирект по коротким ссылкам
- Просмотр статистики использования сервиса
- Мониторинг метрик производительности
- Автоматическое кэширование в Redis

## Быстрый старт

### Требования

- Docker
- Docker Compose

### Запуск

```bash
# Клонирование репозитория
git clone <repository-url>
cd url-shortener
```
```bash
# Запуск всех сервисов
make docker-run
```

### Создание короткой ссылки
```bash
# Запрос
curl -X POST -d '{"url":"https://google.com"}' http://localhost:8080/api/shorten

# Ответ
{"short_url":"http://localhost:8080/l74pSP","original_url":"https://google.com"}
```