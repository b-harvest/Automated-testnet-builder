#!/bin/bash
wget https://raw.githubusercontent.com/Canto-Network/Canto/v4.0.0/Networks/Mainnet/genesis.json
mv genesis.json canto_val1/config/genesis.json

wget https://snapshots.polkachu.com/addrbook/canto/addrbook.json
mv addrbook.json canto_val1/config/addrbook.json

SNAP_RPC="https://endpoint1.bharvest.io/rpc/canto"

LATEST_HEIGHT=$(curl -s $SNAP_RPC/block | jq -r .result.block.header.height); \
BLOCK_HEIGHT=$((LATEST_HEIGHT - 2000)); \
TRUST_HASH=$(curl -s "$SNAP_RPC/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)

sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ; \
s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"$SNAP_RPC,$SNAP_RPC\"| ; \
s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$BLOCK_HEIGHT| ; \
s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"| ; \
s|^(seeds[[:space:]]+=[[:space:]]+).*$|\1\"\"|" canto_val1/config/config.toml

cantod start --home canto_val1