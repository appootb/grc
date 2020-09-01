package grc

import (
	"reflect"
	"testing"
	"time"

	"github.com/appootb/grc/backend"
)

var (
	grc *RemoteConfig
)

func TestMain(m *testing.M) {
	grc, _ = New(WithDebugProvider(),
		WithConfigAutoCreation(),
		WithBasePath("/test"))
	m.Run()
}

func Test_SystemType(t *testing.T) {
	type Config struct {
		IV  int            `default:"1"`
		PIV *int           `default:"10"`
		MV  map[string]int `default:"a:1,b:2"`
		AV  []bool         `default:"false,true"`
	}
	var cfg Config
	err := grc.RegisterConfig("Test_SystemType", &cfg)
	if err != nil {
		t.Fatal(err)
	}

	iv := 10
	expect := Config{
		IV:  1,
		PIV: &iv,
		MV:  map[string]int{"a": 1, "b": 2},
		AV:  []bool{false, true},
	}
	if !reflect.DeepEqual(cfg, expect) {
		t.Fatal("expect:", expect, "actual:", cfg)
	}
}

func Test_SystemType2(t *testing.T) {
	type Config struct {
		EMV map[string]map[string]int32 `default:"a_1:bb_2:1,cc_2:2;b_1:dd_2:19,ee_2:20"`
		EAV [][]string                  `default:"a_1,b_1,c_1;a_2,b_2,c_2"`
	}
	cfg := &Config{}
	err := grc.RegisterConfig("Test_SystemType2", &cfg)
	if err != nil {
		t.Fatal(err)
	}

	expect := &Config{
		EMV: map[string]map[string]int32{
			"a_1": {
				"bb_2": 1,
				"cc_2": 2,
			},
			"b_1": {
				"dd_2": 19,
				"ee_2": 20,
			},
		},
		EAV: [][]string{
			{"a_1", "b_1", "c_1"},
			{"a_2", "b_2", "c_2"},
		},
	}
	if !reflect.DeepEqual(cfg, expect) {
		t.Fatal("expect:", expect, "actual:", cfg)
	}
}

func Test_Struct_SystemType(t *testing.T) {
	type ES struct {
		UV  uint8                `default:"199"`
		MMV map[int]map[int]bool `default:"1:11,22;2:33:true"`
	}
	type Config struct {
		ES
	}
	cfg := Config{}
	err := grc.RegisterConfig("Test_Struct_SystemType", &cfg)
	if err != nil {
		t.Fatal(err)
	}

	expect := Config{
		ES: ES{
			UV: 199,
			MMV: map[int]map[int]bool{
				1: {
					11: false,
					22: false,
				},
				2: {
					33: true,
				},
			},
		},
	}
	if !reflect.DeepEqual(cfg, expect) {
		t.Fatal("expect:", expect, "actual:", cfg)
	}
}

func Test_BaseType(t *testing.T) {
	type Config struct {
		IV Int    `default:"10"`
		MV Map    `default:"a:1,b:2"`
		AV *Array `default:"11,22,33"`
	}
	cfg := &Config{}
	err := grc.RegisterConfig("Test_BaseType", cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.IV.Int8() != 10 {
		t.Fatal("expect IV:", 10, "actual IV:", cfg.IV.String())
	}
	if cfg.MV.Keys().Len() != 2 || cfg.MV.IntVal("a").Int() != 1 || cfg.MV.IntVal("b").Int() != 2 {
		t.Fatal("actual MV:", cfg.MV.String())
	}
	s := cfg.AV.Ints()
	if cfg.AV.Len() != 3 || s[0].Int() != 11 || s[1].Int() != 22 || s[2].Int() != 33 {
		t.Fatal("actual AV:", cfg.AV.String())
	}
}

func Test_Struct_BaseType(t *testing.T) {
	type ES struct {
		SV String `default:"aa"`
	}
	type Config struct {
		ES
	}
	cfg := Config{}
	err := grc.RegisterConfig("Test_Struct_BaseType", &cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.SV.String() != "aa" {
		t.Fatal("actual SV:", cfg.SV.String())
	}
}

func Test_BaseType_Update(t *testing.T) {
	type Config struct {
		SV String `default:"aa"`
	}
	cfg := Config{}
	evt := make(chan bool)
	cfg.SV.Changed(func() {
		if cfg.SV.String() != "aa" {
			evt <- true
		}
	})
	err := grc.RegisterConfig("Test_BaseType_Update", &cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.SV.String() != "aa" {
		t.Fatal("actual SV:", cfg.SV.String())
	}
	c := backend.ConfigItem{
		Type:  "grc.String",
		Value: "bb",
	}
	key := backend.ServiceConfigKey(grc.path, "Test_BaseType_Update")
	err = grc.provider.Set(key+"SV", c.String(), 0)
	if err != nil {
		t.Fatal(err)
	}

	<-evt
	if cfg.SV.String() != "bb" {
		t.Fatal("actual SV:", cfg.SV.String())
	}
}

func Test_BaseType_Callback(t *testing.T) {
	type Config struct {
		FV Float `default:"3.14"`
	}
	cfg := Config{}
	evt := make(chan bool)
	cfg.FV.Changed(func() {
		if cfg.FV.Float64() != 3.14 {
			evt <- true
		}
	})
	err := grc.RegisterConfig("Test_BaseType_Callback", &cfg)
	if err != nil {
		t.Fatal(err)
	}

	c := backend.ConfigItem{
		Type:  "grc.String",
		Value: "2.24",
	}
	key := backend.ServiceConfigKey(grc.path, "Test_BaseType_Callback")
	err = grc.provider.Set(key+"FV", c.String(), 0)
	if err != nil {
		t.Fatal(err)
	}

	<-evt
	if cfg.FV.Float64() != 2.24 {
		t.Fatal("actual FV:", cfg.FV.String())
	}
}

func Test_RegisterNode(t *testing.T) {
	id1, err := grc.RegisterNode("Test_RegisterNode", "node1", WithNodeTTL(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if id1 == 0 {
		t.Fatal(id1)
	}
	//
	id2, err := grc.RegisterNode("Test_RegisterNode", "node2", WithOpsConfig())
	if err != nil {
		t.Fatal(err)
	}
	if id2 == 0 || id1 == id2 {
		t.Fatal(id1, id2)
	}
	time.Sleep(time.Second)
	svc := grc.GetNodes("Test_RegisterNode")
	if _, ok := svc["node1"]; !ok {
		t.Fatal("no service node1")
	}
	if _, ok := svc["node2"]; !ok {
		t.Fatal("no service node2")
	}
}

func Test_ParseDuration(t *testing.T) {
	type Config struct {
		Duration time.Duration `default:"1h"`
	}
	cfg := Config{}
	err := grc.RegisterConfig("Test_ParseDuration", &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Duration != time.Hour {
		t.Fatal("actual:", cfg.Duration)
	}
}

type Time time.Time

func (t *Time) Set(v string) {
	dt, _ := time.Parse(time.RFC3339, v)
	*t = Time(dt)
}

func Test_CustomStaticType(t *testing.T) {
	type Config struct {
		TV Time `default:"2020-06-04T21:00:57-08:00"`
	}
	cfg := Config{}
	err := grc.RegisterConfig("Test_CustomStaticType", &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if time.Time(cfg.TV).Unix() != 1591333257 {
		t.Fatal("actual:", cfg.TV)
	}
}

type Embed struct {
	time.Time
}

func (t *Embed) Set(v string) {
	dt, _ := time.Parse("2006-01-02 15:04:05", v)
	t.Time = dt
}

func Test_CustomEmbedType(t *testing.T) {
	type Config struct {
		EV []Embed `default:"2020-06-01 10:00:00,2020-07-01 12:00:00"`
	}
	cfg := Config{}
	err := grc.RegisterConfig("Test_CustomEmbedType", &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.EV) != 2 || cfg.EV[0].Unix() != 1591005600 || cfg.EV[1].Unix() != 1593604800 {
		t.Fatal("actual", cfg.EV)
	}
}
