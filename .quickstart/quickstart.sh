#!/bin/bash

#build
make -C ../ build

# # consul instance
# # -v /Users/<your_name>/tmp/consul:/consul/data \
docker run -d --name dev-consul -p 8500:8500 -p 8600:8600/udp \
-e CONSUL_BIND_INTERFACE=eth0 \
consul agent --bootstrap -server -ui -client=0.0.0.0

# postgres instance
# -v $PWD/pgmountlayouttmp:/var/lib/postgresql/data \
docker run -d -p 5438:5432 \
--name volumed-postgres-layout \
-e POSTGRES_PASSWORD=password \
-e PGDATA=/var/lib/postgresql/data/pgdata \
postgres:12

# run migration
docker build --no-cache --network=host ../migration/goose

# run kafka
docker run -d --rm --name kafka -p 9092:9092 -p 2181:2181 -e ADVERTISED_HOST=127.0.0.1 -e KAFKA_CREATE_TOPICS="animaladd:1:1" spotify/kafka

# Set consul configuration
root="go-micro-layouts"
appname="pub-sub-layout"
# Consul url
url=http://host.docker.internal:8500/v1/kv/${root}
token=$1
requestbody='{
  "app": {
    "topics": {
      "animalAdd": "animaladd",
      "animalAddRs": "animaladdrs"
    }
  },
  "server": {
    "name": "animal-add",
    "addr": ":8090"
  },
  "broker": {
    "clientid": "animal-add",
    "workers": 1,
    "reader": {
      "commit_interval": "5000ms",
      "min_bytes": 10,
      "max_bytes": 1048576
    },
    "addr": [
      "host.docker.internal:9092"
    ]
  },
  "postgres_primary": {
    "addr": "host.docker.internal:5438",
    "dbname": "postgres",
    "login": "postgres",
    "passw": "password",
    "conn_max": 80,
    "conn_lifetime": 10,
    "conn_maxidletime": 10
  },
  "logger": {
    "loglevel": "debug"
  },
  "metric": {
    "addr": ":8080"
  },
  "trace": {
    "mode": "agent"
  }
}'

# Configuration
# --- DO NOT EDIT BELOW ---
setConsulConfig () {
  echo "### Setting ${root}/${appname} as:"
  echo "${requestbody}"
  if [[ "$(curl -sX PUT -H "X-Consul-Token: ${token}" -d "${requestbody}" ${url}/${appname})" == "true" ]]; then
    echo "### ${url}/${appname} is set"
  else
    echo "### ERROR: Cannot set ${url}/${appname}"
    exit 1
  fi
}
setConsulConfig

#run service
../bin/app
