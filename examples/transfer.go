package main

import (
	"fmt"

	"github.com/skyvein-baas/client-skyvein-golang-api/handlers"
	"github.com/skyvein-baas/client-skyvein-golang-api/models"
)

func main() {
	const unit = 1000000000000
	client := models.Client{
		Addr: "ws://localhost:9944",
		Seed: "melt draft shy egg tomorrow there below flash patient code butter blind",
	}
	bls := handlers.NewBlsTransfer(client)
	ok, err := bls.TransferTo("5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY", 100*unit)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ok)
}
