@echo off

echo Pulling Docker images...

docker compose pull

echo:
echo Restarting Mokosh and MariaDB...

docker compose stop
docker compose up -d

echo Done.