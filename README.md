# Automated-testnet-builder
Based on the code below, the environment code builds a test environment with the mainnet forked version of the 2 validator system.
## Init ENV
Run local_env.sh and make sure to check the **lastblock** number of the sink and run Replay
```
git clone https://github.com/b-harvest/Automated-testnet-builder
cd Automated-testnet-builder

./local_env.sh
#sync complete Ctrl + X, lastblock check

cd replay
go install
cd ..
replay canto_val1/data 1477402
# > exported-genesis.json

rm canto_val1/config/genesis.json
cp exported-genesis.json canto_val1/config/genesis.json
mv exported-genesis.json canto_val2/config/genesis.json
cantod tendermint unsafe-reset-all --home canto_val1
rm canto_val1/config/config.toml
mv canto_val1/config/config.toml.bak canto_val1/config/config.toml

cantod start --home canto_val1 --x-crisis-skip-assert-invariants
cantod start --home canto_val2 --x-crisis-skip-assert-invariants

#Each validator's wallet is in the Home folder.
```

## Rollback process
Rollback process in case of bug after upgrade, Return to the very beginning of the forked network.
```
cantod tendermint unsafe-reset-all --home canto_val1
cantod tendermint unsafe-reset-all --home canto_val2
```
**Returns the binary to a previous version of the upgrade.**(In this example, V2 version)
```
cantod start --home canto_val1
cantod start --home canto_val2
```
Please proceed with the Upgrade Testing process again

## TODO
Establish an IBC environment and integrated testing environment including backend and frontend, module testing cli, pingpub explorer
