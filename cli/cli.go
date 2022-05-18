package cli

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"os"
)

// CLI responsible for processing command line arguments
type CLI struct{}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS #-# Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet #-# Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS #-# Get balance of ADDRESS")
	// fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain #-# Print all the blocks of the blockchain")
	fmt.Println("  reindexutxo #-# Rebuilds the UTXO set")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -allowuser ADDRESS -ipfshas IPFSHASH -mine #-# Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
	fmt.Println("  startnode -miner ADDRESS -storageminer ADDRESS #-# Start a node -miner enables mining")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
	cli.validateArgs()

	// if nodeID == "" {
	// 	fmt.Printf("NODE_ID env. var is not set!")
	// 	os.Exit(1)
	// }

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	// addWalletCmd := flag.NewFlagSet("addwallet", flag.ExitOnError)
	// listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendContentCmd := flag.NewFlagSet("sendcontent", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")

	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")

	sendFrom := sendCmd.String("from", "", "Source wallet private key")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")

	sendContentFrom := sendContentCmd.String("from", "", "Source wallet private key")
	sendContentTo := sendContentCmd.String("to", "", "Destination wallet address")
	sendContentAllow := sendContentCmd.String("allowuser", "", "These user can access to this file")
	sendContentIPFShash := sendContentCmd.String("ipfshash", "", "Hash file of IPFS")
	sendContentAmount := sendContentCmd.Int("amount", 0, "Amount to send")

	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")
	startNodeStorageMiner := startNodeCmd.String("storageminer", "", "Enable storage mining mode and send reward to ADDRESS")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	// case "addwallet":
	// 	err := addWalletCmd.Parse(os.Args[2:])
	// 	if err != nil {
	// 		log.Panic(err)
	// 	}
	// case "listaddresses":

	// 	err := listAddressesCmd.Parse(os.Args[2:])
	// 	if err != nil {
	// 		log.Panic(err)
	// 	}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "sendcontent":
		err := sendContentCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.GetBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.CreateBlockchain(*createBlockchainAddress)
	}

	// if addWalletCmd.Parsed() {
	// 	cli.AddWallet([]byte(os.Args[2]))
	// }

	if createWalletCmd.Parsed() {
		cli.CreateWallet()
	}

	// if listAddressesCmd.Parsed() {
	// 	cli.listAddresses(nodeID)
	// }

	if printChainCmd.Parsed() {
		cli.PrintChain()
	}

	if reindexUTXOCmd.Parsed() {
		cli.ReindexUTXO()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.Send(*sendFrom, *sendTo, *sendAmount, *sendMine)
	}

	if sendContentCmd.Parsed() {

		alwuser := strings.Split(*sendContentAllow, "_")
		cli.SendProposal(*sendContentFrom, *sendContentTo, *sendContentAmount, alwuser, *sendContentIPFShash)
	}

	if startNodeCmd.Parsed() {
		StorageMiningAddress = *startNodeStorageMiner
		cli.StartNode(*startNodeMiner)
	}
}
