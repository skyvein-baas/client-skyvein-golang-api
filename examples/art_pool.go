package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/skyvein-baas/client-skyvein-golang-api/handlers"
	"github.com/skyvein-baas/client-skyvein-golang-api/models"
)

func main() {
	client := models.Client{
		Addr: "ws://localhost:9944",
		Seed: "melt draft shy egg tomorrow there below flash patient code butter blind",
	}
	artPool := handlers.NewArtPool(client).Feeless()
	exitChan := make(chan string, 1)
	// 添加芯片
	go func() {
		for i := 0; i < 400; i++ {
			rstNfcAdd, err := artPool.NfcAdd(models.NfcAddReq{
				Arr: []models.NfcParm{
					{
						NfcId:  "x" + fmt.Sprint(i),
						NfcKey: "x" + fmt.Sprint(i),
					},
				},
			})
			if err != nil {
				if strings.Contains(err.Error(), "Priority is too low") {
					time.Sleep(time.Second)
					continue
				}
				exitChan <- err.Error() + ", x" + fmt.Sprint(i)
				// fmt.Println(err, i)
				break
			} else {
				fmt.Println("[NfcAdd]", models.SS58Addr(rstNfcAdd.Who))
			}
		}
	}()
	go func() {
		for i := 0; i < 400; i++ {
			rstNfcAdd, err := artPool.NfcAdd(models.NfcAddReq{
				Arr: []models.NfcParm{
					{
						NfcId:  "y" + fmt.Sprint(i),
						NfcKey: "y" + fmt.Sprint(i),
					},
				},
			})
			if err != nil {
				if strings.Contains(err.Error(), "Priority is too low") {
					time.Sleep(time.Second)
					continue
				}
				exitChan <- err.Error() + ", y" + fmt.Sprint(i)
				// fmt.Println(err, i)
				break
			} else {
				fmt.Println("[NfcAdd]", models.SS58Addr(rstNfcAdd.Who))
			}
		}
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			rstNfcAdd, err := artPool.NfcAdd(models.NfcAddReq{
				Arr: []models.NfcParm{
					{
						NfcId:  fmt.Sprint(i),
						NfcKey: fmt.Sprint(i),
					},
				},
			})
			if err != nil {
				if strings.Contains(err.Error(), "Priority is too low") {
					time.Sleep(time.Second)
					continue
				}
				exitChan <- err.Error() + ", main" + fmt.Sprint(i)
				// fmt.Println(err, i)
				break
			} else {
				fmt.Println("[NfcAdd]", models.SS58Addr(rstNfcAdd.Who))
			}
		}
	}()
	fmt.Println(<-exitChan)
}
