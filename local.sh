templ generate -v

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o main.fcgi

docker compose stop

docker compose up -d