# interactive test to poke at the system.
go install github.com/bigcommerce-labs/integrity/cmd/integrity

go install github.com/bigcommerce-labs/integrity/cmd/endpoint
endpoint -p 3456 &
PIDX=$!

endpoint -p 3457 &
PIDY=$!

sleep 2
trap "kill $PIDX; kill $PIDY" SIGINT SIGTERM EXIT

integrity test1.json
integrity test2.json

echo