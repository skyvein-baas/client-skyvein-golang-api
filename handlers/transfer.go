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

type BlsTransfer struct {
	Client models.Client
}

func NewBlsTransfer(cli models.Client) *BlsTransfer {
	return &BlsTransfer{
		Client: cli,
	}
}

func (b *BlsTransfer) TransferTo(who string, amount uint64) (ok bool, err error) {
	// 创建连接
	api, err := gsrpc.NewSubstrateAPI(b.Client.Addr)
	if err != nil {
		return
	}
	eventApi, err := gsrpc.NewSubstrateAPI(b.Client.Addr)
	if err != nil {
		return
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return
	}
	types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})

	keypair, _ := signature.KeyringPairFromSecret(b.Client.Seed, 42)
	// 构造请求参数
	accountIdByt := models.DecodeSS58Address(who)
	t1 := models.NewMultiAddressFromAccountID(accountIdByt)
	amt := types.NewUCompactFromUInt(amount)
	c, err := types.NewCall(meta, "Balances.transfer", t1, amt)
	if err != nil {
		return
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
	okGet, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return
	} else if !okGet {
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
				events := types.EventRecords{}
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

				for _, e := range events.Balances_Transfer {
					fmt.Printf("\tBalances:Transfer:: (phase=%#v)\n", e.Phase)
					from, _ := subkey.SS58Address(e.From[:], 42)
					to, _ := subkey.SS58Address(e.To[:], 42)
					fmt.Printf("\t\t%+v, %+v, %+v\n", from, to, e.Value)
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
	ok = true
	return
}
