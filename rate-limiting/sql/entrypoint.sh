#!/bin/bash

if [ ! -f /var/lib/postgresql/data/.initialized ]; then
    for f in /docker-entrypoint-initdb.d/*.sql; do
        psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "$f"
    done
    touch /var/lib/postgresql/data/.initialized
fi
exec docker-entrypoint.sh postgres