package models

import (
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
)

type MyExtrinsicSignatureV4 struct {
	Signer    MultiAddress
	Signature types.MultiSignature
	Era       types.ExtrinsicEra // extra via system::CheckEra
	Nonce     types.UCompact     // extra via system::CheckNonce (Compact<Index> where Index is u32))
	Tip       types.UCompact     // extra via balances::TakeFees (Compact<Balance> where Balance is u128))
}
