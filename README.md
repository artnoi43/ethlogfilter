# ethlogfilter

ethlogfilter is a small, single-file script for printing Ethereum logs
based on some criteria. The current version can filter logs using `go-ethereum`.

It outputs the logs in JSON format, and is especially useful for Web3 devs.

## Command examples

Let's say we want filter all logs from [this Mainnet transaction on block
15847894 (`0x07fff3cd11172e3878900dd22e8e905674651aa5f91f04ff35926150d2db9671`)](https://etherscan.io/tx/0x07fff3cd11172e3878900dd22e8e905674651aa5f91f04ff35926150d2db9671#eventlog),
then we can do:

```bash
ethlogfilter\
        --block 15847894\
        --tx-hashes 0x07fff3cd11172e3878900dd22e8e905674651aa5f91f04ff35926150d2db9671;
```

Or if you want to filter all logs from address `0x007` and `0x789`,
with topics `0x696969` and `0x11210` during blocks 200-300, then you can do:

```bash
ethlogfilter\
        --from-block 200\
        --to-block 300\
        --addresses 0x007 0x789\
        --topics 0x696969 0x11210;
```

The output of this script is JSON.

## Configuration and command flags

1. Block numbers (3 `uint64`s)

   > Config file field for `config.FromBlock`: `from_block`  
   > Config file field for `config.ToBlock`: `to_block`  
   > Config file field for `config.LogBlock`: `block`  
   > CLI argument flag for `config.FromBlock`: `--from-block`  
   > CLI argument flag for `config.ToBlock`: `--to-block`  
   > CLI argument flag for `config.LogBlock`: `--block`

   With this option, you can either specify a specific block number with
   _log block number_, or specify a _range_ of block numbers
   using _from block_ and _to block_.

   Specifying a _log block_ overwrites _from_ and _to_ blocks with the value.

   To filter logs in a single block, use _log block_ or
   use the same block number for both _from_ and _to_ block.

   If you omit _from block_, then it filters from **the 1st block**,
   and if you omit _to_block_, then it filters to **the latest block**.

2. Contract addresses (`[]string`)

   > Config file field: `addresses`  
   > CLI argument flag `-a` or `--addresses`

   You can use contract addresses to filter event logs - only logs emitted
   from one of the user-provided addresses will be filtered

3. Log topics (`[]string`)

   > Config file field: `topics`  
   > CLI argument flag: `--topics`

   You can use log topics to filter event logs - only logs whose topics match
   one of the user-provided topics will be filtered.

4. Transaction hashes

   > Config file field: `tx_hashes`  
   > CLI argument flag: `-x` or `--tx-hashes`

   User can choose to filter only logs with TX hash matching one of the
   provided TX hashes. Note that this is done _after_ the Ethereum client
   gets the logs (`ethereum.FilterQuery` does not contain field `TxHash`).
