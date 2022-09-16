# Automated-testnet-builder
Based on the code below, the environment code builds a test environment with the mainnet forked version of the 1 validator system.

Please check the corresponding REPO for more information

https://github.com/cosmosquad-labs/replay/tree/crescent-v2-testnet

If the version of the block-demon changes, the REPLAY dependency should also change.
## Init ENV
Run local_env.sh and make sure to check the **lastblock** number of the sink and run Replay
```
git clone https://github.com/b-harvest/Automated-testnet-builder
cd Automated-testnet-builder

./local_env.sh
#sync complete Ctrl + X, lastblock check

git clone https://github.com/cosmosquad-labs/replay
cd replay
go install
replay cre_validator <lastblock>
# > exported-genesis.json

rm cre_validator/config/genesis.json
cp exported-genesis.json cre_validator/config/genesis.json
mv exported-genesis.json cre_sentry/config/genesis.json
crescentd tendermint unsafe-reset-all --home cre_validator
rm cre_validator/config/config.toml
mv cre_validator/config/config.toml.bak cre_validator/config/config.toml

crescentd start --home cre_validator
crescentd start --home cre_sentry
```

## Upgrade Testing
TBD items can always be changed, please check the status of the last block on the mainnet and fill it out
```
export BINARY=crescentd
export CHAINID=crescent-1-local-test
export TITLE=v3

#crescentd status | jq | grep latest_block_height
export UPGRADEHEIGHT=<TBD>

export KEY=validator

# submit upgrade proposal
$BINARY tx gov submit-proposal software-upgrade $TITLE \
	--title $TITLE \
  --upgrade-height $UPGRADEHEIGHT \
  --upgrade-info $TITLE \
  --description $TITLE \
	--deposit 500000000ucre \
	--gas 400000 \
	--from $KEY \
  --keyring-backend test \
	--chain-id $CHAINID \
	--broadcast-mode block \
	--output json -y


# vote
# crescentd q gov proposals | jq 
export PROPOSALID=<TBD>

crescentd tx gov vote $PROPOSALID yes \
  --from $KEY \
  --keyring-backend test \
	--chain-id $CHAINID \
	--broadcast-mode block \
	--output json -y
```

## Rollback process
Rollback process in case of bug after upgrade, Return to the very beginning of the forked network.
```
crescentd tendermint unsafe-reset-all --home cre_validator
crescentd tendermint unsafe-reset-all --home cre_sentry
crescentd start --home cre_validator
crescentd start --home cre_sentry
```

## TODO
Establish an IBC environment and integrated testing environment including backend and frontend, module testing cli, pingpub explorer
