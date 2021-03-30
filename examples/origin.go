package main

import (
	"fmt"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v2"
	"github.com/centrifuge/go-substrate-rpc-client/v2/config"
	// "github.com/centrifuge/go-substrate-rpc-client/v2/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v2/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	subkey "github.com/vedhavyas/go-subkey"

	"github.com/skyvein-baas/client-skyvein-golang-api/handlers"
	"github.com/skyvein-baas/client-skyvein-golang-api/models"
)

// func init() {
// 	b64 := "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48"
// 	pubkey, _ := types.HexDecodeString(b64)
// 	addr, _ := subkey.SS58Address(pubkey, 42)
// 	fmt.Println(addr)
// 	decoded := models.DecodeSS58Address(addr)
// 	fmt.Println(pubkey)
// 	fmt.Println(decoded)
// }

const (
	unit = 1000000000000
)

func main() {
	cfg := config.Default()
	api, err := gsrpc.NewSubstrateAPI(cfg.RPCURL)
	if err != nil {
		panic(err)
	}
	eventApi, err := gsrpc.NewSubstrateAPI(cfg.RPCURL)
	if err != nil {
		panic(err)
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		panic(err)
	}
	types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})

	t1seed := "melt draft shy egg tomorrow there below flash patient code butter blind"
	t1key, _ := signature.KeyringPairFromSecret(t1seed, 42)
	// Create a call, transferring 12345 units to Bob
	// bob, err := models.NewMultiAddressFromHexAccountID("0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48")
	// if err != nil {
	// 	panic(err)
	// }

	// amount := types.NewUCompactFromUInt(1000 * unit)
	// t1 := models.NewMultiAddressFromAccountID(t1key.PublicKey)
	// c, err := types.NewCall(meta, "Balances.transfer", t1, amount)
	// if err != nil {
	// 	panic(err)
	// }
	c, err := types.NewCall(meta, "TraceSource.register_product",
		"33",
		models.OptionProps{
			HasValue: true,
			Value: []models.Prop{
				{Name: "11", Value: "22"},
			},
		})
	if err != nil {
		panic(err)
	}
	// sudo call
	c, err = types.NewCall(meta, "Feeless.feeless", c)
	if err != nil {
		panic(err)
	}

	// Create the extrinsic
	ext := models.NewMyExtrinsic(c)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		panic(err)
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		panic(err)
	}

	// Get the nonce for Alice
	key, err := types.CreateStorageKey(meta, "System", "Account", t1key.PublicKey, nil)
	if err != nil {
		panic(err)
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		panic(err)
	} else if !ok {
		panic("account not exist")
	}

	nonce := uint32(accountInfo.Nonce)
	o := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: rv.TransactionVersion,
	}

	// aliceAddr, _ := subkey.SS58Address(signature.TestKeyringPairAlice.PublicKey, 42)
	// toAddr, _ := subkey.SS58Address(t1key.PublicKey, 42)
	// fmt.Printf("Sending %v from %v to %v with nonce %v\n", amount, aliceAddr, toAddr, nonce)

	// Sign the transaction using Alice's default account
	err = ext.Sign(t1key, o)
	if err != nil {
		panic(err)
	}

	// Subscribe to system events via storage
	eventKey, err := types.CreateStorageKey(meta, "System", "Events", nil, nil)
	if err != nil {
		panic(err)
	}

	eventSub, err := eventApi.RPC.State.SubscribeStorageRaw([]types.StorageKey{eventKey})
	if err != nil {
		panic(err)
	}
	defer eventSub.Unsubscribe()

	// ext = myext.Extrinsic
	// Do the transfer and track the actual status
	sub, err := handlers.AuthorSubmitAndWatchExtrinsic(api.Client, ext)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	var exitCh chan bool = make(chan bool, 3)
	var isInBlock, isEventOver bool
	var errRst string
	for {
		select {
		case _ = <-exitCh:
			if isInBlock && isEventOver {
				close(exitCh)
				return
			}
		case status := <-sub.Chan():
			// fmt.Printf("Transaction status: %#v\n", status)
			if status.IsInBlock {
				isInBlock = true
				exitCh <- false
			}
		case set := <-eventSub.Chan():
			for _, chng := range set.Changes {
				if !types.Eq(chng.StorageKey, eventKey) || !chng.HasStorageData {
					// skip, we are only interested in events with content
					continue
				}
				events := models.TraceSourceEventRecords{}
				err = types.EventRecordsRaw(chng.StorageData).DecodeEventRecords(meta, &events)
				if err != nil {
					fmt.Println("DecodeEventRecords", err)
					continue
				}
				for _, e := range events.System_ExtrinsicSuccess {
					fmt.Printf("\tSystem:ExtrinsicSuccess:: (phase=%#v)\n", e.Phase)
				}
				for _, e := range events.System_ExtrinsicFailed {
					fmt.Printf("\tSystem:ExtrinsicFailed:: (phase=%#v)\n", e.Phase)
					fmt.Printf("\t\t%+v\n", e.DispatchError)
					dispatchError := e.DispatchError
					if dispatchError.HasModule {
						errStr := string(meta.AsMetadataV12.Modules[dispatchError.Module].Name) + "." + string(meta.AsMetadataV12.Modules[dispatchError.Module].Errors[dispatchError.Error].Name)
						errRst = errStr
						fmt.Println(errStr)
						isEventOver = true
						exitCh <- false
					}
				}
				for _, e := range events.TraceSource_ProductRegistered {
					fmt.Printf("\tTraceSource:ProductRegistered:: (phase=%#v)\n", e.Phase)
					who, _ := subkey.SS58Address(e.Who[:], 42)
					owner, _ := subkey.SS58Address(e.Owner[:], 42)
					fmt.Printf("\t\t%+v, %+v, %+v\n", who, e.Id, owner)
					isEventOver = true
					exitCh <- false
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
							fmt.Println(errStr)
						}
						isEventOver = true
					}
					exitCh <- false
				}
			}
		}

	}
	if errRst != "" {
		fmt.Println(errRst)
		return
	}
}
