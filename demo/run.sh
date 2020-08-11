#!/bin/bash
set -e # exit on non-zero status code

# post-run cleanup
cleanup () {
  docker-compose kill
  docker-compose rm -f
}
trap 'cleanup ; printf "Tests have been killed via signal.\n"' HUP INT QUIT PIPE TERM

# build and run 5 docker images (what each container does is in data/client-script.sh)
docker-compose up -d

COLLECTIVE_PUB_KEY=""

docker stop -t 60 peer_master

while [ -z "$COLLECTIVE_PUB_KEY" ]
do
    COLLECTIVE_PUB_KEY=$(docker exec drand1 drand show chain-info | grep public_key | cut -d'"' -f 4)
    sleep 2
done

echo "DISTRIBUTED_PUB_KEY=${COLLECTIVE_PUB_KEY}" > .env

docker-compose up -d

echo ""
echo "Congratulations! the drand network is running."
echo ""
echo "Query a node's API: "
echo "  Linux: curl CONTAINER_IP:PORT/api/public "
echo "  alternative: docker exec drand1 call_api "
echo ""

docker-compose logs -f
cleanup