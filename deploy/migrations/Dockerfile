FROM postgres:16-alpine

WORKDIR /app

COPY backend/migrations /migrations
COPY deploy/migrations/run-migrations.sh /usr/local/bin/run-migrations

RUN chmod +x /usr/local/bin/run-migrations

USER postgres

ENTRYPOINT ["run-migrations"]
