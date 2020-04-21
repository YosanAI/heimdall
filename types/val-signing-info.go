package types

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
)

// ValidatorSigningInfo defines the signing info for a validator
type ValidatorSigningInfo struct {
	Signer HeimdallAddress `json:"signer"`

	// height at which validator was first a candidate OR was unjailed
	StartHeight int64 `json:"startHeight"`
	// index offset into signed block bit array
	IndexOffset int64 `json:"indexOffset"`
	// timestamp validator cannot be unjailed until
	JailedUntil time.Time `json:"jailedUntil"`
	// whether or not a validator has been tombstoned (killed out of validator set)
	// Tombstoned bool `protobuf:"varint,5,opt,name=tombstoned,proto3" json:"tombstoned,omitempty"`
	// missed blocks counter (to avoid scanning the array every time)
	MissedBlocksCounter int64 `json:"missed_blocks_counter,omitempty"`
}

// NewValidatorSigningInfo creates a new ValidatorSigningInfo instance
func NewValidatorSigningInfo(
	condAddr HeimdallAddress, startHeight, indexOffset int64,
	jailedUntil time.Time, tombstoned bool, missedBlocksCounter int64,
) ValidatorSigningInfo {

	return ValidatorSigningInfo{
		Signer:      condAddr,
		StartHeight: startHeight,
		IndexOffset: indexOffset,
		JailedUntil: jailedUntil,
		// Tombstoned:          tombstoned,
		MissedBlocksCounter: missedBlocksCounter,
	}
}

// String implements the stringer interface for ValidatorSigningInfo
func (i ValidatorSigningInfo) String() string {
	return fmt.Sprintf(`Validator Signing Info:
  Address:               %s
  Start Height:          %d
  Index Offset:          %d
  Jailed Until:          %v
  Missed Blocks Counter: %d`,
		i.Signer, i.StartHeight, i.IndexOffset, i.JailedUntil,
		i.MissedBlocksCounter)
}

// amino marshall validator
func MarshallValSigningInfo(cdc *codec.Codec, valSigningInfo ValidatorSigningInfo) (bz []byte, err error) {
	bz, err = cdc.MarshalBinaryBare(valSigningInfo)
	if err != nil {
		return bz, err
	}
	return bz, nil
}

// amono unmarshall validator
func UnmarshallValSigningInfo(cdc *codec.Codec, value []byte) (ValidatorSigningInfo, error) {
	var valSigningInfo ValidatorSigningInfo
	// unmarshall validator and return
	err := cdc.UnmarshalBinaryBare(value, &valSigningInfo)
	if err != nil {
		return valSigningInfo, err
	}
	return valSigningInfo, nil
}
