#!/bin/bash
if ! docker compose -f docker-compose-dev.yaml up -d; then
    echo "action: test_echo_server | result: fail"
    exit 1
fi

sleep 2

CONTAINER_ID=$(docker run -d --network tp0_testing_net alpine:latest sh -c "apk add --no-cache netcat-openbsd && sleep 60")

if [ -z "$CONTAINER_ID" ]; then
    echo "action: test_echo_server | result: fail"
    exit 1
fi

sleep 2

TEST_MSG="Hello Server"

RESPONSE=$(docker exec $CONTAINER_ID sh -c "echo '$TEST_MSG' | nc server 12345")

docker rm -f $CONTAINER_ID > /dev/null

if [ "$RESPONSE" = "$TEST_MSG" ]; then
    echo "action: test_echo_server | result: success"
    exit 0
else
    echo "action: test_echo_server | result: fail"
    exit 1
fi