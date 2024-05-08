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
