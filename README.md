# Deploying Your Go Web App To Shared Hosting

## Known Limitations

 - Can't use C-based Go Libraries, must be Purely Go based (due to linking issue).
 - Websockets dont work
 - GRPC will not work
 - HTTP2/3 probably not available, cannot provide early header hints either
 - Early flushing not possible
 - Unable to get errors in error_log or CPanel logs, you'll need to write to a file

# HostGator Specific Limitation

 - FcgidProcessLifeTime: 3600 seconds (1 hour). You're max run time if one hour before the process is killed.

## Enabling Fast-CGI / CGI 

All you need to do is add the Apache HTTPd handlers inside your `.htaccess` file.

**FILE: `public_html/.htaccess`**

```ini
# For Fast CGI you can use any extension, not limited to .fgci
AddHandler fcgid-script .fcgi

# For CGI, again any extension you want
AddHandler cgi-script .cgi .pl .plx .ppl .perl .py
```


# Switching From HTTP to FastCGI/CGI

Let says you some simple HTTP server setup in Go.

```go
http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Hello World!\n")
}))
```

To convert this to FastCGI, you swap `http.ListenAndServe(":8080", ...` to `fcgi.Serve(nil,`. So here is a full example
of FastCGI and CGI. Please note only one of these can run you do not set up CGI and FastCGI and HTTP, its one or the other.

```go
package main

import (
    "fmt"
    "net/http"
    "net/http/cgi"  // For CGI
    "net/http/fcgi"  // For FastCGI
)

func main() {
    // CGI
    // Only takes a handler
    cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "Hello World!\n")
    }))
    
    // Or FastCGI
    // Can tak a listener and handler, listen is for a unix socket, or port, setting it to nil will use IPC.
    fcgi.Serve(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "Hello World!\n")
    }))
}
```

## Compiling Go Web App To Run On Share Hosting

Due to glibc unknowns you will need to use `CGO_ENABLED=0`


**Linux/Mac**
```shell
go install github.com/a-h/templ/cmd/templ@latest
templ generate -v
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o main.fcgi
```

**Windows Powershell**
```shell
$env:CGO_ENABLED="0"
$env:GOOS="linux"
$env:GOARCH="amd64"
go install github.com/a-h/templ/cmd/templ@latest
templ generate -v
go build -ldflags "-s -w" -o main.fcgi
```

**Windows Terminal**
```shell
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go install github.com/a-h/templ/cmd/templ@latest
templ generate -v
go build -ldflags "-s -w" -o main.fcgi
```


## Local Development with Docker

You can run FastCGI via Docker with Apache HTTPd.

**File: `Dockerfile`**

```dockerfile
# Use the official Apache image from the Docker Hub
FROM httpd:2.4-buster

# Install necessary packages
RUN apt-get update && apt-get install -y \
    libapache2-mod-fcgid libfcgi0ldbl libfcgi-bin \
    --no-install-recommends && rm -rf /var/lib/apt/lists/*

# Manually enable modules by updating httpd.conf (Apache configuration)
RUN echo 'LoadModule rewrite_module modules/mod_rewrite.so' >> /usr/local/apache2/conf/httpd.conf
RUN echo 'LoadModule fcgid_module /usr/lib/apache2/modules/mod_fcgid.so' >> /usr/local/apache2/conf/httpd.conf

# Update the Apache configuration to handle FastCGI scripts
RUN echo 'AddHandler fcgid-script .fcgi' >> /usr/local/apache2/conf/httpd.conf
RUN echo '<Directory "/usr/local/apache2/cgi-bin/">' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    AllowOverride None' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    Options +ExecCGI -MultiViews +SymLinksIfOwnerMatch' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    Require all granted' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    AddHandler fcgid-script .fcgi' >> /usr/local/apache2/conf/httpd.conf
RUN echo '</Directory>' >> /usr/local/apache2/conf/httpd.conf
RUN echo '<IfModule mod_fcgid.c>' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    # Directory for sockets and shared memory file' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    FcgidIPCDir /var/lib/apache2/fcgid/sock' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    FcgidProcessTableFile /var/lib/apache2/fcgid/shm' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    FcgidOutputBufferSize 0' >> /usr/local/apache2/conf/httpd.conf
RUN echo '</IfModule>' >> /usr/local/apache2/conf/httpd.conf
RUN echo '<Location "/cgi-bin/main.fcgi/events">' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    SetEnv no-gzip 1' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    SetEnv no-buffer 1' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    SetEnv proxy-nokeepalive 1' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    Header always set Cache-Control "no-cache"' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    Header always set Content-Type "text/event-stream"' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    Header always set Connection "keep-alive"' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    Header always set X-Accel-Buffering "no"' >> /usr/local/apache2/conf/httpd.conf
RUN echo '    SetEnvIf Request_URI "^/cgi-bin/main.fcgi/events$" nofilter' >> /usr/local/apache2/conf/httpd.conf
RUN echo '</Location>' >> /usr/local/apache2/conf/httpd.conf
RUN echo 'Protocols h2c http/1.1' >> /usr/local/apache2/conf/httpd.conf
RUN echo 'LogLevel trace8' >> /usr/local/apache2/conf/httpd.conf
RUN echo 'KeepAlive On' >> /usr/local/apache2/conf/httpd.conf
RUN echo 'MaxKeepAliveRequests 1000' >> /usr/local/apache2/conf/httpd.conf
RUN echo 'KeepAliveTimeout 600' >> /usr/local/apache2/conf/httpd.conf

RUN chown -R daemon:daemon /var/lib/apache2/fcgid/sock && chmod -R 755 /var/lib/apache2/fcgid/sock

# Expose port 80
EXPOSE 80
```

**File: `docker-compose.yml`**

```yml
services:
  apache-fastcgi:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "80:80"
    volumes:
      - ./main.fcgi:/usr/local/apache2/cgi-bin/main.fcgi
```

The `docker compose up -d` and in your browser go to http://localhost/cgi-bin/main.fcgi/. There is a ./local.sh and
./local.ps1 to auto compile and reload docker.

## Deployment to Production

When you deploy with CGI, you only need to ensure the binary file is set to 0755 or "execute". For FastCGI, you need to
delete the file first, upload the new file, and change the file to also be 0755 or "execute".

**CGI**

 - Upload file
 - Set file permissions to 0755

**FastCGI**

 - Delete existing file
 - Upload new file
 - Set file permissions to 0755

```shell
CGO_ENABLED=0 go build -ldflags "-s -w" -o main.fcgi

sshpass -p "PASSWORD" ssh USERNAME@DOMAIN.COM "rm -f /home1/USERNAME/public_html/main.fcgi"

sshpass -p "PASSWORD" scp main.fcgi USERNAME@DOMAIN.COM:/home1/USERNAME/public_html/main.fcgi

sshpass -p "PASSWORD" ssh USERNAME@DOMAIN.COM "chmod 0755 /home1/USERNAME/public_html/main.fcgi"
```

## Relative Routes

One issue from local to production is the path may be different. One way to handle this is to override the path based on
the binary name which is included in URL.

```go
func main() {
	//...
    // This will override a route like /main.fcgi/hello to /hello
	http.Handle("/", dynamicPathAdjustMiddleware(r))
	//...
    if err := fcgi.Serve(nil, http.DefaultServeMux); err != nil {
        fmt.Println("Error serving FastCGI:", err)
    }
}

func dynamicPathAdjustMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Print incoming path for debugging
		//fmt.Printf("Original URL Path: %s\n", r.URL.Path)

		// Automatically find and trim up to '/main.fcgi'
		splitPath := strings.SplitN(r.URL.Path, "/main.fcgi", 2)
		if len(splitPath) > 1 {
			newPath := splitPath[1]
			if newPath == "" || newPath[0] != '/' {
				newPath = "/" + newPath
			}
			r.URL.Path = newPath
			//fmt.Printf("Adjusted URL Path to: %s\n", r.URL.Path)
		}

		// Proceed with the modified request
		next.ServeHTTP(w, r)
	})
}
```