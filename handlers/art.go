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

type Art struct {
	Client    models.Client
	IsFeeless bool
}

func (a *Art) Feeless() *Art {
	a.IsFeeless = true
	return a
}

func NewArt(cli models.Client) *Art {
	return &Art{
		Client: cli,
	}
}

// 注册艺术品
func (a *Art) RegisterArt(req models.RegisterArtReq) (out *models.RegisterArtRst, err error) {
	// 创建连接
	api, err := gsrpc.NewSubstrateAPI(a.Client.Addr)
	if err != nil {
		return
	}
	eventApi, err := gsrpc.NewSubstrateAPI(a.Client.Addr)
	if err != nil {
		return
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return
	}
	types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})

	keypair, _ := signature.KeyringPairFromSecret(a.Client.Seed, 42)
	// 构造请求参数
	var accountIdByt []byte
	if req.Owner != "" {
		accountIdByt = models.DecodeSS58Address(req.Owner)
	}
	owner := types.NewAccountID(accountIdByt)
	c, err := types.NewCall(meta, "Art.register_art",
		req.ID, owner, req.EntityIds, req.Props, req.HashMethod, req.Hash, req.DnaMethod, req.Dna)
	if err != nil {
		return
	}
	if a.IsFeeless {
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
	var rst models.RegisterArtRst
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
				events := models.ArtEventRecords{}
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
				for _, e := range events.Art_ArtRegistered {
					fmt.Printf("\tArt:ArtRegistered:: (phase=%#v)\n", e.Phase)
					// get result
					rst.Who = e.Who[:]
					rst.Id = e.Id
					rst.Owner = e.Owner[:]
					who, _ := subkey.SS58Address(e.Who[:], 42)
					owner, _ := subkey.SS58Address(e.Owner[:], 42)
					fmt.Printf("\t\t%+v, %+v, %+v\n", who, e.Id, owner)
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

// 授权
func (a *Art) ApproveForAll(req models.ApproveForAllReq) (out *models.ApproveForAllRst, err error) {
	// 创建连接
	api, err := gsrpc.NewSubstrateAPI(a.Client.Addr)
	if err != nil {
		return
	}
	eventApi, err := gsrpc.NewSubstrateAPI(a.Client.Addr)
	if err != nil {
		return
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return
	}
	types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})

	keypair, _ := signature.KeyringPairFromSecret(a.Client.Seed, 42)
	// 构造请求参数
	var accountIdByt []byte
	if req.To != "" {
		accountIdByt = models.DecodeSS58Address(req.To)
	}
	to := types.NewAccountID(accountIdByt)
	var isClear types.OptionBool
	if req.IsClear {
		isClear = types.NewOptionBool(types.Bool(req.IsClear))
	}
	c, err := types.NewCall(meta, "Art.approve_for_all",
		to, req.Approved, isClear)
	if err != nil {
		return
	}
	if a.IsFeeless {
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
	var rst models.ApproveForAllRst
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
				events := models.ArtEventRecords{}
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
				for _, e := range events.Art_ApproveForAll {
					fmt.Printf("\tArt:ApproveForAll:: (phase=%#v)\n", e.Phase)
					// get result
					rst.Caller = e.Caller[:]
					rst.To = e.To[:]
					rst.Approved = e.Approved
					rst.IsClear = e.IsClear
					caller, _ := subkey.SS58Address(e.Caller[:], 42)
					to, _ := subkey.SS58Address(e.To[:], 42)
					fmt.Printf("\t\t%+v, %+v, %+v, %+v\n", caller, to, e.Approved, e.IsClear)
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

// 实体授权
func (a *Art) ApproveFor(req models.ApproveForReq) (out *models.ApproveForRst, err error) {
	// 创建连接
	api, err := gsrpc.NewSubstrateAPI(a.Client.Addr)
	if err != nil {
		return
	}
	eventApi, err := gsrpc.NewSubstrateAPI(a.Client.Addr)
	if err != nil {
		return
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return
	}
	types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})

	keypair, _ := signature.KeyringPairFromSecret(a.Client.Seed, 42)
	fmt.Println("keypair.URI", keypair.URI)
	// 构造请求参数
	var accountIdByt []byte
	if req.To != "" {
		accountIdByt = models.DecodeSS58Address(req.To)
	}
	to := types.NewAccountID(accountIdByt)
	c, err := types.NewCall(meta, "Art.approve_for",
		to, req.EntityIds)
	if err != nil {
		return
	}
	if a.IsFeeless {
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
	var rst models.ApproveForRst
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
				events := models.ArtEventRecords{}
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
				for _, e := range events.Art_ApproveFor {
					fmt.Printf("\tArt:ApproveFor:: (phase=%#v)\n", e.Phase)
					// get result
					rst.Caller = e.Caller[:]
					rst.To = e.To[:]
					caller, _ := subkey.SS58Address(e.Caller[:], 42)
					to, _ := subkey.SS58Address(e.To[:], 42)
					fmt.Printf("\t\t%+v, %+v\n", caller, to)
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

// 艺术品添加实物
func (a *Art) ArtRaise(req models.ArtRaiseReq) (out *models.ArtRaiseRst, err error) {
	// 创建连接
	api, err := gsrpc.NewSubstrateAPI(a.Client.Addr)
	if err != nil {
		return
	}
	eventApi, err := gsrpc.NewSubstrateAPI(a.Client.Addr)
	if err != nil {
		return
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return
	}
	types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})

	keypair, _ := signature.KeyringPairFromSecret(a.Client.Seed, 42)
	// 构造请求参数
	c, err := types.NewCall(meta, "Art.art_raise",
		req.ArtId, req.EntityIds)
	if err != nil {
		return
	}
	if a.IsFeeless {
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
	var rst models.ArtRaiseRst
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
				events := models.ArtEventRecords{}
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
				for _, e := range events.Art_ArtRaise {
					fmt.Printf("\tArt:ArtRaise:: (phase=%#v)\n", e.Phase)
					// get result
					rst.Who = e.Who[:]
					rst.ArtId = e.ArtId
					who, _ := subkey.SS58Address(e.Who[:], 42)
					fmt.Printf("\t\t%+v, %+v\n", who, e.ArtId)
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

// 添加芯片
func (a *Art) NfcAdd(req models.NfcAddReq) (out *models.NfcAddRst, err error) {
	// 创建连接
	api, err := gsrpc.NewSubstrateAPI(a.Client.Addr)
	if err != nil {
		return
	}
	eventApi, err := gsrpc.NewSubstrateAPI(a.Client.Addr)
	if err != nil {
		return
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return
	}
	types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})

	keypair, _ := signature.KeyringPairFromSecret(a.Client.Seed, 42)
	// 构造请求参数
	c, err := types.NewCall(meta, "Art.nfc_add",
		req.Arr)
	if err != nil {
		return
	}
	if a.IsFeeless {
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
	var rst models.NfcAddRst
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
				events := models.ArtEventRecords{}
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
				for _, e := range events.Art_NfcAdd {
					fmt.Printf("\tArt:NfcAdd:: (phase=%#v)\n", e.Phase)
					// get result
					rst.Who = e.Who[:]
					who, _ := subkey.SS58Address(e.Who[:], 42)
					fmt.Printf("\t\t%+v\n", who)
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

// 芯片绑定
func (a *Art) NfcBind(req models.NfcBindReq) (out *models.NfcBindRst, err error) {
	// 创建连接
	api, err := gsrpc.NewSubstrateAPI(a.Client.Addr)
	if err != nil {
		return
	}
	eventApi, err := gsrpc.NewSubstrateAPI(a.Client.Addr)
	if err != nil {
		return
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return
	}
	types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})

	keypair, _ := signature.KeyringPairFromSecret(a.Client.Seed, 42)
	// 构造请求参数
	c, err := types.NewCall(meta, "Art.nfc_bind",
		req.EntityId, req.NfcId)
	if err != nil {
		return
	}
	if a.IsFeeless {
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
	var rst models.NfcBindRst
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
				events := models.ArtEventRecords{}
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
				for _, e := range events.Art_NfcBind {
					fmt.Printf("\tArt:Art_NfcBind:: (phase=%#v)\n", e.Phase)
					// get result
					rst.Who = e.Who[:]
					who, _ := subkey.SS58Address(e.Who[:], 42)
					rst.EntityId = e.EntityId
					rst.NfcId = e.NfcId
					fmt.Printf("\t\t%+v, %+v, %+v\n", who, e.EntityId, e.NfcId)
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