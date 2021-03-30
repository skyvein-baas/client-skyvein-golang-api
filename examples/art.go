package main

import (
	"fmt"

	"github.com/skyvein-baas/client-skyvein-golang-api/handlers"
	"github.com/skyvein-baas/client-skyvein-golang-api/models"
)

func main() {
	client := models.Client{
		Addr: "ws://localhost:9944",
		Seed: "melt draft shy egg tomorrow there below flash patient code butter blind",
	}
	art := handlers.NewArt(client).Feeless()
	// 授权
	// rstApproveForAll, err := art.ApproveForAll(models.ApproveForAllReq{
	// 	To:       "5FHneW46xGXgs5mUiveU4sbTyGBzmstUspZC92UhjJM694ty",
	// 	Approved: true,
	// 	IsClear:  false,
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("[ApproveForAll]", models.SS58Addr(rstApproveForAll.Caller), models.SS58Addr(rstApproveForAll.To), rstApproveForAll.Approved, rstApproveForAll.IsClear)
	// 注册艺术品
	rstRegisterArt, err := art.RegisterArt(models.RegisterArtReq{
		ID:        "11",
		Owner:     "5FHneW46xGXgs5mUiveU4sbTyGBzmstUspZC92UhjJM694ty",
		EntityIds: []string{"11", "22"},
		Props: []models.ArtProperty{
			{
				Name:  "11",
				Value: "11",
			},
		},
		HashMethod: "11",
		Hash:       "11",
		DnaMethod:  "11",
		Dna:        "11",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("[RegisterArt]", models.SS58Addr(rstRegisterArt.Who), rstRegisterArt.Id, models.SS58Addr(rstRegisterArt.Owner))
	// // 艺术品添加实体
	// rstArtRaise, err := art.ArtRaise(models.ArtRaiseReq{
	// 	ArtId:     "11",
	// 	EntityIds: []string{"13", "14"},
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("[ArtRaise]", models.SS58Addr(rstArtRaise.Who), rstArtRaise.ArtId)
	// 添加芯片
	// rstNfcAdd, err := art.NfcAdd(models.NfcAddReq{
	// 	Arr: []models.NfcParm{
	// 		{
	// 			NfcId:  "11",
	// 			NfcKey: "111",
	// 		},
	// 		{
	// 			NfcId:  "22",
	// 			NfcKey: "222",
	// 		},
	// 	},
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("[NfcAdd]", models.SS58Addr(rstNfcAdd.Who))
	// 芯片绑定
	// rstNfcBind, err := art.NfcBind(models.NfcBindReq{
	// 	EntityId: "11///13",
	// 	NfcId:    "11",
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("[NfcBind]", rstNfcBind.EntityId, rstNfcBind.NfcId)
	// 实体授权
	// rstApproveFor, err := art.ApproveFor(models.ApproveForReq{
	// 	To:        "5FHneW46xGXgs5mUiveU4sbTyGBzmstUspZC92UhjJM694ty",
	// 	EntityIds: []string{"13", "14"},
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("[ApproveFor]", models.SS58Addr(rstApproveFor.Caller), models.SS58Addr(rstApproveFor.To))
}
