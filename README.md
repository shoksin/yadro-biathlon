## System prototype for biathlon competitions

1. Клонируйте репозиторий:  
   ``` 
   git clone https://github.com/shoksin/yadro-biathlon.git
   cd yadro-biathlon
   ```
2. Установите зависимости:  
   ``` go mod download ```
3. Подготовьте файл конфигурации и файл событий
4. Запустите программу:
   с параметрами по умолчанию:
   ```
   go run cmd/main.go
   ```
   или с дополнительными параметрами:
   ```
   go run cmd/main.go -events_file="./config/events" -config_file="./config/config.json" -result_file="resultTable"
   ```
