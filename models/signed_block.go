package models

import (
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
)

type MySignedBlock struct {
	Block         MyBlock             `json:"block"`
	Justification types.Justification `json:"justification"`
}

// Block encoded with header and extrinsics
type MyBlock struct {
	Header     types.Header
	Extrinsics []MyExtrinsic
}
