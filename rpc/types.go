package rpc

const (
	BLOCK_HEADER_SIZE = 256
)

type reqKey int

const Bckey reqKey = 5

type Vin struct {
	Coinbase  string    `json:"coinbase"`
	Txid      string    `json:"txid"`
	Vout      uint32    `json:"vout"`
	ScriptSig ScriptSig `json:"scriptSig"`
	Sequence  uint32    `json:"sequence"`
	Witness   []string  `json:"txinwitness"`
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

type Vout struct {
	Value        float64            `json:"value"`
	N            uint32             `json:"n"`
	ScriptPubKey ScriptPubKeyResult `json:"scriptPubKey"`
}

type ScriptPubKeyResult struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex,omitempty"`
	ReqSigs   int32    `json:"reqSigs,omitempty"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses,omitempty"`
}

type GetBlock struct {
	Blockhash string  `json:"blockhash"`
	Verbose   float64 `json:"verbosity,omitempty"`
}

type GetBlockHeader struct {
	Blockhash string `json:"blockhash"`
	Verbose   bool   `json:"verbosity,omitempty"`
}

type SendRawTransaction struct {
	Data       string  `json:"hexstring"`
	MaxFeeRate float32 `json:"max_fee_rate,omitempty"`
}

type CreateRawTransactionCmd struct {
	Inputs   []RawTransactionInput
	Amounts  map[string]float64 `jsonrpcusage:"{\"address\":amount,...}"` // In BTC
	LockTime *int64
}

type RawTransactionInput struct {
	Txid string `json:"txid"`
	Vout int    `json:"vout"`
}

type GetTxOut struct {
	TxId           string
	Vout           float64
	IncludeMempool bool
}

type GetTransaction struct {
	TxId              string
	Include_WatchOnly bool
	Verbose           bool
}

type GetTxoutResp struct {
	BestBlock     Hex     `json:"bestblock"`
	Confirmations int64   `json:"confirmations"`
	Value         float64 `json:"value"`
	ScriptPubKey  string  `json:"scriptPubKey"`
	Coinbase      bool    `json:"coinbase"`
}

type GetRawTransactionCmd struct {
	Txid      string
	Verbose   bool `jsonrpcdefault:"false"`
	BlockHash string
}

// Response
type Hex string

type Params []interface{}

type TransactionData struct {
	Amount             int         `json:"amount"`
	Fee                int         `json:"fee"`
	Confirmations      int         `json:"confirmations"`
	Generated          bool        `json:"generated"`
	Trusted            bool        `json:"trusted"`
	BlockHash          Hex         `json:"blockhash"`
	BlockHeight        int         `json:"blockheight"`
	BlockIndex         int         `json:"blockindex"`
	TxID               Hex         `json:"txid"`
	WalletConflicts    []string    `json:"walletconflicts"`
	Time               int64       `json:"time"`
	TimeReceived       int64       `json:"timereceived"`
	Comment            string      `json:"comment"`
	BIP125_Replaceable string      `json:"bip125_replaceable"`
	Details            interface{} `json:"details"`
	Hex                Hex         `json:"hex"`
	Decoded            interface{} `json:"decoded"`
}

type TransactionID string

type BlockDataResp struct {
	Hash          Hex         `json:"hash"`
	Confirmations int         `json:"confirmations"`
	Size          int         `json:"size"`
	StrippedSize  int         `json:"strippedsize"`
	Weight        int         `json:"weight"`
	Height        int         `json:"height"`
	Version       int         `json:"version"`
	VersionHex    Hex         `json:"versionHex"`
	MerkleRoot    Hex         `json:"merkleroot"`
	Timestamp     int64       `json:"time"`
	MedianTime    int64       `json:"mediantime"`
	Transactions  interface{} `json:"tx"`
	Nonce         int         `json:"nonce"`
	Bits          Hex         `json:"bits"`
	Difficulty    int         `json:"difficulty"`
	ChainWork     Hex         `json:"chainwork"`
	NumOfTxs      int         `json:"nTx"`
	PrevBlockHash Hex         `json:"previousblockhash"`
}

type ValidateAddressResp struct {
	IsValid        bool   `json:"isvalid"`
	Address        string `json:"address,omitempty"`
	IsScript       bool   `json:"isscript,omitempty"`
	IsWitness      bool   `json:"iswitness,omitempty"`
	WitnessVersion int32  `json:"witness_version,omitempty"`
	WitnessProgram string `json:"witness_program,omitempty"`
}

type GetNetworkInfoRes struct {
	Version            float32     `json:"version"`
	Subversion         float32     `json:"subversion"`
	ProtocolVersion    float32     `json:"protocolversion"`
	LocalServices      Hex         `json:"localservices"`
	LocalServicesNames []string    `json:"localservicesnames"`
	LocalRelay         bool        `json:"localrelay"`
	TimeOffset         uint        `json:"timeoffset"`
	Connections        uint        `json:"connections"`
	Connections_in     uint        `json:"connections_in"`
	Connections_out    uint        `json:"connections_out"`
	NetworkActive      bool        `json:"networkactive"`
	Networks           interface{} `json:"networks"`
	RelayFee           float32     `json:"relayfee"`
	IncrementalFee     float32     `json:"incrementalfee"`
	LocalAddresses     interface{} `json:"localaddresses"`
	Warnings           string      `json:"warnings"`
}

type GetMempoolInfoRes struct {
	Loaded           bool    `json:"loaded"`
	Size             uint    `json:"size"`
	Bytes            uint    `json:"bytes"`
	Usage            uint    `json:"usage"`
	MaxMempool       uint    `json:"maxmempool"`
	MempoolMinFee    float32 `json:"mempoolminfee"`
	MinRelayTXFee    float32 `json:"minrelaytxfee"`
	UnbroadcastCount uint    `json:"unbroadcastcount"`
}

type TxRawResult struct {
	Hex           string `json:"hex"`
	Txid          string `json:"txid"`
	Hash          string `json:"hash,omitempty"`
	Size          int32  `json:"size,omitempty"`
	Vsize         int32  `json:"vsize,omitempty"`
	Weight        int32  `json:"weight,omitempty"`
	Version       uint32 `json:"version"`
	LockTime      uint32 `json:"locktime"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	BlockHash     string `json:"blockhash,omitempty"`
	Confirmations uint64 `json:"confirmations,omitempty"`
	Time          int64  `json:"time,omitempty"`
	Blocktime     int64  `json:"blocktime,omitempty"`
}

type InfoChainResult struct {
	Version         int32   `json:"version"`
	ProtocolVersion int32   `json:"protocolversion"`
	Blocks          int32   `json:"blocks"`
	TimeOffset      int64   `json:"timeoffset"`
	Connections     int32   `json:"connections"`
	Proxy           string  `json:"proxy"`
	Difficulty      float64 `json:"difficulty"`
	TestNet         bool    `json:"testnet"`
	RelayFee        float64 `json:"relayfee"`
	Errors          string  `json:"errors"`
}

type DecodeRawTransactionCmd struct {
	HexTx string
}

type BlockStatsResult struct {
	AvgFee             float64   `json:"avgfee"`
	AvgFeeRate         float64   `json:"avgfeerate"`
	AvgTxSize          float64   `json:"avgtxsize"`
	BlockHash          Hex       `json:"blockhash"`
	FeeratePercentiles []float64 `json:"feerate_percentiles"`
	Height             uint      `json:"height"`
	Ins                uint      `json:"ins"`
	Maxfee             float64   `json:"maxfee"`
	MaxFeeRate         float64   `json:"maxfeerate"`
	MaxTxSize          float64   `json:"maxtxsize"`
	MedianFee          float64   `json:"medianfee"`
	MedianTime         float64   `json:"mediantime"`
	MedianTxSize       float64   `json:"mediantxsize"`
	MinFee             float64   `json:"minfee"`
	MinFeeRate         float64   `json:"minfeerate"`
	MinTxSize          float64   `json:"mintxsize"`
	Outs               uint      `json:"outs"`
	Subsidy            float64   `json:"subsidy"`
	SWTotalSize        uint      `json:"swtotal_size"`
	SWTotalWeight      uint      `json:"swtotal_weight"`
	SWTxs              uint      `json:"swtxs"`
	Time               uint      `json:"time"`
	TotalOut           float64   `json:"total_out"`
	TotalSize          float64   `json:"total_size"`
	TotalWeight        uint      `json:"total_weight"`
	TotalFee           float64   `json:"totalfee"`
	Txs                uint      `json:"txs"`
	UTXOIncrease       int       `json:"utxo_increase"`
	UTXOSizeInc        int       `json:"utxo_size_inc"`
}

type GetBlockChainInfoResult struct {
	Chain                string      `json:"chain"`
	Blocks               int32       `json:"blocks"`
	Headers              int32       `json:"headers"`
	BestBlockHash        string      `json:"bestblockhash"`
	Difficulty           float64     `json:"difficulty"`
	MedianTime           int64       `json:"mediantime"`
	VerificationProgress float64     `json:"verificationprogress,omitempty"`
	InitialBlockDownload bool        `json:"initialblockdownload,omitempty"`
	Pruned               bool        `json:"pruned"`
	PruneHeight          int32       `json:"pruneheight,omitempty"`
	AutomaticPruning     bool        `json:"automatic_pruning,omitempty"`
	ChainWork            string      `json:"chainwork,omitempty"`
	SizeOnDisk           int64       `json:"size_on_disk,omitempty"`
	SoftForks            interface{} `json:"softforks"`
}

type FeeResults struct {
	Base       float64 `json:"base"`
	Modified   float64 `json:"modified"`
	Ancestor   float64 `json:"ancestor"`
	Descendant float64 `json:"descendant"`
}

type MempoolEntryResult struct {
	Vsize             int        `json:"vsize" `
	Weight            int        `json:"weight"`
	Fee               float64    `json:"fee"`
	ModifiedFee       float64    `json:"modifiedfee"`
	Time              int        `json:"time"`
	Height            int        `json:"height"`
	Descendantcount   int        `json:"descendantcount"`
	Descendantsize    int        `json:"descendantsize"`
	Descendantfees    float64    `json:"descendantfees"`
	Ancestorcount     int        `json:"ancestorcount"`
	Ancestorsize      int        `json:"ancestorsize"`
	Ancestorfees      float64    `json:"ancestorfees"`
	Wtxid             string     `json:"wtxid"`
	Fees              FeeResults `json:"fees"`
	Depends           []string   `json:"depends"`
	SpentBy           []string   `json:"spentby"`
	Bip125Replaceable bool       `json:"bip125-replaceable"`
	Unbroadcast       bool       `json:"unbroadcast"`
}
