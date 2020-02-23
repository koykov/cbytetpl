package cbytetpl

import (
	"bytes"
	"testing"
)

var (
	cutFmtOrigin = []byte(`
[
	{
		{%= val0 %},
		{%= val1 %}
	},
    {
        {%= val2 %},
        {%= val3 %}
    }
]
`)
	cutFmtExpect = []byte(`[{{%= val0 %},{%= val1 %}},{{%= val2 %},{%= val3 %}}]`)

	primHrExpect = []byte(`raw: [{
tpl: {%= val0 %}
raw: ,
tpl: {%= val1 %}
raw: },{
tpl: {%= val2 %}
raw: ,
tpl: {%= val3 %}
raw: }]`)

	uEOTOrigin = []byte(`foo {%= var1 %} bar {%= var1 end`)
)

func TestParse_CutComments(t *testing.T) {
	tpl := []byte(`{# this is a test template #}
		Payload line #0
		{# some comment #}
		Payload line #1
		{# EOT #}`)
	exp := []byte(`		Payload line #0
		Payload line #1
`)
	p := &Parser{tpl: tpl}
	p.cutComments()
	if !bytes.Equal(exp, p.tpl) {
		t.Errorf("comment cut test failed\nexp: %s\ngot: %s", string(exp), string(p.tpl))
	}
}

func TestParse_CutFmt(t *testing.T) {
	p := &Parser{tpl: cutFmtOrigin}
	p.cutFmt()
	if !bytes.Equal(cutFmtExpect, p.tpl) {
		t.Errorf("fmt cut test failed\nexp: %s\ngot: %s", string(cutFmtExpect), string(p.tpl))
	}
}

func TestParse_Primitive(t *testing.T) {
	tree, _ := Parse(cutFmtOrigin, false)
	r := tree.humanReadable()
	if bytes.Equal(r, primHrExpect) {
		t.Errorf("prim test failed\nexp: %s\ngot: %s", string(primHrExpect), string(r))
	}
}

func TestParse_uEOT(t *testing.T) {
	_, err := Parse(uEOTOrigin, false)
	if err != ErrUnexpectedEOF {
		t.Errorf("unexpected EOT fail\nexp: %s\ngot: %s", ErrUnexpectedEOF, err)
	}
}
