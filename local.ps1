$env:CGO_ENABLED="0"
$env:GOOS="linux"
$env:GOARCH="amd64"

templ generate -v

go build -ldflags "-s -w" -o main.fcgi

docker compose stop

docker compose up -d