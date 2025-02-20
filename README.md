# Dumper

Утилита для создания и загрузки дампов PostgreSQL баз данных.

## Возможности

- Создание дампов баз данных из разных окружений
- Автоматическое управление локальным PostgreSQL в Docker
- Загрузка дампов в локальную базу данных
- Поддержка нескольких окружений
- Режим отладки для детальной информации

## Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/your-username/dumper.git
cd dumper
```

2. Установите зависимости:
```bash
go mod download
```

3. Скопируйте пример конфигурации и отредактируйте его:
```bash
cp config.example.yaml config.yaml
```

## Конфигурация

Отредактируйте `config.yaml` и укажите параметры подключения к вашим базам данных:

```yaml
environments:
  - name: dev
    db_dsn: postgres://user:password@dev-host:5432/database?sslmode=disable
  
  - name: stage
    db_dsn: postgres://user:password@stage-host:5432/database?sslmode=verify-full
```

## Использование

1. Запустите утилиту:
```bash
go run . [--debug]
```

2. Выберите окружение из списка
3. Используйте доступные команды:
   - Создание дампа
   - Загрузка дампа в локальную базу
   - Смена окружения

## Требования

- Go 1.21 или выше
- Docker
- PostgreSQL клиент (для создания дампов) 