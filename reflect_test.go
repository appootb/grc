package grc

import (
	"fmt"
	"reflect"
	"testing"
)

type Config struct {
	IV Int   `default:"123" comment:"test"`
	AV Slice `default:"aa,bb" comment:"xx"`
	ES struct {
		MV Map `default:"a:1,b:2" comment:"mmmm"`
	}
	EIV int        `default:"1" comment:"internal int"`
	ESV [][]string `default:"cc,dd" comment:"internal slice"`
}

func Test_ParseConfigItems(t *testing.T) {
	fmt.Println(parseConfig(reflect.TypeOf(&Config{}), ""))
}
