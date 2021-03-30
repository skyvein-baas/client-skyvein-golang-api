package models

import (
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
)

type TraceSourceEventRecords struct {
	types.EventRecords
	Feeless_FeelessDone               []EventFeelessFeelessDone
	TraceSource_ProductRegistered     []EventTraceSourceProductRegistered
	TraceSource_ShipmentRegistered    []EventTraceSourceShipmentRegistered
	TraceSource_ShipmentStatusUpdated []EventTraceSourceShipmentStatusUpdated
	TraceSource_ShipmentScanDone      []EventTraceSourceShipmentScanDone
}

type EventTraceSourceProductRegistered struct {
	Phase  types.Phase
	Who    types.AccountID
	Id     string
	Owner  types.AccountID
	Topics []types.Hash
}

type EventFeelessFeelessDone struct {
	Phase  types.Phase
	Result types.DispatchResult
	Who    types.AccountID
	Topics []types.Hash
}

type EventTraceSourceShipmentRegistered struct {
	Phase  types.Phase
	Who    types.AccountID
	Id     string
	Owner  types.AccountID
	Topics []types.Hash
}

type EventTraceSourceShipmentStatusUpdated struct {
	Phase          types.Phase
	Who            types.AccountID
	Id             string
	EventIdx       types.U128
	ShipmentStatus string
	Topics         []types.Hash
}

type EventTraceSourceShipmentScanDone struct {
	Phase  types.Phase
	Topics []types.Hash
}
