# Remove active containers
docker rm -f $(docker ps -aq)

export COMPOSE_PROJECT_NAME=bfl
docker-compose -f docker-compose.yml down

# Clear directory
rm data/ -rf
