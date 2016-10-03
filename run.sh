#!/bin/bash

# This script pops up sample endpoints and an integrity server
# then lets it run continuously.

go install github.com/bigcommerce-labs/integrity/cmd/integrity

go install github.com/bigcommerce-labs/integrity/cmd/endpoint
endpoint -p 3456 &
PIDX=$!

endpoint -p 3457 &
PIDY=$!

sleep 2
trap "kill $PIDX; kill $PIDY" SIGINT SIGTERM EXIT

integrity ./cmd/endpoint/

echo