@echo off

echo Starting Mokosh and MariaDB...

docker compose up -d
docker compose logs -f

echo Done.