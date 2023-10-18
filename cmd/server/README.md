# cmd/server

В данной директории будет содержаться код Сервера, который скомпилируется в бинарное приложение

Команда для запуска с флагами:

go run -ldflags "-X main.Version=v1.0.1 -X 'main.Date=$(date +'%Y/%m/%d %H:%M:%S')' -X 'main.Commit=$(git log --pretty=format:'%h' -n 1)'" cmd/server/main.go