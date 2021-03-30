package models

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v2/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
)

type RegisterReportReq struct {
	ID    string       // report.id
	Props []ReportProp // report属性
}

type RegisterReportRst struct {
	Who []byte // 操作人
	Com []byte // 企业账户
	Id  string // report.id
}

type OptionReportProps struct {
	HasValue bool
	Value    []ReportProp
}

type ReportProp struct {
	Name  string
	Value string
}

func (m OptionReportProps) Encode(encoder scale.Encoder) (err error) {
	if !m.HasValue {
		err = encoder.PushByte(0)
		if err != nil {
			return
		}
	} else {
		err = encoder.PushByte(1)
		if err != nil {
			return
		}
		err = encoder.Encode(m.Value)
		if err != nil {
			return
		}
	}
	return
}

func (m OptionReportProps) Decode(decoder scale.Decoder) (err error) {
	b, _ := decoder.ReadOneByte()
	switch b {
	case 0:
		m.HasValue = false
	case 1:
		m.HasValue = true
		err = decoder.Decode(&m.Value)
		if err != nil {
			return
		}
	default:
		return fmt.Errorf("Unknown byte prefix for encoded OptionReportProps: %d", b)
	}
	return
}

func (m ReportProp) Encode(encoder scale.Encoder) (err error) {
	err = encoder.Encode(m.Name)
	if err != nil {
		return
	}

	err = encoder.Encode(m.Value)
	if err != nil {
		return
	}
	return
}

func (m ReportProp) Decode(decoder scale.Decoder) (err error) {
	err = decoder.Decode(&m.Name)
	if err != nil {
		return
	}

	err = decoder.Decode(&m.Value)
	if err != nil {
		return
	}
	return
}

type PacsDepositEventRecords struct {
	types.EventRecords
	Feeless_FeelessDone          []EventFeelessFeelessDone
	PacsDeposit_ReportRegistered []EventPacsDepositReportRegistered
}

type EventPacsDepositReportRegistered struct {
	Phase  types.Phase
	Who    types.AccountID
	Com    types.AccountID
	Id     string
	Topics []types.Hash
}
