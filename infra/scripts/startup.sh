#!/bin/bash
APP_DIR="/app"

echo "DATABASE_URL="$DATABASE_URL"" >> ${APP_DIR}/.env
echo "POSTGRES_ENABLESSL="$POSTGRES_ENABLESSL"" >> ${APP_DIR}/.env

echo "${@}" | xargs -I % /bin/bash -c '%'