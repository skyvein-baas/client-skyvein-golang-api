package main

import (
	"fmt"
	"time"

	"github.com/skyvein-baas/client-skyvein-golang-api/handlers"
	"github.com/skyvein-baas/client-skyvein-golang-api/models"
)

func main() {
	client := models.Client{
		Addr: "ws://localhost:9944",
		Seed: "melt draft shy egg tomorrow there below flash patient code butter blind",
	}
	tsrc := handlers.NewTraceSource(client).Feeless()
	// 注册产品
	rstRegisterProduct, err := tsrc.RegisterProduct(models.RegisterProductReq{
		ID: "31",
		Props: []models.Prop{
			{Name: "11", Value: "22"},
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("[RegisterProduct]", models.SS58Addr(rstRegisterProduct.Who), rstRegisterProduct.Id, models.SS58Addr(rstRegisterProduct.Owner))
	// 注册批次
	rstRegisterShipment, err := tsrc.RegisterShipment(models.RegisterShipmentReq{
		ID: "32",
		ProductIds: []string{
			"31",
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("[RegisterShipment]", models.SS58Addr(rstRegisterShipment.Who), rstRegisterShipment.Id, models.SS58Addr(rstRegisterShipment.Owner))
	// 提交批次
	rstTrackShipment1, err := tsrc.TrackShipment(models.TrackShipmentReq{
		ID: "32",
		Operation: models.ShippingOperationEnum{
			IsPickup: true,
		},
		Timestamp: time.Now(),
		Location: &models.ReadPoint{
			Latitude:  "1",
			Longitude: "2",
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("[TrackShipment]", models.SS58Addr(rstTrackShipment1.Who), rstTrackShipment1.Id, rstTrackShipment1.EventIdx, rstTrackShipment1.ShipmentStatus)
	// 提交批次
	_, err = tsrc.TrackShipment(models.TrackShipmentReq{
		ID: "32",
		Operation: models.ShippingOperationEnum{
			IsScan: true,
		},
		Timestamp: time.Now(),
		Location: &models.ReadPoint{
			Latitude:  "1",
			Longitude: "2",
		},
		Readings: []models.ReadingParm{
			{
				DeviceId:    "11",
				ReadingType: models.EnumReadingType{IsHumidity: true},
				Timestamp:   time.Now(),
				SensorValue: "213",
			},
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("[TrackShipment]")
	// 提交批次
	rstTrackShipment2, err := tsrc.TrackShipment(models.TrackShipmentReq{
		ID: "32",
		Operation: models.ShippingOperationEnum{
			IsDeliver: true,
		},
		Timestamp: time.Now(),
		Location: &models.ReadPoint{
			Latitude:  "6",
			Longitude: "7",
		},
		Readings: []models.ReadingParm{
			{
				DeviceId:    "11",
				ReadingType: models.EnumReadingType{IsShock: true},
				Timestamp:   time.Now(),
				SensorValue: "213",
			},
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("[TrackShipment]", models.SS58Addr(rstTrackShipment2.Who), rstTrackShipment2.Id, rstTrackShipment2.EventIdx, rstTrackShipment2.ShipmentStatus)
}
