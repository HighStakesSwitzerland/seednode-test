# SeedNode Test app

This tool runs a seed node for multiple tendermint based blockchain (i.e. Terra, Cosmos, Akash, etc, thanks to TinySeed), crawls the network and expose the list of peers with the geolocation.

###Configuration

```bash
git clone https://github.com/HighStakesSwitzerland/seednode-test
go mod tidy
npm install
npm run build
go install .
./seednode-test
# here execution will generate the config file and exit (with an error)
# adapt the chain-id and ports as needed, and rerun
# when the app is running, deploy the frontend using nginx or something else
```

A file `$HOME/.seednode-test/config.toml` will be generated if it doesn't exist yet, with some default parameters, and the program will exit.

You need to fill the and `chain_id` for every chain and start it again.
It may take few minutes/hours before discovering peers, depending on the network. 

## License

[Blue Oak Model License 1.0.0](https://blueoakcouncil.org/license/1.0.0)
