package main

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	subkey "github.com/vedhavyas/go-subkey"

	"github.com/skyvein-baas/client-skyvein-golang-api/handlers"
	"github.com/skyvein-baas/client-skyvein-golang-api/models"
)

func main() {
	cfg := models.Client{
		Addr: "ws://localhost:9944",
	}

	err := handlers.WatchHistoricalBlocks(cfg, &JustDisplay{})
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

type JustDisplay struct{}

func (jd *JustDisplay) HandleFunc(moduleName string, funcName string, argsByt []byte) (err error) {
	fmt.Println("[function]", moduleName, funcName)
	return
}

func (jd *JustDisplay) HandleEvent(meta *types.Metadata, rawEvents types.EventRecordsRaw) (err error) {
	events := MyEventRecords{}
	rawEvents.DecodeEventRecords(meta, &events)
	for _, e := range events.Balances_Endowed {
		fmt.Printf("\tBalances:Endowed:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%#x, %v\n", e.Who, e.Balance)
	}
	for _, e := range events.Balances_DustLost {
		fmt.Printf("\tBalances:DustLost:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%#x, %v\n", e.Who, e.Balance)
	}
	for _, e := range events.Balances_Transfer {
		fmt.Printf("\tBalances:Transfer:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%v, %v, %v\n", e.From, e.To, e.Value)
	}
	for _, e := range events.Balances_BalanceSet {
		fmt.Printf("\tBalances:BalanceSet:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%v, %v, %v\n", e.Who, e.Free, e.Reserved)
	}
	for _, e := range events.Balances_Deposit {
		fmt.Printf("\tBalances:Deposit:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%v, %v\n", e.Who, e.Balance)
	}
	for _, e := range events.Grandpa_NewAuthorities {
		fmt.Printf("\tGrandpa:NewAuthorities:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%v\n", e.NewAuthorities)
	}
	for _, e := range events.Grandpa_Paused {
		fmt.Printf("\tGrandpa:Paused:: (phase=%#v)\n", e.Phase)
	}
	for _, e := range events.Grandpa_Resumed {
		fmt.Printf("\tGrandpa:Resumed:: (phase=%#v)\n", e.Phase)
	}
	for _, e := range events.ImOnline_HeartbeatReceived {
		fmt.Printf("\tImOnline:HeartbeatReceived:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%#x\n", e.AuthorityID)
	}
	for _, e := range events.ImOnline_AllGood {
		fmt.Printf("\tImOnline:AllGood:: (phase=%#v)\n", e.Phase)
	}
	for _, e := range events.ImOnline_SomeOffline {
		fmt.Printf("\tImOnline:SomeOffline:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%v\n", e.IdentificationTuples)
	}
	for _, e := range events.Indices_IndexAssigned {
		fmt.Printf("\tIndices:IndexAssigned:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%#x%v\n", e.AccountID, e.AccountIndex)
	}
	for _, e := range events.Indices_IndexFreed {
		fmt.Printf("\tIndices:IndexFreed:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%v\n", e.AccountIndex)
	}
	for _, e := range events.Offences_Offence {
		fmt.Printf("\tOffences:Offence:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%v%v\n", e.Kind, e.OpaqueTimeSlot)
	}
	for _, e := range events.Session_NewSession {
		fmt.Printf("\tSession:NewSession:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%v\n", e.SessionIndex)
	}
	for _, e := range events.Staking_Reward {
		fmt.Printf("\tStaking:Reward:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%v\n", e.Amount)
	}
	for _, e := range events.Staking_Slash {
		fmt.Printf("\tStaking:Slash:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%#x%v\n", e.AccountID, e.Balance)
	}
	for _, e := range events.Staking_OldSlashingReportDiscarded {
		fmt.Printf("\tStaking:OldSlashingReportDiscarded:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%v\n", e.SessionIndex)
	}
	for _, e := range events.System_ExtrinsicSuccess {
		fmt.Printf("\tSystem:ExtrinsicSuccess:: (phase=%#v)\n", e.Phase)
	}
	for _, e := range events.System_ExtrinsicFailed {
		fmt.Printf("\tSystem:ExtrinsicFailed:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%v\n", e.DispatchError)
	}
	for _, e := range events.System_CodeUpdated {
		fmt.Printf("\tSystem:CodeUpdated:: (phase=%#v)\n", e.Phase)
	}
	for _, e := range events.System_NewAccount {
		fmt.Printf("\tSystem:NewAccount:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%#x\n", e.Who)
	}
	for _, e := range events.System_KilledAccount {
		fmt.Printf("\tSystem:KilledAccount:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%#X\n", e.Who)
	}
	for _, e := range events.TraceSource_ProductRegistered {
		fmt.Printf("\tTraceSource:ProductRegistered:: (phase=%#v)\n", e.Phase)
		who, _ := subkey.SS58Address(e.Who[:], 42)
		owner, _ := subkey.SS58Address(e.Owner[:], 42)
		fmt.Printf("\t\t%+v, %+v, %+v\n", who, e.Id, owner)
	}
	for _, e := range events.TraceSource_ShipmentRegistered {
		fmt.Printf("\tTraceSource:ShipmentRegistered:: (phase=%#v)\n", e.Phase)
		who, _ := subkey.SS58Address(e.Who[:], 42)
		owner, _ := subkey.SS58Address(e.Owner[:], 42)
		fmt.Printf("\t\t%+v, %+v, %+v\n", who, e.Id, owner)
	}
	for _, e := range events.TraceSource_ShipmentStatusUpdated {
		fmt.Printf("\tTraceSource:ShipmentStatusUpdated:: (phase=%#v)\n", e.Phase)
		fmt.Printf("\t\t%+v, %+v, %+v\n", e.Who[:], e.Id, e.ShipmentStatus)
	}
	for _, e := range events.TraceSource_ShipmentScanDone {
		fmt.Printf("\tTraceSource:ShipmentScanDone:: (phase=%#v)\n", e.Phase)
	}
	for _, e := range events.PacsDeposit_ReportRegistered {
		fmt.Printf("\tPacsDeposit:ReportRegistered:: (phase=%#v)\n", e.Phase)
		who, _ := subkey.SS58Address(e.Who[:], 42)
		com, _ := subkey.SS58Address(e.Com[:], 42)
		fmt.Printf("\t\t%+v, %+v, %+v\n", who, e.Id, com)
	}
	return
}

type MyEventRecords struct {
	types.EventRecords
	Feeless_FeelessDone               []models.EventFeelessFeelessDone
	TraceSource_ProductRegistered     []models.EventTraceSourceProductRegistered
	TraceSource_ShipmentRegistered    []models.EventTraceSourceShipmentRegistered
	TraceSource_ShipmentStatusUpdated []models.EventTraceSourceShipmentStatusUpdated
	TraceSource_ShipmentScanDone      []models.EventTraceSourceShipmentScanDone
	PacsDeposit_ReportRegistered      []models.EventPacsDepositReportRegistered
}
