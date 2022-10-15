package rpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/wallet"
	"go.neonxp.dev/jsonrpc2/rpc"
	"go.neonxp.dev/jsonrpc2/rpc/middleware"
	"go.neonxp.dev/jsonrpc2/transport"

	"github.com/montanaflynn/stats"
)

// API version constants
// const (
// 	jsonrpcSemverString = "0.1.0"
// 	jsonrpcSemverMajor  = 0
// 	jsonrpcSemverMinor  = 1
// 	jsonrpcSemverPatch  = 0
// )

// "getbestblockhash":      handleGetbestblockhash,
// "getblockchaininfo":     handleGetblockchaininfo,
// "getblockcount":         handleGetblockcount,
// "getblockfilter":        handleGetblockfilter,
// "getblockhash":          handleGetblockhash,
// "getblockheader":        handleGetblockheader,
// "getblockstats":         handleGetblockstats,
// "getchaintips":          handleGetchaintips,
// "getchaintxstats":       handleGetchaintxstats,
// "getdifficulty":         handleGetdifficulty,
// "getmempoolancestors":   handleGetmempoolancestors,
// "getmempooldescendants": handleGetmempooldescendants,
// "getmempoolentry":       handleGetmempoolentry,
// "getmempoolinfo":        handleGetmempoolinfo,
// "getrawmempool":         handleGetrawmempool,
// "gettxout":              handleGettxout,
// "gettxoutproof":         handleGettxoutproof,
// "gettxoutsetinfo":       handleGettxoutsetinfo,
// "preciousblock":         handlePreciousblock,
// "pruneblockchain":       handlePruneblockchain,
// "savemempool":           handleSavemempool,
// "scantxoutset":          handleScantxoutset,
// "verifychain":           handleVerifychain,
// "verifytxoutproof":      handleVerifytxoutproof,

func InitJSONRPCServer(uri string) *rpc.RpcServer {
	s := rpc.New(
		rpc.WithLogger(rpc.StdLogger),
		rpc.WithTransport(&transport.HTTP{Bind: uri, CORSOrigin: "*"}),
	)
	s.Use(
		// rpc.WithTransport(&transport.TCP{Bind: ":3005"}),
		rpc.WithMiddleware(middleware.Logger(rpc.StdLogger)))

	s.Register("getblock", rpc.H(handleGetblock))
	s.Register("getbestblockhash", rpc.HS(handleGetBestBlockhash))
	s.Register("getblockcount", rpc.HS(handleGetblockcount))
	s.Register("validateaddress", rpc.H(handleValidateaddress))
	s.Register("createrawtransaction", rpc.H(handleCreaterawtransaction))
	s.Register("sendrawtransaction", rpc.H(handleSendrawtransaction))
	s.Register("gettxout", rpc.H(handleGettxout))

	s.Register("gettransaction", rpc.H(handleGettransaction))
	s.Register("getrawtransaction", rpc.H(handleGetrawtransaction))
	s.Register("getrawmempool", rpc.H(handleGetrawmempool))
	s.Register("getnetworkhashps", rpc.H(handleGetnetworkhashPs))
	s.Register("getnetworkinfo", rpc.HS(handleGetnetworkinfo))
	s.Register("getmempoolinfo", rpc.HS(handleGetmempoolinfo))
	s.Register("getmempoolentry", rpc.H(handleGetmempoolentry))
	s.Register("getblockchaininfo", rpc.HS(handleGetblockchaininfo))
	s.Register("getdifficulty", rpc.HS(handleGetdifficulty))
	s.Register("getchaintips", rpc.HS(handleGetchaintips))
	s.Register("getblockheader", rpc.H(handleGetblockheader))
	s.Register("decoderawtransaction", rpc.H(handleDecoderawtransaction))
	s.Register("getblockstats", rpc.H(handleGetblockstats))
	s.Register("listtransactions", rpc.H(handleListtransactions))
	return s
}

func handleGetblock(ctx context.Context, rawparams *Params) (interface{}, error) {

	// cast params from request
	params := GetBlock{Verbose: 0}
	err := CastParamsTo(rawparams, &params, 1, 2)
	if err != nil {
		return nil, err
	}
	bc := ctx.Value(Bckey).(*blockchain.Blockchain)
	// var bc = blockchain.NewBlockchainView()
	best, err := bc.GetLastBlock()
	if err != nil {
		return nil, err
	}
	var blocks = blockchain.Blockchain{DB: bc.DB}

	blkhashDec, err := HashToBytes(params.Blockhash)
	if err != nil {
		return nil, NewRPCError(ErrRPCInvalidAddressOrKey, err.Error())
	}
	blockdata, err := blocks.GetBlock(blkhashDec)
	if err != nil {
		return nil, NewRPCError(ErrRPCInternal.Code, err.Error())
	}
	result, err := BlockDataToResp(blockdata, params.Verbose)
	result.Confirmations = best.Height - blockdata.Height
	if err != nil {
		return nil, err
	}

	if params.Verbose == 0 {
		return DataToHex(result)
	} else {
		return result, nil
	}

}

func BlockDataToResp(T blockchain.Block, verbose int) (BlockDataResp, error) {
	resp := BlockDataResp{}

	hash, err := BytesToHex(T.Hash)
	if err != nil {
		return BlockDataResp{}, err
	}
	resp.Hash = Hex(hash)

	resp.NumOfTxs = len(T.Transactions)

	resp.StrippedSize = BLOCK_HEADER_SIZE

	resp.Size = resp.StrippedSize + resp.NumOfTxs*128

	resp.Height = T.Height

	resp.Weight = 4 * 1024 * 1024

	resp.Version = 100000000

	resp.VersionHex = Hex("0x" + fmt.Sprintf("%x", resp.Version))

	if verbose == 1 {
		var txHashes []TransactionID
		for _, tx := range T.Transactions {
			txID, _ := BytesToHex(tx.ID)
			txHashes = append(txHashes, TransactionID(txID))
		}
		resp.Transactions = txHashes
	} else {
		resp.Transactions = T.Transactions
	}

	resp.Timestamp = T.Timestamp

	resp.MedianTime = resp.Timestamp + 3600

	resp.Nonce = T.Nonce

	resp.Bits = "1d00ffff"

	resp.Difficulty = 2 << 16

	resp.ChainWork = "1d1d1d1d1d"

	resp.PrevBlockHash = Hex(T.PrevBlockHash)

	return resp, nil

}

func handleGetBestBlockhash(ctx context.Context) (interface{}, error) {
	result := Hex("")

	bc := ctx.Value(Bckey).(*blockchain.Blockchain)
	// var bc = blockchain.NewBlockchainView()

	var blocks = blockchain.Blockchain{DB: bc.DB}
	latest, err := blocks.GetLastBlock()

	if err != nil {
		return nil, NewRPCError(ErrRPCInternal.Code, err.Error())
	}
	result, err = DataToHex(latest)
	return result, err

}

func handleGetblockcount(ctx context.Context) (interface{}, error) {

	bc := ctx.Value(Bckey).(*blockchain.Blockchain)
	// var bc = blockchain.NewBlockchainView()
	var blocks = blockchain.Blockchain{DB: bc.DB}

	lastcount := blocks.GetBestHeight()

	return lastcount, nil
}

// func handleGetblockfilter(ctx context.Context, rawparams *Params) (interface{},error) {

// }

func handleCreaterawtransaction(ctx context.Context, rawparams *Params) (interface{}, error) {

	// var bc = blockchain.NewBlockchainView()

	// var blocks = blockchain.Blockchain{DB: bc.DB}

	nofparams := len(*rawparams)

	if nofparams < 2 || nofparams > 3 {
		return nil, ErrRPCInvalidParams
	}

	rawTxInputs := []RawTransactionInput{}

	err := json.Unmarshal([]byte((*rawparams)[0].(string)), &rawTxInputs)

	if err != nil {
		return nil, err
	}
	txvins := []blockchain.TXInput{}
	for _, vin := range rawTxInputs {

		// _, err := blocks.FindTransaction([]byte(vin.Txid))
		// if err != nil {
		// 	return nil, err
		// }
		txvins = append(txvins, blockchain.TXInput{Txid: []byte(vin.Txid), Vout: vin.Vout})
	}

	rawAmounts := map[string]float64{}

	err = json.Unmarshal([]byte((*rawparams)[1].(string)), &rawAmounts)

	if err != nil {
		return nil, err
	}
	txouts := []blockchain.TXOutput{}

	for addr, amount := range rawAmounts {
		if amount < 0 {
			return nil, ErrRPCInvalidParams
		}
		txouts = append(txouts, *blockchain.NewTXOutput(int(amount), addr))
	}

	tx := blockchain.Transaction{Vin: txvins, Vout: txouts}

	tx.ID = tx.Hash()

	return DataToHex(tx)
}

func handleValidateaddress(ctx context.Context, rawparam *Params) (interface{}, error) {
	var address = (*rawparam)[0].(string)

	result := ValidateAddressResp{}

	isAddress := wallet.ValidateAddress(address)

	if isAddress {
		result.IsValid = false
		result.Address = address
		result.IsScript = false
		result.IsWitness = false

	}

	return result, nil
}

func handleSendrawtransaction(ctx context.Context, rawparams *Params) (interface{}, error) {
	var params = SendRawTransaction{}

	err := CastParamsTo(rawparams, &params, 1, 2)

	if err != nil {
		return nil, err
	}
	if len(params.Data)%2 != 0 {
		params.Data = "0" + params.Data
	}

	rawData, err := HextoBytes(params.Data)

	if err != nil {
		return nil, err
	}

	tx := blockchain.Transaction{}

	_, err = DataToObject(rawData, &tx)

	if err != nil {
		return nil, err
	}

	return tx.Hash(), nil

}

func handleGettxout(ctx context.Context, rawparams *Params) (interface{}, error) {
	params := GetTxOut{IncludeMempool: true}

	CastParamsTo(rawparams, &params, 2, 3)

	bc := ctx.Value(Bckey).(*blockchain.Blockchain)

	best, err := bc.GetLastBlock()

	if err != nil {
		return nil, err
	}
	bestBlockHash, err := DataToHex(best)
	if err != nil {
		return nil, err
	}
	tx, err := bc.FindTransaction([]byte(params.TxId))
	if err != nil {
		return nil, err
	}
	txouts := tx.Vout
	if len(txouts) <= 0 {
		errStr := fmt.Sprintf("Output  for txid: %s does not exist", params.TxId)
		return nil, NewRPCError(ErrRPCNoTxInfo, errStr)
	}
	vout := params.Vout
	if err != nil {
		return nil, NewRPCError(ErrRPCInvalidTxVout, fmt.Sprintf("Output index: %d  for txid: %s does not exist", params.Vout, params.TxId))
	}
	txout := txouts[vout]
	isCoinbase := tx.IsCoinbase()
	resp := GetTxoutResp{BestBlock: bestBlockHash, Confirmations: 1, Value: float64(txout.Value), Coinbase: isCoinbase}
	return resp, nil
}
func handleGettransaction(ctx context.Context, rawparams *Params) (interface{}, error) {

	res := TransactionData{}
	bc := ctx.Value(Bckey).(*blockchain.Blockchain)

	params := GetTransaction{}
	CastParamsTo(rawparams, &params, 1, 3)

	tx, err := bc.FindTransaction([]byte(params.TxId))
	if err != nil {
		return nil, err
	}

	best, err := bc.GetLastBlock()
	if err != nil {
		return nil, err
	}

	block, err := FindBlockHasTx(bc, params.TxId)
	if err != nil {
		return nil, err
	}
	amount_inp := 0
	for _, txinp := range tx.Vin {
		amount_inp += txinp.Vout
	}

	amount_out := 0
	for _, txout := range tx.Vout {
		amount_out += txout.Value
	}

	res.Amount = amount_inp
	res.Fee = amount_inp - amount_out
	res.Confirmations = best.Height - block.Height
	res.BIP125_Replaceable = "no"
	res.BlockHash = Hex(block.Hash)
	res.BlockHeight = block.Height
	res.Comment = "cc"
	res.Decoded, _ = DataToHex(tx)
	res.Generated = true
	res.Time = block.Timestamp
	res.TimeReceived = block.Timestamp
	res.Trusted = true
	res.TxID = Hex(tx.ID)
	res.WalletConflicts = []string{}
	return res, nil
}

func FindBlockHasTx(bc *blockchain.Blockchain, txid string) (*blockchain.Block, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, []byte(txid)) {
				return block, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return &blockchain.Block{}, nil

}

func handleGetrawtransaction(ctx context.Context, rawparams *Params) (interface{}, error) {

	bc := ctx.Value(Bckey).(*blockchain.Blockchain)

	params := GetRawTransactionCmd{Verbose: false}
	CastParamsTo(rawparams, &params, 1, 3)
	block := blockchain.Block{}
	bestH := bc.GetBestHeight()
	if params.BlockHash == "" {
		_block, err := FindBlockHasTx(bc, params.Txid)
		if err != nil {
			return nil, err
		}
		block = *_block
	} else {
		_block, err := bc.GetBlock([]byte(params.BlockHash))
		if err != nil {
			return nil, err
		}
		block = _block
	}
	var txp *blockchain.Transaction
	for _, _tx := range block.Transactions {
		if bytes.Equal(_tx.ID, []byte(params.Txid)) {
			txp = _tx
			break
		}
	}
	if txp == nil {
		return nil, NewRPCError(ErrRPCNoTxInfo, "Transaction Not FOUND!!")
	}

	// tx, err := bc.FindTransaction([]byte(params.Txid))
	tx := *txp
	res := TxRawResult{}
	res.Hex = string(tx.ID)
	res.Txid = string(tx.ID)
	res.Hash = string(tx.ID)
	res.Size = 128
	res.Vsize = res.Size
	res.Weight = res.Vsize * 4
	res.Version = 100000000
	res.LockTime = 0
	res.Vin = tx.Vin
	res.Vout = tx.Vout
	res.BlockHash = hex.EncodeToString(block.Hash)
	res.Confirmations = uint64(bestH) - uint64(block.Height)
	res.Time = 0
	res.Blocktime = block.Timestamp
	if !params.Verbose {

		return res, nil
	} else {
		return DataToHex(res)
	}

}

func handleGetrawmempool(ctx context.Context, rawparams *Params) (interface{}, error) {
	return []string{}, nil
}

func handleGetnetworkhashPs(ctx context.Context, rawparams *Params) (interface{}, error) {
	return float64(1.5), nil

}

func handleGetnetworkinfo(ctx context.Context) (interface{}, error) {
	res := GetNetworkInfoRes{}

	res.Version = 100000000
	res.Connections = 5
	res.Connections_in = 5
	res.Connections_out = 5
	res.IncrementalFee = 0
	res.Networks = struct {
		Name                   string `json:"name"`
		Limited                bool   `json:"limited"`
		Reachable              bool   `json:"reachable"`
		Proxy                  string `json:"proxy"`
		ProxyRandomCredentials bool   `json:"proxy_randomize_credentials"`
	}{Name: "onion", Limited: false, Reachable: true, ProxyRandomCredentials: true}
	res.LocalAddresses = struct {
		Address string `json:"address"`
		Port    uint   `json:"port"`
		Score   uint   `json:"score"`
	}{Address: "0.0.0.0", Port: 8332, Score: 1}
	res.LocalRelay = false
	res.NetworkActive = true
	res.ProtocolVersion = 100100000
	res.RelayFee = 0
	res.Subversion = 100
	res.TimeOffset = 1660554000
	return res, nil
}

func handleGetmempoolinfo(ctx context.Context) (interface{}, error) {
	res := GetMempoolInfoRes{}
	res.Loaded = true
	res.Size = 0
	res.Bytes = 0
	res.Usage = 1
	res.MaxMempool = 0
	res.MinRelayTXFee = 9999999
	res.UnbroadcastCount = 0
	return res, nil
}

func handleGetmempoolentry(ctx context.Context, rawparams *Params) (interface{}, error) {
	return nil, nil
}

// func handleGetinfo(ctx context.Context, rawparams *Params) (interface{}, error) {

// 	bc := ctx.Value(Bckey).(*blockchain.Blockchain)
// 	best, err := bc.GetLastBlock()
// 	if err != nil {
// 		return nil, err
// 	}
// 	res := InfoChainResult{}
// 	res.Version = 100000000
// 	res.Blocks = int32(best.Height)
// 	res.TimeOffset = 1660554000
// 	res.ProtocolVersion = 100100000
// 	res.Connections = 0
// 	res.Proxy = ""
// 	res.Difficulty = 2 << 16
// 	res.TestNet = false
// 	res.RelayFee = 0

// 	return res, nil

// }

func handleGetdifficulty(ctx context.Context) (interface{}, error) {
	return 2 << 16, nil
}

func handleGetchaintips(ctx context.Context) (interface{}, error) {
	bc := ctx.Value(Bckey).(*blockchain.Blockchain)
	best, err := bc.GetLastBlock()
	if err != nil {
		return nil, err
	}
	res := struct {
		Height    uint   `json:"height"`
		Hash      Hex    `json:"hash"`
		BranchLen int    `json:"branchlen"`
		Status    string `json:"status"`
	}{
		Height:    uint(best.Height),
		Hash:      Hex(hex.EncodeToString(best.Hash)),
		BranchLen: 0,
		Status:    "active",
	}
	return res, nil
}

func handleGetblockheader(ctx context.Context, rawparams *Params) (interface{}, error) {
	params := GetBlockHeader{Verbose: true}
	err := CastParamsTo(rawparams, &params, 1, 2)
	if err != nil {
		return nil, err
	}
	bc := ctx.Value(Bckey).(*blockchain.Blockchain)
	// var bc = blockchain.NewBlockchainView()
	best, err := bc.GetLastBlock()

	if err != nil {
		return nil, err
	}
	var blocks = blockchain.Blockchain{DB: bc.DB}

	blkhashDec, err := HashToBytes(params.Blockhash)
	if err != nil {
		return nil, NewRPCError(ErrRPCInvalidAddressOrKey, err.Error())
	}
	blockdata, err := blocks.GetBlock(blkhashDec)
	if err != nil {
		return nil, NewRPCError(ErrRPCInternal.Code, err.Error())
	}
	blockdata.Transactions = nil
	tmp := 0
	if params.Verbose {
		tmp = 1
	}
	result, err := BlockDataToResp(blockdata, tmp)

	result.Confirmations = best.Height - blockdata.Height
	if err != nil {
		return nil, err
	}

	if !params.Verbose {
		return DataToHex(result)
	} else {
		return result, nil
	}

}

func handleDecoderawtransaction(ctx context.Context, rawparams *Params) (interface{}, error) {
	params := DecodeRawTransactionCmd{}
	err := CastParamsTo(rawparams, &params, 1, 1)
	if err != nil {
		return nil, err
	}
	hexStr := params.HexTx
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	var tx blockchain.Transaction
	err = HexUnmarshal(hexStr, &tx)

	if err != nil {
		return nil, NewRPCError(ErrRPCDecodeHexString, err.Error())
	}
	res := TxRawResult{}
	res.Hex = string(tx.ID)
	res.Txid = string(tx.ID)
	res.Hash = string(tx.ID)
	res.Size = 128
	res.Vsize = res.Size
	res.Weight = res.Vsize * 4
	res.Version = 100000000
	res.LockTime = 0
	res.Vin = tx.Vin
	res.Vout = tx.Vout
	res.Time = 0
	return res, nil
}

func handleGetblockstats(ctx context.Context, rawparams *Params) (interface{}, error) {

	bc := ctx.Value(Bckey).(*blockchain.Blockchain)
	block := blockchain.Block{}
	params := (*rawparams)[0]
	var err error
	if reflect.TypeOf(params).Kind() == reflect.Float64 {
		block, err = bc.GetBlockByNumber(int(params.(float64)))
		if err != nil {
			return nil, err
		}
	} else {
		block, err = bc.GetBlock([]byte(params.(string)))
		if err != nil {
			return nil, err
		}
	}

	fees := []float64{}
	txsizes := []float64{}
	ins := 0
	outs := 0
	for _, tx := range block.Transactions {
		amount_inp := 0
		for _, txinp := range tx.Vin {
			amount_inp += txinp.Vout
			ins++
		}

		amount_out := 0
		for _, txout := range tx.Vout {
			amount_out += txout.Value
			outs++
		}
		fees = append(fees, float64(amount_inp)-float64(amount_out))
		hexTx, err := DataToHex(tx)
		if err != nil {
			return nil, err
		}
		txsizes = append(txsizes, float64(1.0*len(hexTx)/2))
	}

	res := BlockStatsResult{}

	_avgfee, _ := stats.Mean(stats.LoadRawData(fees))
	res.AvgFee = _avgfee

	res.AvgFeeRate = res.AvgFee / 128
	res.AvgTxSize, _ = stats.Mean(txsizes)

	res.BlockHash = Hex(hex.EncodeToString(block.Hash))
	res.FeeratePercentiles = append(res.FeeratePercentiles, res.AvgFeeRate, res.AvgFeeRate*2, res.AvgFeeRate*4, res.AvgFeeRate*8, res.AvgFeeRate*16)
	res.Height = uint(block.Height)
	res.Ins = uint(ins) - 1

	_maxfee, _ := stats.Max(stats.LoadRawData(fees))
	res.Maxfee = _maxfee

	res.MaxFeeRate = res.Maxfee / 128
	res.MaxTxSize, _ = stats.Max(txsizes)

	_minfee, _ := stats.Min(stats.LoadRawData(fees))
	res.MinFee = _minfee

	res.MinFeeRate = res.MinFee / 128
	res.MinTxSize, _ = stats.Min(txsizes)

	_medianfee, _ := stats.Median(stats.LoadRawData(fees))
	res.MedianFee = _medianfee

	res.MedianTxSize, _ = stats.Median(txsizes)

	res.MedianTime = 30
	res.Outs = uint(outs)
	res.Time = uint(block.Timestamp)
	_sum, _ := stats.Sum(stats.LoadRawData(fees))
	res.TotalFee = _sum
	res.TotalSize, _ = stats.Sum(txsizes)
	res.TotalOut = res.TotalFee
	res.Txs = uint(len(txsizes))
	res.UTXOIncrease = int(res.Outs - res.Ins)
	res.UTXOSizeInc = 0
	fmt.Printf("EVERYTHING: %+v", res)
	return res, nil

}

func handleListtransactions(ctx context.Context, rawparams *Params) (interface{}, error) {
	bc := ctx.Value(Bckey).(*blockchain.Blockchain)

	best := bc.GetBestHeight()
	utxoSets := bc.FindUTXO()
	res := []TransactionData{}
	for txid := range utxoSets {
		block, err := FindBlockHasTx(bc, txid)
		if err != nil {
			return nil, err
		}
		amount_inp := 0
		amount_out := 0
		index := -1
		for i, tx := range block.Transactions {
			for _, txinp := range tx.Vin {
				amount_inp += txinp.Vout
			}

			for _, txout := range tx.Vout {
				amount_out += txout.Value
			}
			if bytes.Equal(tx.ID, []byte(txid)) {
				index = i
			}

		}
		res = append(res, TransactionData{
			Amount:             amount_out,
			Fee:                amount_inp - amount_out,
			Confirmations:      best - block.Height,
			Generated:          true,
			Trusted:            true,
			BlockHash:          Hex(hex.EncodeToString(block.Hash)),
			BlockHeight:        block.Height,
			BlockIndex:         index,
			TxID:               Hex(txid),
			WalletConflicts:    []string{},
			Time:               block.Timestamp,
			TimeReceived:       block.Timestamp,
			BIP125_Replaceable: "no",
		})
	}
	return res, nil
}

func handleGetblockchaininfo(ctx context.Context) (interface{}, error) {
	bc := ctx.Value(Bckey).(*blockchain.Blockchain)

	best, err := bc.GetLastBlock()
	if err != nil {
		return nil, err
	}
	filep, err := filepath.Abs("./blockchain.db")
	if err != nil {
		return nil, err
	}
	dbfile, err := os.Stat(filep)
	if err != nil {
		return nil, err
	}

	res := GetBlockChainInfoResult{
		Chain:                "main",
		Blocks:               int32(best.Height),
		Headers:              int32(best.Height),
		BestBlockHash:        string(best.Hash),
		Difficulty:           2 << 16,
		MedianTime:           5,
		VerificationProgress: 1,
		InitialBlockDownload: false,
		ChainWork:            "0x01",
		SizeOnDisk:           dbfile.Size(),
		Pruned:               false,
		SoftForks:            nil,
	}

	return res, nil
}
