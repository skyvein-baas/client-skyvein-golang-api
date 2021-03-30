package handlers

import (
	"fmt"
	"time"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v2"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"

	"github.com/skyvein-baas/client-skyvein-golang-api/models"
)

type FuncAndEventOfBlock interface {
	HandleFunc(moduleName string, funcName string, argsByt []byte) error
	HandleEvent(meta *types.Metadata, rawEvents types.EventRecordsRaw) error
}

func WatchHistoricalBlocks(cfg models.Client, handler FuncAndEventOfBlock) (err error) {
	// 创建连接
	api, err := gsrpc.NewSubstrateAPI(cfg.Addr)
	if err != nil {
		return
	}
	curr := 0
	for {
		meta, err := api.RPC.State.GetMetadataLatest()
		if err != nil {
			return err
		}
		latestBlockHash, err := api.RPC.Chain.GetBlockHashLatest()
		if err != nil {
			return err
		}
		lastOriBlock, err := ChainGetBlock(api.Client, &latestBlockHash)
		if err != nil {
			return err
		}
		orilast := lastOriBlock.Block.Header.Number
		last := int(orilast)
		if curr >= last {
			// // close
			// (*MyClient)(unsafe.Pointer(&api.Client)).Close()
			// api, meta = nil, nil
			time.Sleep(time.Second * 10)
			continue
		}
		for i := curr; i <= last; i++ {
			blockHash, err := api.RPC.Chain.GetBlockHash(uint64(i))
			if err != nil {
				return err
			}

			// oriBlock, err := api.RPC.Chain.GetBlock(blockHash)
			oriBlock, err := ChainGetBlock(api.Client, &blockHash)
			if err != nil {
				return err
			}

			if len(oriBlock.Block.Extrinsics) > 0 {
				// ignore only timestamp.set
				if len(oriBlock.Block.Extrinsics) == 1 {
					moduleName := meta.AsMetadataV12.Modules[oriBlock.Block.Extrinsics[0].Method.CallIndex.SectionIndex].Name
					funcName := meta.AsMetadataV12.Modules[oriBlock.Block.Extrinsics[0].Method.CallIndex.SectionIndex].Calls[oriBlock.Block.Extrinsics[0].Method.CallIndex.MethodIndex].Name
					if moduleName == "Timestamp" && funcName == "set" {
						continue
					}
				}
				eventKey, err := types.CreateStorageKey(meta, "System", "Events", nil, nil)
				if err != nil {
					return err
				}
				oriEvents := types.EventRecordsRaw{}
				_, err = api.RPC.State.GetStorage(eventKey, &oriEvents, blockHash)
				if err != nil {
					return err
				}
				// fmt.Printf("oriBlock %+v\n", oriBlock)
				fmt.Println("[height]", oriBlock.Block.Header.Number)
				fmt.Println("========")
				for j := 0; j < len(oriBlock.Block.Extrinsics); j++ {
					var moduleName, funcName string
					var argsByt []byte
					extrinsic := oriBlock.Block.Extrinsics[j]
					moduleName = string(meta.AsMetadataV12.Modules[extrinsic.Method.CallIndex.SectionIndex].Name)
					funcName = string(meta.AsMetadataV12.Modules[extrinsic.Method.CallIndex.SectionIndex].Calls[extrinsic.Method.CallIndex.MethodIndex].Name)
					argsByt = extrinsic.Method.Args

					err = handler.HandleFunc(moduleName, funcName, argsByt)
					if err != nil {
						return err
					}
				}
				fmt.Println("[event]")
				err = handler.HandleEvent(meta, oriEvents)
				if err != nil {
					return err
				}
				fmt.Println("========")

			}
		}
		curr = last + 1
	}
}
