templ generate -v

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o main.fcgi

sshpass -p "PASSWORD" ssh USERNAME@DOMAIN.COM "rm -f /home1/USERNAME/public_html/main.fcgi"

sshpass -p "PASSWORD" scp main.fcgi USERNAME@DOMAIN.COM:/home1/USERNAME/public_html/main.fcgi

sshpass -p "PASSWORD" ssh USERNAME@DOMAIN.COM "chmod 0755 /home1/USERNAME/public_html/main.fcgi"
