package jsonx_test

import (
	"encoding/json"
	"testing"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client/jsonx"
)

func TestScalars(t *testing.T) {
	var v struct {
		A jsonx.Text     `json:"a"`
		B jsonx.Text     `json:"b"`
		C jsonx.Text     `json:"c"`
		D jsonx.Flag     `json:"d"`
		E jsonx.Flag     `json:"e"`
		F jsonx.Flag     `json:"f"`
		G jsonx.Flag     `json:"g"`
		H jsonx.Count    `json:"h"`
		I jsonx.Count    `json:"i"`
		J jsonx.TextList `json:"j"`
		K jsonx.TextList `json:"k"`
		L jsonx.TextList `json:"l"`
	}
	in := `{"a":"x","b":7,"c":null,"d":true,"e":"1","f":0,"g":null,"h":12,"i":"34","j":["p",2],"k":null,"l":""}`
	if err := json.Unmarshal([]byte(in), &v); err != nil {
		t.Fatal(err)
	}
	if v.A != "x" || v.B != "7" || v.C != "" || !v.D || !v.E || v.F || v.G || v.H != 12 || v.I != 34 {
		t.Fatalf("unexpected %+v", v)
	}
	if got := v.J.Strings(); len(got) != 2 || got[0] != "p" || got[1] != "2" {
		t.Fatalf("unexpected list %v", got)
	}
	if v.K.Strings() == nil || len(v.K) != 0 || len(v.L) != 0 {
		t.Fatalf("null and empty string should be empty lists: %v %v", v.K, v.L)
	}
}

func TestRejectsGarbage(t *testing.T) {
	var f jsonx.Flag
	if err := json.Unmarshal([]byte(`"maybe"`), &f); err == nil {
		t.Fatal("expected error for flag")
	}
	var c jsonx.Count
	if err := json.Unmarshal([]byte(`"many"`), &c); err == nil {
		t.Fatal("expected error for count")
	}
	var l jsonx.TextList
	if err := json.Unmarshal([]byte(`{"a":1}`), &l); err == nil {
		t.Fatal("expected error for list")
	}
}
