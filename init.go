package dyntpl

import "github.com/koykov/cbytealg"

func init() {
	RegisterModFn("default", "def", modDefault)
	RegisterModFn("ifThen", "if", modIfThen)
	RegisterModFn("ifThenElse", "ifel", modIfThenElse)

	RegisterModFn("jsonEscape", "je", modJsonEscape)
	RegisterModFn("jsonQuote", "jq", modJsonQuote)
	RegisterModFn("htmlEscape", "he", modHtmlEscape)

	RegisterCondFn("lenEq0", condLenEq0)
	RegisterCondFn("lenGt0", condLenGt0)
	RegisterCondFn("lenGtq0", condLenGtq0)

	cbytealg.RegisterAnyToBytesFn(ByteBufToBytes)
}
