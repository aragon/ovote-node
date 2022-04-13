package eth

import (
	"encoding/hex"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/vocdoni/arbo"
)

func TestParseEventNewProcess(t *testing.T) {
	c := qt.New(t)
	// log bytes generated from the contract newProcess event.
	// creator, id: 1, txHash: 0, censusRoot: 1111, censusSize: 1000,
	// resPubStartBlock: 6697316, resPubWindow: 1000, minParticipation: 0,
	// minPositiveVotes: 60
	dHex := "000000000000000000000000a6a2e217af2f983ee55a6e2195c1763a9420f8ad" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000457" +
		"00000000000000000000000000000000000000000000000000000000000003e8" +
		"0000000000000000000000000000000000000000000000000000000000663164" +
		"00000000000000000000000000000000000000000000000000000000000003e8" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"000000000000000000000000000000000000000000000000000000000000003c"
	d, err := hex.DecodeString(dHex)
	c.Assert(err, qt.IsNil)

	e, err := parseEventNewProcess(d)
	c.Assert(err, qt.IsNil)

	c.Assert(e.Creator.String(), qt.Equals,
		"0xa6a2E217aF2f983ee55A6e2195C1763a9420f8ad")
	c.Assert(e.ProcessID, qt.Equals, uint64(1))
	c.Assert(hex.EncodeToString(e.TxHash[:]), qt.Equals,
		"0000000000000000000000000000000000000000000000000000000000000000")
	c.Assert(hex.EncodeToString(e.CensusRoot[:]), qt.Equals,
		"5704000000000000000000000000000000000000000000000000000000000000")
	c.Assert(e.CensusSize, qt.Equals, uint64(1000))
	c.Assert(e.ResPubStartBlock, qt.Equals, uint64(6697316))
	c.Assert(e.ResPubWindow, qt.Equals, uint64(1000))
	c.Assert(e.MinParticipation, qt.Equals, uint8(0))
	c.Assert(e.MinPositiveVotes, qt.Equals, uint8(60))

	// log bytes generated from the contract newProcess event.
	// creator, id: 2, txHash:
	// 0x7e1975a6bf513022a8cc382a3cdb1e1dbcd58ebb1cb9abf11e64aadb21262516,
	// censusRoot:
	// 3997482243935470019154908634129466064231369626981967795243271053776626526277,
	// censusSize: 12345, resPubStartBlock: 6697316, resPubWindow: 2500,
	// minParticipation: 30, minPositiveVotes: 50
	dHex = "000000000000000000000000a6a2e217af2f983ee55a6e2195c1763a9420f8ad" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"7e1975a6bf513022a8cc382a3cdb1e1dbcd58ebb1cb9abf11e64aadb21262516" +
		"08d67ea943c2daebe8b75017e7019efe37891ed6b67dd79b7a245aa634a62845" +
		"0000000000000000000000000000000000000000000000000000000000003039" +
		"0000000000000000000000000000000000000000000000000000000000663164" +
		"00000000000000000000000000000000000000000000000000000000000009c4" +
		"000000000000000000000000000000000000000000000000000000000000001e" +
		"0000000000000000000000000000000000000000000000000000000000000032"
	d, err = hex.DecodeString(dHex)
	c.Assert(err, qt.IsNil)

	e, err = parseEventNewProcess(d)
	c.Assert(err, qt.IsNil)

	c.Assert(e.Creator.String(), qt.Equals,
		"0xa6a2E217aF2f983ee55A6e2195C1763a9420f8ad")
	c.Assert(e.ProcessID, qt.Equals, uint64(2))
	c.Assert(hex.EncodeToString(e.TxHash[:]), qt.Equals,
		"7e1975a6bf513022a8cc382a3cdb1e1dbcd58ebb1cb9abf11e64aadb21262516")
	c.Assert(arbo.BytesToBigInt(e.CensusRoot[:]).String(), qt.Equals,
		"3997482243935470019154908634129466064231369626981967795243271053776626526277")
	c.Assert(e.CensusSize, qt.Equals, uint64(12345))
	c.Assert(e.ResPubStartBlock, qt.Equals, uint64(6697316))
	c.Assert(e.ResPubWindow, qt.Equals, uint64(2500))
	c.Assert(e.MinParticipation, qt.Equals, uint8(30))
	c.Assert(e.MinPositiveVotes, qt.Equals, uint8(50))

	// log bytes generated from the contract newProcess event.
	dHex = "000000000000000000000000a6a2e217af2f983ee55a6e2195c1763a9420f8ad" +
		"0000000000000000000000000000000000000000000000000000000000000006" +
		"0000000000000000000000000000000000000000000000000000000000003039" +
		"08d67ea943c2daebe8b75017e7019efe37891ed6b67dd79b7a245aa634a62845" +
		"00000000000000000000000000000000000000000000000000000000000003e8" +
		"00000000000000000000000000000000000000000000000000000000006646b7" +
		"000000000000000000000000000000000000000000000000000000000000000a" +
		"000000000000000000000000000000000000000000000000000000000000000a" +
		"000000000000000000000000000000000000000000000000000000000000003c"

	d, err = hex.DecodeString(dHex)
	c.Assert(err, qt.IsNil)

	e, err = parseEventNewProcess(d)
	c.Assert(err, qt.IsNil)

	c.Assert(e.Creator.String(), qt.Equals,
		"0xa6a2E217aF2f983ee55A6e2195C1763a9420f8ad")
	c.Assert(e.ProcessID, qt.Equals, uint64(6))
	c.Assert(hex.EncodeToString(e.TxHash[:]), qt.Equals,
		"0000000000000000000000000000000000000000000000000000000000003039")
	c.Assert(arbo.BytesToBigInt(e.CensusRoot[:]).String(), qt.Equals,
		"3997482243935470019154908634129466064231369626981967795243271053776626526277")
	c.Assert(e.CensusSize, qt.Equals, uint64(1000))
	c.Assert(e.ResPubStartBlock, qt.Equals, uint64(6702775))
	c.Assert(e.ResPubWindow, qt.Equals, uint64(10))
	c.Assert(e.MinParticipation, qt.Equals, uint8(10))
	c.Assert(e.MinPositiveVotes, qt.Equals, uint8(60))
}
