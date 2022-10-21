package rpc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
)

func HashToBytes(hash string) ([]byte, error) {
	if len(hash) == 0 {
		return nil, fmt.Errorf("EMPTY HASH")
	}
	hash = strings.TrimPrefix(hash, "0x")

	rawBytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, fmt.Errorf("CANT DECODE")
	}

	return rawBytes, nil
}

func BytesToHex(rawData []byte) (string, error) {
	if len(rawData) == 0 {
		return "", fmt.Errorf("EMPTY BYTES")
	}
	data := hex.EncodeToString(rawData)

	return data, nil
}
func StringToInt(numstr string) (int, error) {
	return strconv.Atoi(numstr)
}

func HextoBytes(hexData string) ([]byte, error) {
	if len(hexData)%2 != 0 {
		hexData = "0" + hexData
	}

	data, err := hex.DecodeString(hexData)
	return data, err
}

func CastParamsTo(rawparams *Params, T interface{}, min, max int) error {
	var valuePtr = reflect.ValueOf(T)
	var valueT reflect.Value
	if valuePtr.Kind() == reflect.Struct {
		fmt.Printf("nhu cc %+v", T)
		valueT = valuePtr
	}

	if valuePtr.Kind() == reflect.Ptr {
		valueT = valuePtr.Elem()
	}
	if !valueT.CanSet() {
		fmt.Println("KHONG SET set dc ")
	}

	// var numOfFields = valueT.NumField()

	var params = (*rawparams)
	if len(params) < min || len(params) > max {
		return ErrRPCInvalidParams
	}

	for i := 0; i < len(params); i++ {
		field := valueT.FieldByIndex([]int{i})
		field.Set(reflect.ValueOf(params[i]))
		// field.SetString(params[i].(string))

	}
	return nil
}

func DataToHex(T interface{}) (Hex, error) {
	respBuffer := new(bytes.Buffer)
	json.NewEncoder(respBuffer).Encode(T)
	respData, err := BytesToHex(respBuffer.Bytes())
	return Hex(respData), err
}

func HexUnmarshal(input string, T interface{}) error {
	rawInput, err := hex.DecodeString(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(rawInput, &T)

}

func DataToObject(data []byte, T interface{}) (interface{}, error) {
	err := json.Unmarshal(data, &T)
	return T, err
}

func Vin2Result(tx blockchain.Transaction, vin blockchain.TXInput) (*Vin, error) {
	sig, _ := BytesToHex(vin.Signature)
	coinbase := tx.IsCoinbase()
	res := Vin{
		Txid:      hex.EncodeToString(vin.Txid),
		Vout:      uint32(vin.Vout),
		Sequence:  0,
		Witness:   nil,
		ScriptSig: ScriptSig{Asm: "tss", Hex: sig},
		Coinbase:  fmt.Sprint(coinbase),
	}
	return &res, nil

}

func Vout2Result(tx blockchain.Transaction, vout blockchain.TXOutput) (*Vout, error) {

	hash, err := BytesToHex(vout.PubKeyHash)
	if err != nil {
		return nil, err
	}
	res := Vout{
		Value: float64(vout.Value),
		ScriptPubKey: ScriptPubKeyResult{
			Asm:     "TSS",
			Hex:     hash,
			ReqSigs: 1,
			Type:    "pubkeyhash",
		},
	}

	for i, _vout := range tx.Vout {
		if bytes.Equal(_vout.PubKeyHash, vout.PubKeyHash) {
			res.N = uint32(i)
			return &res, nil
		}
	}

	return nil, NewRPCError(ErrRPCInvalidTxVout, "cant find vout")
}
