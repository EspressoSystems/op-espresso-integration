package espresso

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func removeWhitespace(s string) string {
	// Split the string on whitespace then concatenate the segments
	return strings.Join(strings.Fields(s), "")
}

// Reference data taken from the reference sequencer implementation
// (https://github.com/EspressoSystems/espresso-sequencer/blob/main/data)

var ReferenceNmtRoot NmtRoot = NmtRoot{
	Root: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var ReferenceL1BLockInfo L1BlockInfo = L1BlockInfo{
	Number:    123,
	Timestamp: *NewU256().SetUint64(0x456),
}

var ReferenceHeader Header = Header{
	Timestamp:        789,
	L1Block:          ReferenceL1BLockInfo,
	TransactionsRoot: ReferenceNmtRoot,
}

func TestEspressoTypesNmtRootJson(t *testing.T) {
	data := []byte(removeWhitespace(`{
		"root": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]
	}`))

	// Check encoding.
	encoded, err := json.Marshal(ReferenceNmtRoot)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	require.Equal(t, encoded, data)

	// Check decoding
	var decoded NmtRoot
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	require.Equal(t, decoded, ReferenceNmtRoot)
}

func TestEspressoTypesL1BLockInfoJson(t *testing.T) {
	data := []byte(removeWhitespace(`{
		"number": 123,
		"timestamp": "0x456"
	}`))

	// Check encoding.
	encoded, err := json.Marshal(ReferenceL1BLockInfo)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	require.Equal(t, encoded, data)

	// Check decoding
	var decoded L1BlockInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	require.Equal(t, decoded, ReferenceL1BLockInfo)
}

func TestEspressoTypesHeaderJson(t *testing.T) {
	data := []byte(removeWhitespace(`{
		"timestamp": 789,
		"l1_block": {
			"number": 123,
			"timestamp": "0x456"
		},
		"transactions_root": {
			"root": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]
		}
	}`))

	// Check encoding.
	encoded, err := json.Marshal(ReferenceHeader)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	require.Equal(t, encoded, data)

	// Check decoding
	var decoded Header
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	require.Equal(t, decoded, ReferenceHeader)
}

// Commitment tests ported from the reference sequencer implementation
// (https://github.com/EspressoSystems/espresso-sequencer/blob/main/sequencer/src/block.rs)

func TestEspressoTypesNmtRootCommit(t *testing.T) {
	require.Equal(t, ReferenceNmtRoot.Commit(), Commitment{251, 80, 232, 195, 91, 2, 138, 18, 240, 231, 31, 172, 54, 204, 90, 42, 215, 42, 72, 187, 15, 28, 128, 67, 149, 117, 26, 114, 232, 57, 190, 10})
}

func TestEspressoTypesL1BlockInfoCommit(t *testing.T) {
	require.Equal(t, ReferenceL1BLockInfo.Commit(), Commitment{139, 253, 167, 177, 129, 217, 11, 21, 181, 233, 68, 140, 237, 249, 195, 175, 40, 72, 77, 226, 12, 14, 75, 33, 137, 188, 230, 107, 186, 225, 150, 201})
}

func TestEspressoTypesHeaderCommit(t *testing.T) {
	require.Equal(t, ReferenceHeader.Commit(), Commitment{219, 19, 178, 0, 171, 225, 163, 206, 36, 8, 34, 84, 163, 96, 248, 52, 233, 85, 254, 89, 93, 249, 113, 142, 95, 204, 102, 103, 74, 250, 45, 103})
}
