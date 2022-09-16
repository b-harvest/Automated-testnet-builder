#!/bin/bash
wget https://blocksnapshot.s3.ap-northeast-2.amazonaws.com/crescent-1-genesis.json
mv crescent-1-genesis.json cre_validator/config/genesis.json

wget https://blocksnapshot.s3.ap-northeast-2.amazonaws.com/addrbook.json
mv addrbook.json cre_validator/config/addrbook.json

SNAP_RPC="http://54.95.40.202:26657"

LATEST_HEIGHT=$(curl -s $SNAP_RPC/block | jq -r .result.block.header.height); \
BLOCK_HEIGHT=$((LATEST_HEIGHT - 2000)); \
TRUST_HASH=$(curl -s "$SNAP_RPC/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)

sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ; \
s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"$SNAP_RPC,$SNAP_RPC\"| ; \
s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$BLOCK_HEIGHT| ; \
s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"| ; \
s|^(seeds[[:space:]]+=[[:space:]]+).*$|\1\"\"|" cre_validator/config/config.toml

crescentd start --home cre_validator