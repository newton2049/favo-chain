#!/bin/sh

set -e

FAVO_EDGE_BIN=./favo-chain
CHAIN_CUSTOM_OPTIONS=$(tr "\n" " " << EOL
--block-gas-limit 10000000
--epoch-size 10
--chain-id 51001
--name favo-chain-docker
--premine 0x228466F2C715CbEC05dEAbfAc040ce3619d7CF0B:0xD3C21BCECCEDA1000000
--premine 0xca48694ebcB2548dF5030372BE4dAad694ef174e:0xD3C21BCECCEDA1000000
EOL
)

case "$1" in

   "init")
      case "$2" in 
         "ibft")
         if [ -f "$GENESIS_PATH" ]; then
              echo "Secrets have already been generated."
         else
              echo "Generating secrets..."
              secrets=$("$FAVO_EDGE_BIN" secrets init --insecure --num 4 --data-dir /data/data- --json)
              echo "Secrets have been successfully generated"
              echo "Generating IBFT Genesis file..."
              cd /data && /favo-chain/favo-chain genesis $CHAIN_CUSTOM_OPTIONS \
                --dir genesis.json \
                --consensus ibft \
                --ibft-validators-prefix-path data- \
                --validator-set-size=4 \
                --bootnode "/dns4/node-1/tcp/1478/p2p/$(echo "$secrets" | jq -r '.[0] | .node_id')" \
                --bootnode "/dns4/node-2/tcp/1478/p2p/$(echo "$secrets" | jq -r '.[1] | .node_id')"
         fi
              ;;
          "favobft")
              echo "Generating FavoBFT secrets..."
              secrets=$("$FAVO_EDGE_BIN" favobft-secrets init --insecure --num 4 --data-dir /data/data- --json)
              echo "Secrets have been successfully generated"

              echo "Generating manifest..."
              "$FAVO_EDGE_BIN" manifest --path /data/manifest.json --validators-path /data --validators-prefix data-

              echo "Generating FavoBFT Genesis file..."
              "$FAVO_EDGE_BIN" genesis $CHAIN_CUSTOM_OPTIONS \
                --dir /data/genesis.json \
                --consensus favobft \
                --manifest /data/manifest.json \
                --validator-set-size=4 \
                --bootnode "/dns4/node-1/tcp/1478/p2p/$(echo "$secrets" | jq -r '.[0] | .node_id')" \
                --bootnode "/dns4/node-2/tcp/1478/p2p/$(echo "$secrets" | jq -r '.[1] | .node_id')"
              ;;
      esac
      ;;

   *)
      echo "Executing favo-chain..."
      exec "$FAVO_EDGE_BIN" "$@"
      ;;

esac
