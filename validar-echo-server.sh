#!/bin/bash

# Create a temporary container with netcat installed
CONTAINER_ID=$(docker run -d --network tp0_testing_net alpine:latest sh -c "apk add --no-cache netcat-openbsd && sleep 10")

if [ -z "$CONTAINER_ID" ]; then
    echo "action: test_echo_server | result: fail"
    exit 1
fi

# Wait for netcat installation
sleep 2

# Test message to send
TEST_MSG="Hello Server"

# Send test message and capture response
RESPONSE=$(docker exec $CONTAINER_ID sh -c "echo '$TEST_MSG' | nc -w 5 server 12345")

# Clean up
docker rm -f $CONTAINER_ID > /dev/null

# Compare response (trimming any whitespace/newlines)
if [ "$(echo -n "$RESPONSE" | tr -d '\n\r')" = "$(echo -n "$TEST_MSG" | tr -d '\n\r')" ]; then
    echo "action: test_echo_server | result: success"
    exit 0
else
    echo "action: test_echo_server | result: fail"
    exit 1
fi