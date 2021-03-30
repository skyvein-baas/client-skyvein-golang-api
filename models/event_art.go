package models

import (
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
)

type ArtEventRecords struct {
	types.EventRecords
	Feeless_FeelessDone []EventFeelessFeelessDone
	Art_ArtRegistered   []EventArtArtRegistered
	Art_ApproveForAll   []EventArtApproveForAll
	Art_ApproveFor      []EventArtApproveFor
	Art_ArtRaise        []EventArtArtRaise
	Art_NfcAdd          []EventArtNfcAdd
	Art_NfcBind         []EventArtNfcBind
}

type EventArtArtRegistered struct {
	Phase  types.Phase
	Who    types.AccountID
	Owner  types.AccountID
	Id     string
	Topics []types.Hash
}

type EventArtApproveForAll struct {
	Phase    types.Phase
	Caller   types.AccountID
	To       types.AccountID
	Approved bool
	IsClear  bool
	Topics   []types.Hash
}

type EventArtArtRaise struct {
	Phase  types.Phase
	Who    types.AccountID
	ArtId  string
	Topics []types.Hash
}

type EventArtNfcAdd struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}

type EventArtNfcBind struct {
	Phase    types.Phase
	Who      types.AccountID
	EntityId string
	NfcId    string
	Topics   []types.Hash
}

type EventArtApproveFor struct {
	Phase  types.Phase
	Caller types.AccountID
	To     types.AccountID
	Topics []types.Hash
}
