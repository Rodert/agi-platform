#!/bin/sh
set -eu

for migration in $(find /bootstrap/migrations -type f -name '*.sql' | sort); do
  echo "Applying migration: $(basename "$migration")"
  mysql --protocol=socket -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < "$migration"
done

if [ -f /bootstrap/seeds/seed.sql ]; then
  echo "Applying seed data"
  mysql --protocol=socket -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < /bootstrap/seeds/seed.sql
fi
