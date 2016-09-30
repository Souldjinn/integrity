# interactive test to poke at the system.
go install github.com/bigcommerce-labs/integrity/cmd/integrity

go install github.com/bigcommerce-labs/integrity/cmd/endpoint
endpoint &
PID=$!
sleep 2
trap "kill $PID" SIGINT SIGTERM EXIT

integrity
echo