#!/bin/bash

# Находим ID контейнера по имени
CONTAINER_ID=$(docker ps -aqf "name=postgres-urleater")

# Проверяем, найден ли контейнер
if [ -z "$CONTAINER_ID" ]; then
    echo "Контейнер с именем, соответствующим 'postgres-urleater', не найден."
    exit 1
fi

echo "Найден контейнер с ID: $CONTAINER_ID"

# SQL-команда для выполнения
SQL_COMMAND="TRUNCATE urls; TRUNCATE users CASCADE;"

# Выполняем команду psql внутри контейнера
docker exec -i "$CONTAINER_ID" psql "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -c "$SQL_COMMAND"

echo "Команда выполнена."
