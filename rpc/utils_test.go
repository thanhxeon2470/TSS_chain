package rpc

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCastParamsTo(t *testing.T) {
	prms := Params{"0xasjdkafawoijgiajrg", "0"}
	T := GetBlock{}

	CastParamsTo(&prms, &T, 2, 2)
	fmt.Printf("Ket que ne: %v", T)

	if reflect.TypeOf(T) != reflect.TypeOf(ErrRPCInvalidParams) {
		t.Log("xui qua di")
		t.Errorf("Loi nhu cc k dung params %v", T)
	}

}
