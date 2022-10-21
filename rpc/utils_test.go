package rpc

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
)

func TestCastParamsTo(t *testing.T) {
	// prms := Params{"0xasjdkafawoijgiajrg", "0"}
	T := GetBlock{}

	// CastParamsTo(&prms, &T, 2, 2)
	fmt.Printf("Ket que ne: %v", T)

	if reflect.TypeOf(T) != reflect.TypeOf(ErrRPCInvalidParams) {
		t.Log("xui qua di")
		t.Errorf("Loi nhu cc k dung params %v", T)
	}

}

func TestVin2Result(t *testing.T) {
	tx := blockchain.Transaction{}
	vin := blockchain.TXInput{
		Txid:      []byte("abeffe0a4sa6ba4b6a7b8a5ba5a5b5ab4ab4a5ba6bca5bab4ab5"),
		Vout:      5,
		Signature: []byte("abeffe0a4sa6ba4b6a7b8a5ba5a5b5ab4ab4a5ba6bca5bab4ab5"),
		PubKey:    []byte("abeffe0a4sa6ba4b6a7b8a5ba5a5b5ab4ab4a5ba6bca5bab4ab5"),
	}
	res, err := Vin2Result(tx, vin)

	if err != nil {
		t.Errorf("Loi j v ne: %s", err)
	}
	fmt.Printf("akskaksks: %+v", res)
	t.Logf("thanh qua troi hay:%+v", res)
}

func TestVout2Result(t *testing.T) {
	tx := blockchain.Transaction{Vout: []blockchain.TXOutput{blockchain.TXOutput{Value: 10, PubKeyHash: []byte("abeffe0a4sa6ba4b6a7b8a5ba5a5b5ab4ab4a5ba6bca5bab4ab5")}}}
	vout := blockchain.TXOutput{
		Value:      5,
		PubKeyHash: []byte("abeffe0a4sa6ba4b6a7b8a5ba5a5b5ab4ab4a5ba6bca5bab4ab5"),
	}
	res, err := Vout2Result(tx, vout)

	if err != nil {
		t.Errorf("Loi j v ne: %s", err)
	}
	fmt.Printf("\n\n\n\n\nakskaksks: %+v", res)
	t.Logf("thanh qua troi hay:%+v", res)
}
