
# Create data dir
mkdir -p data/ca-certs
mkdir -p data/logs

# Copy configtx.yml to data folder
cp configtx.yaml data/configtx.yaml


export COMPOSE_PROJECT_NAME=bfl
# Start docker-composer file
docker-compose -f docker-compose.yml up -d
