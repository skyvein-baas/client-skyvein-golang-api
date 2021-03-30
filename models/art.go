package models

import (
	"github.com/centrifuge/go-substrate-rpc-client/v2/scale"
)

type RegisterArtReq struct {
	ID         string
	Owner      string
	EntityIds  []string
	Props      []ArtProperty
	HashMethod string
	Hash       string
	DnaMethod  string
	Dna        string
}

type RegisterArtRst struct {
	Who   []byte // 操作人
	Owner []byte // 所属组织
	Id    string // art_id
}

type ArtProperty struct {
	Name  string
	Value string
}

func (m ArtProperty) Encode(encoder scale.Encoder) (err error) {
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

func (m ArtProperty) Decode(decoder scale.Decoder) (err error) {
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

type ApproveForAllReq struct {
	To       string // 给谁
	Approved bool   // true全权同意 false全权拒绝
	IsClear  bool   // 是否清除授权
}

type ApproveForAllRst struct {
	Caller   []byte // 操作人
	To       []byte // 给谁
	Approved bool   // 授权状态
	IsClear  bool   // 是否清除
}

type ApproveForReq struct {
	To        string // 给谁
	EntityIds []string
}

type ApproveForRst struct {
	Caller []byte // 操作人
	To     []byte // 给谁
}

type ArtRaiseReq struct {
	ArtId     string
	EntityIds []string
}

type ArtRaiseRst struct {
	Who   []byte // 操作人
	ArtId string // 艺术品id
}

type NfcAddReq struct {
	Arr []NfcParm
}

type NfcParm struct {
	NfcId  string
	NfcKey string
}

func (m NfcParm) Encode(encoder scale.Encoder) (err error) {
	err = encoder.Encode(m.NfcId)
	if err != nil {
		return
	}

	err = encoder.Encode(m.NfcKey)
	if err != nil {
		return
	}
	return
}

func (m NfcParm) Decode(decoder scale.Decoder) (err error) {
	err = decoder.Decode(&m.NfcId)
	if err != nil {
		return
	}

	err = decoder.Decode(&m.NfcKey)
	if err != nil {
		return
	}
	return
}

type NfcAddRst struct {
	Who []byte // 操作人
}

type NfcBindReq struct {
	EntityId string
	NfcId    string
}

type NfcBindRst struct {
	Who      []byte // 操作人
	EntityId string
	NfcId    string
}
