#!/bin/sh

# read this container's IP address
IP_ADDR=`ip a | grep global | grep -oE "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b" | tail -n -2 | head -n 1`
IP_ADDR_PORT="${IP_ADDR}:8000"

# Generate key pair
echo "Generating key pair..."
rm -rf /root/.drand/
drand generate-keypair --tls-disable "${IP_ADDR_PORT}"

# Boot the drand deamon in background
nohup drand start --tls-disable --public-listen 0.0.0.0:8081 --goshimmerAPIurl "$GOSHIMMER" & # add "--verbose 2" here for more details

# Wait for all containers to have done the same
sleep 5

# Now nodes wait for the leader to run DKG; leader starts DKG
if [[ "$LEADER" == 1 ]]; then
    echo "Running DKG..."
    drand share --leader --nodes 5 --threshold 3 --secret "0Q0rRqhUX99nn4SoYME90McZKk+MNtx0OLT5/HTk1tE=" --period 10s
else
    sleep 5
    drand share --connect "testdrng-drand_0:8000" --tls-disable --nodes 5 --threshold 3 --secret "0Q0rRqhUX99nn4SoYME90McZKk+MNtx0OLT5/HTk1tE="
fi

# Let the deamon alive for long enough
while true
do
sleep 2
done

echo "Done"