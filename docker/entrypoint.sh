#!bin/sh

if [ ! -f "/root/.drand/key/drand_id.private" ]; then
    drand generate-keypair --tls-disable "${DRAND_PUBLIC_ADDRESS}"
fi

exec drand $@
