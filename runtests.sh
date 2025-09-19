#!/bin/bash

# Initialize the event counter in Redis
redis-cli -n 2 SET event_counter 0

# Run all tests
go test ./test/... -v -count=1

# Read the final event count from Redis
count=$(redis-cli -n 2 GET event_counter)

# Print the final event count
printf "\n\nTotal events created: %s\n\n" "$count"

