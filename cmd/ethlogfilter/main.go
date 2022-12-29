package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/artnoi43/gsl/gslutils"
	"github.com/artnoi43/gsl/soyutils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/net/context"
)

type config struct {
	ConfigFile string `yaml:"-" arg:"-c,--config" placeholder:"FILE" help:"Config file to read"`
	Verbose    bool   `yaml:"-" arg:"-v,--verbose" help:"Verbose output (CLI arg only, placing this in config file won't work)"`
	OutputFile string `yaml:"-" arg:"-o,--outfile" help:"Write JSON output to this file"`

	NodeURL   string   `yaml:"node_url" arg:"-n,--node-url" placeholder:"NODE_URL" help:"HTTP or WS URL of an Ethereum node"`
	Addresses []string `yaml:"addresses" arg:"-a,--addresses" placeholder:"ADDR [ADDR..]" help:"Contract address"`
	Topics    []string `yaml:"topics" arg:"--topics" placeholder:"TOPICS [TOPIC..]" help:"Logs topics"`
	TxHashes  []string `yaml:"tx_hashes" arg:"-x,--tx-hashes" placeholder:"HASH [HASH..]" help:"Transaction hashes"`
	FromBlock uint64   `yaml:"from_block" arg:"-f,--from-block" placeholder:"FROM_BLOCK" help:"Filter from block"`
	ToBlock   uint64   `yaml:"to_block" arg:"-t,--to-block" placeholder:"TO_BLOCK" help:"Filter to block"`
	LogBlock  uint64   `yaml:"block" arg:"-b,--block" placeholder:"LOG_BLOCK" help:"Filter logs from this block (overwrites FROM_BLOCK and TO_BLOCK)"`
}

func main() {
	// Parse config
	argConf := new(config)
	arg.MustParse(argConf)

	var configFile string
	if len(argConf.ConfigFile) > 0 {
		configFile = argConf.ConfigFile
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			panic("failed to get config location in user's homedir: " + err.Error())
		}
		configFile = home + "/.config/ethlogfilter/config.yaml"
	}

	// Read in config
	fileConf, err := soyutils.ReadFileYAMLPointer[config](configFile)
	if err != nil {
		panic("read config failed: " + err.Error())
	}

	// Overwrite config from file with CLI args
	fileConf = mergeConfig(fileConf, argConf)

	client, err := ethclient.Dial(fileConf.NodeURL)
	if err != nil {
		panic("new client failed: " + err.Error())
	}

	var addresses []common.Address
	for _, addrString := range fileConf.Addresses {
		address := common.HexToAddress(addrString)
		addresses = append(addresses, address)
	}

	var topics []common.Hash
	for _, topicsStr := range fileConf.Topics {
		topic := common.HexToHash(topicsStr)
		topics = append(topics, topic)
	}

	if fileConf.Verbose {
		fmt.Println("Filter addresses", addresses)
		fmt.Println("Filter topics", topics)
		fmt.Println("Filter txHashes", fileConf.TxHashes)
	}

	// If |conf.LogBlock| is given, then fromBlock and toBlock is both |conf.LogBlock|,
	// otherwise `conf.FromBlock| and |conf.ToBlock| are used.
	fromBlock, toBlock := chooseBlock(fileConf.FromBlock, fileConf.ToBlock, fileConf.LogBlock)

	logs, err := client.FilterLogs(context.Background(), ethereum.FilterQuery{
		FromBlock: chooseBlockNumber(fromBlock),
		ToBlock:   chooseBlockNumber(toBlock),
		Addresses: chooseAddresses(addresses),
		Topics:    chooseTopics(topics),
	})

	// Collect |logs| ([]types.Log) into a []*types.Log,
	// and filtering for TxHash if we got one from the user.
	var targetLogs []*types.Log
	if len(fileConf.TxHashes) > 0 {
		targetLogs = gslutils.CollectPointersIf(&logs, func(log types.Log) bool {
			return gslutils.Contains(fileConf.TxHashes, log.TxHash.String())
		})
	} else {
		targetLogs = gslutils.CollectPointers(&logs)
	}

	if err != nil {
		panic("failed to filterLogs: " + err.Error())
	}

	logsJson, err := json.Marshal(targetLogs)
	if err != nil {
		panic("failed to marshal event logs to json: " + err.Error())
	}

	if len(fileConf.OutputFile) != 0 {
		if err := os.WriteFile(fileConf.OutputFile, logsJson, os.ModePerm); err != nil {
			fmt.Println("failed to write JSON output to file", fileConf.OutputFile)
			fmt.Println(err.Error())
		}
	}

	fmt.Printf("%s\n", logsJson)
}

// Use |logBlock| as fromBlock and toBlock if not 0
func chooseBlock(fromBlock, toBlock, logBlock uint64) (uint64, uint64) {
	if logBlock == 0 {
		return fromBlock, toBlock
	}
	return logBlock, logBlock
}

// Return nil if |b| == 0 to get desired behavior from ethclient.Client.FilterQuery
func chooseBlockNumber(b uint64) *big.Int {
	if b == 0 {
		return nil
	}

	return big.NewInt(int64(b))
}

func chooseAddresses(addresses []common.Address) []common.Address {
	if len(addresses) == 0 {
		return nil
	}

	return addresses
}

func chooseTopics(topics []common.Hash) [][]common.Hash {
	if len(topics) == 0 {
		return nil
	}

	return [][]common.Hash{topics}
}

// mergeConfig returns a merged config from |a| and |b|
// mergeConfig only uses a field value from |a| ONLY if that field is null/zero-valued in |b|,
// i.e. it uses b to overwrite |a|.
func mergeConfig(a, b *config) *config {
	if b == nil {
		if a == nil {
			return nil
		}
		return a
	}

	out := new(config)

	// Verbosity can only be toggled in arg
	out.Verbose = b.Verbose

	if len(a.NodeURL) == 0 {
		if len(b.NodeURL) != 0 {
			out.NodeURL = b.NodeURL
		}
	} else {
		out.NodeURL = a.NodeURL
	}

	if len(b.Addresses) != 0 {
		out.Addresses = b.Addresses
	} else {
		out.Addresses = a.Addresses
	}

	if len(b.Topics) > 0 {
		out.Topics = b.Topics
	} else {
		out.Topics = a.Topics
	}

	if len(b.TxHashes) != 0 {
		out.TxHashes = b.TxHashes
	} else {
		out.TxHashes = a.TxHashes
	}

	if b.FromBlock != 0 {
		out.FromBlock = b.FromBlock
	} else {
		out.FromBlock = a.FromBlock
	}

	if b.ToBlock != 0 {
		out.ToBlock = b.ToBlock
	} else {
		out.ToBlock = a.ToBlock
	}

	if b.OutputFile != "" {
		out.OutputFile = b.OutputFile
	} else {
		out.OutputFile = b.OutputFile
	}

	return out
}
