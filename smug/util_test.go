package smug

import (
	"testing"
)

func TestChunkSplit(t *testing.T) {

	inp := "ABCDEF"
	outp := ChunkSplit(inp, 2)
	if len(outp) != 3 {
		t.Errorf("expected 3 parts")
	}
	if outp[0] != "AB" || outp[2] != "EF" {
		t.Errorf("err: output bogus")
	}

}
