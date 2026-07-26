#!/bin/sh
set -eu

mysql --protocol=socket -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
  filename VARCHAR(255) NOT NULL PRIMARY KEY,
  applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用数据库迁移记录';
SQL

for migration in $(find /bootstrap/migrations -type f -name '*.sql' | sort); do
  filename=$(basename "$migration")
  echo "Applying migration: $filename"
  mysql --protocol=socket -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < "$migration"
  mysql --protocol=socket -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" -e "INSERT IGNORE INTO schema_migrations (filename) VALUES ('$filename')"
done

if [ -f /bootstrap/seeds/seed.sql ]; then
  echo "Applying seed data"
  mysql --protocol=socket -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < /bootstrap/seeds/seed.sql
fi
