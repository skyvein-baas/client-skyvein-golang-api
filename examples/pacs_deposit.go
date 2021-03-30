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
	psrc := handlers.NewPacsDeposit(client).Feeless()
	// 注册产品
	rstRegisterProduct, err := psrc.RegisterReport(models.RegisterReportReq{
		ID: "11114",
		Props: []models.ReportProp{
			{Name: "xxx33333333333333333", Value: "222"},
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("[RegisterReport]", models.SS58Addr(rstRegisterProduct.Who), models.SS58Addr(rstRegisterProduct.Com), rstRegisterProduct.Id)
}
