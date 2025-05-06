## System prototype for biathlon competitions

## Запуск с помощью Docker:

```bash
   docker run ...
```

## Запуск с клонирование проекта на локальную машину:
1. Клонируйте репозиторий:  
   ```bash
   git clone https://github.com/shoksin/yadro-biathlon.git
   cd yadro-biathlon
   ```
2. Установите зависимости:  
   ```bash
   go mod download 
   ```
3. Подготовьте файл конфигурации и файл событий
4. Запустите программу:
   с параметрами по умолчанию:
   ```bash
   go run cmd/main.go
   ```
   или с дополнительными параметрами:
   ```bash
   go run cmd/main.go -events_file="./internal/config/events" -config_file="./internal/config/config.json" -result_file="resultTable"
   ```
