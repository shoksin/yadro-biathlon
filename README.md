## System prototype for biathlon competitions

## Структура проекта:
.
├── cmd
│     └── main.go
├── Dockerfile
├── go.mod
├── internal
      ├── config
            ├── config.json
            ├── events
            ├── loadConfig.go
            └── loadConfig_test.go
      ├── events
            ├── events.go
            └── events_test.go
      ├── messages
            └── messages.go
      ├── models
            ├── competitor.go
            └── event.go
      ├── processor
            ├── processor.go
            └── processor_test.go
      └── utils
            ├── timeUtils.go
            └── timeUtils_test.go
├── README.md


## Сборка и запуск через Docker:

1. Соберите Docker образ:
```bash
   docker build -t biathlon-processor .
```
2. Запустите контейнер с параметрами по умолчанию:
   --config_file=./internal/config/config.json,  
   --events_file=./internal/config/events,  
   --result_file=./results/resultingTable  
```bash
  docker run -v $(pwd)/results:/app/results biathlon-processor
```
3. Или запустите с собственными параметрами:
```bash
  docker run -v $(pwd)/results:/app/results -v $(pwd)/logs:/app/logs biathlon-processor --config_file=./internal/config/config.json --events_file=./internal/config/events --result_file=./results/resultTable --save_logs=./logs/race.log
```
4. Если хотите использовать свои конфигурационные файлы:
```bash
   docker run -v $(pwd)/results:/app/results -v $(pwd)/config:/app/custom-config biathlon-processor --config_file=./custom-config/config.json --events_file=./custom-config/events --result_file=./results/resultTable
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
3. Подготовьте файл конфигурации и файл событий (по умолчанию лежат в ./internal/config)
4. Запустите программу:
   с параметрами по умолчанию:
   ```bash
   go run cmd/main.go
   ```
   или с дополнительными параметрами:
   ```bash
   go run cmd/main.go -events_file="./internal/config/events" -config_file="./internal/config/config.json" -result_file="resultTable"
   ```

