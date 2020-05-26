package grc

import (
	"context"
	"testing"
	"time"

	"github.com/appootb/grc/backend/memory"
)

type TestConfig struct {
	IV Int   `default:"123" comment:"test"`
	AV Slice `default:"aa,bb" comment:"xx"`
	ES struct {
		MV Map `default:"a:1,b:2" comment:"mmmm"`
	}
	EIV int                          `default:"1" comment:"internal int"`
	ESV map[string]map[string]string `default:"cc:aa:ee,xx:yy;dd" comment:"internal slice"`
}

func Test_NewWithProvider(t *testing.T) {
	provider, err := memory.NewProvider()
	if err != nil {
		t.Fatal(err)
	}
	cfg := TestConfig{}
	rc, err := NewWithProvider(context.TODO(), provider, "/base_path")
	if err != nil {
		t.Fatal(err)
	}
	cfg.AV.Changed(func() {
		t.Log("av cfg changed", cfg.AV.Strings())
	})
	err = rc.RegisterConfig("test", &cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("done", cfg)
	time.Sleep(time.Second)
}
