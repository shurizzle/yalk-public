FROM golang:latest

WORKDIR /app
COPY . .

EXPOSE 80 443

# It is adised to use '0.0.0.0' to allow all incoming connections
ENV HOST_ADDR="0.0.0.0"
ENV HTTP_PORT=80
ENV HTTPS_PORT=443
ENV WEB_URL="https://localhost"

ENV SOCKET_PORT=9988
ENV SOCKET_TRANSPORT="tcp"

ENV DB_ADDR="172.17.0.2"
ENV DB_PORT=5432
ENV DB_NAME="db_chat"
ENV DB_USER="postgres"
ENV DB_PASSWORD="changeme"
ENV DB_SSLMODE="disable"


 CMD ["go", "run", "."]
