package handlers

import (
	"errors"
	"fmt"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v2"
	"github.com/centrifuge/go-substrate-rpc-client/v2/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	subkey "github.com/vedhavyas/go-subkey"

	"github.com/skyvein-baas/client-skyvein-golang-api/models"
)

type PacsDeposit struct {
	Client    models.Client
	IsFeeless bool
}

func (pd *PacsDeposit) Feeless() *PacsDeposit {
	pd.IsFeeless = true
	return pd
}

func NewPacsDeposit(cli models.Client) *PacsDeposit {
	return &PacsDeposit{
		Client: cli,
	}
}

// 注册产品
func (pd *PacsDeposit) RegisterReport(req models.RegisterReportReq) (
	out *models.RegisterReportRst, err error) {
	// 创建连接
	api, err := gsrpc.NewSubstrateAPI(pd.Client.Addr)
	if err != nil {
		return
	}
	eventApi, err := gsrpc.NewSubstrateAPI(pd.Client.Addr)
	if err != nil {
		return
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return
	}
	types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})

	keypair, _ := signature.KeyringPairFromSecret(pd.Client.Seed, 42)
	// 构造请求参数
	args := models.OptionReportProps{}
	if len(req.Props) > 0 {
		args.HasValue = true
		args.Value = req.Props
	}
	c, err := types.NewCall(meta, "PacsDeposit.register_report",
		req.ID, args)
	if err != nil {
		return
	}
	if pd.IsFeeless {
		c, err = types.NewCall(meta, "Feeless.feeless", c)
		if err != nil {
			return
		}
	}

	// Create the extrinsic
	ext := models.NewMyExtrinsic(c)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return
	}

	// Get the nonce
	key, err := types.CreateStorageKey(meta, "System", "Account", keypair.PublicKey, nil)
	if err != nil {
		return
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return
	} else if !ok {
		err = errors.New("account not exist")
		return
	}
	nonce := uint32(accountInfo.Nonce)

	// Sign the transaction
	o := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: rv.TransactionVersion,
	}
	err = ext.Sign(keypair, o)
	if err != nil {
		return
	}

	// Subscribe to system events via storage
	eventKey, err := types.CreateStorageKey(meta, "System", "Events", nil, nil)
	if err != nil {
		return
	}

	// track events
	eventSub, err := eventApi.RPC.State.SubscribeStorageRaw([]types.StorageKey{eventKey})
	if err != nil {
		return
	}
	defer eventSub.Unsubscribe()

	// ext = ext.Extrinsic
	// Do it and track the actual status
	sub, err := AuthorSubmitAndWatchExtrinsic(api.Client, ext)
	if err != nil {
		return
	}
	defer sub.Unsubscribe()

	var exitCh chan bool = make(chan bool, 3)
	var isInBlock, isEventOver bool
	var errRst string
	var rst models.RegisterReportRst
forEnd:
	for {
		select {
		case <-exitCh:
			if isInBlock && isEventOver {
				close(exitCh)
				break forEnd
			}
		case status := <-sub.Chan():
			// fmt.Printf("Transaction status: %#v\n", status)
			if status.IsInBlock {
				isInBlock = true
				exitCh <- false
			}
		// 获取事件通知
		case set := <-eventSub.Chan():
			for _, chng := range set.Changes {
				if !types.Eq(chng.StorageKey, eventKey) || !chng.HasStorageData {
					// skip, we are only interested in events with content
					continue
				}
				events := models.PacsDepositEventRecords{}
				err = types.EventRecordsRaw(chng.StorageData).DecodeEventRecords(meta, &events)
				if err != nil {
					fmt.Println("[DecodeEventRecords]", err)
					continue
				}
				// for _, e := range events.System_ExtrinsicSuccess {
				// 	fmt.Printf("\tSystem:ExtrinsicSuccess:: (phase=%#v)\n", e.Phase)
				// }
				for _, e := range events.System_ExtrinsicFailed {
					// fmt.Printf("\tSystem:ExtrinsicFailed:: (phase=%#v)\n", e.Phase)
					fmt.Printf("\t\t%+v\n", e.DispatchError)
					dispatchError := e.DispatchError
					if dispatchError.HasModule {
						errStr := string(meta.AsMetadataV12.Modules[dispatchError.Module].Name) + "." + string(meta.AsMetadataV12.Modules[dispatchError.Module].Errors[dispatchError.Error].Name)
						errRst = errStr
						isEventOver = true
						exitCh <- false
					}
				}
				for _, e := range events.Feeless_FeelessDone {
					fmt.Printf("\tFeeless:FeelessDone:: (phase=%#v)\n", e.Phase)
					who, _ := subkey.SS58Address(e.Who[:], 42)
					fmt.Printf("\t\t%+v, %+v\n", who, e.Result)
					if !e.Result.Ok {
						dispatchError := e.Result.Error
						if dispatchError.HasModule {
							errStr := string(meta.AsMetadataV12.Modules[dispatchError.Module].Name) + "." + string(meta.AsMetadataV12.Modules[dispatchError.Module].Errors[dispatchError.Error].Name)
							errRst = errStr
						}
						isEventOver = true
					}
					exitCh <- false
				}
				for _, e := range events.PacsDeposit_ReportRegistered {
					fmt.Printf("\tPacsDeposit:ReportRegistered:: (phase=%#v)\n", e.Phase)
					// get result
					rst.Who = e.Who[:]
					rst.Com = e.Com[:]
					rst.Id = e.Id
					who, _ := subkey.SS58Address(e.Who[:], 42)
					com, _ := subkey.SS58Address(e.Com[:], 42)
					fmt.Printf("\t\t%+v, %+v, %+v\n", who, e.Id, com)
					isEventOver = true
					exitCh <- false
				}
			}
		}
	}
	if errRst != "" {
		err = errors.New(errRst)
		return
	}
	out = &rst
	return
}
