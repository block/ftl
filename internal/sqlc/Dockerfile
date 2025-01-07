FROM alpine:latest

RUN apk add --no-cache bash curl tar

WORKDIR /app

COPY scripts/download-sqlc /app/download-sqlc

RUN chmod +x /app/download-sqlc

ENV SQLC_VERSION=""

CMD ["/app/download-sqlc"]
