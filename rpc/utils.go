package rpc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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
	data := "0x" + hex.EncodeToString(rawData)

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
