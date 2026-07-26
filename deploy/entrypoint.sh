#!/bin/sh
set -eu

/app/migrate
exec supervisord -c /etc/supervisord.conf
