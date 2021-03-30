package models

import (
	"fmt"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v2/scale"
)

type RegisterProductReq struct {
	ID    string // 产品id
	Props []Prop // 产品属性
}

type RegisterProductRst struct {
	Who   []byte // 操作人
	Owner []byte // 所属组织
	Id    string // 产品id
}

type RegisterShipmentReq struct {
	ID         string   // 批次id
	ProductIds []string // 产品id
}

type RegisterShipmentRst struct {
	Who   []byte // 操作人
	Owner []byte // 所属组织
	Id    string // 批次id
}

type TrackShipmentReq struct {
	ID        string                // 批次id
	Operation ShippingOperationEnum // 类型
	Timestamp time.Time             //
	Location  *ReadPoint            // 位置
	Readings  []ReadingParm         // 参数
}

type TrackShipmentRst struct {
	Who            []byte // 操作人
	Id             string // 批次id
	EventIdx       uint64 // event.index
	ShipmentStatus string // 批次状态
}

type ReadingParm struct {
	DeviceId    string
	ReadingType EnumReadingType
	Timestamp   time.Time
	SensorValue string
}

type Prop struct {
	Name  string
	Value string
}

type OptionProps struct {
	HasValue bool
	Value    []Prop
}

func (m OptionProps) Encode(encoder scale.Encoder) (err error) {
	if !m.HasValue {
		err = encoder.PushByte(0)
		if err != nil {
			return
		}
	} else {
		err = encoder.PushByte(1)
		if err != nil {
			return
		}
		err = encoder.Encode(m.Value)
		if err != nil {
			return
		}
	}
	return
}

func (m OptionProps) Decode(decoder scale.Decoder) (err error) {
	b, _ := decoder.ReadOneByte()
	switch b {
	case 0:
		m.HasValue = false
	case 1:
		m.HasValue = true
		err = decoder.Decode(&m.Value)
		if err != nil {
			return
		}
	default:
		return fmt.Errorf("Unknown byte prefix for encoded OptionProps: %d", b)
	}
	return
}

func (m Prop) Encode(encoder scale.Encoder) (err error) {
	err = encoder.Encode(m.Name)
	if err != nil {
		return
	}

	err = encoder.Encode(m.Value)
	if err != nil {
		return
	}
	return
}

func (m Prop) Decode(decoder scale.Decoder) (err error) {
	err = decoder.Decode(&m.Name)
	if err != nil {
		return
	}

	err = decoder.Decode(&m.Value)
	if err != nil {
		return
	}
	return
}

type ShippingOperationEnum struct {
	IsPickup  bool
	IsScan    bool
	IsDeliver bool
}

func (e ShippingOperationEnum) Encode(encoder scale.Encoder) (err error) {
	if e.IsPickup {
		err = encoder.PushByte(0)
	} else if e.IsScan {
		err = encoder.PushByte(1)
	} else if e.IsDeliver {
		err = encoder.PushByte(2)
	}
	return
}

func (e *ShippingOperationEnum) Decode(decoder scale.Decoder) (err error) {
	b, err := decoder.ReadOneByte()
	switch b {
	case 0:
		e.IsPickup = true
	case 1:
		e.IsScan = true
	case 2:
		e.IsDeliver = true
	default:
		return fmt.Errorf("unknown ShippingOperationEnum enum: %v", b)
	}
	return
}

type OptionReadPoint struct {
	HasValue bool
	Value    ReadPoint
}

func (m OptionReadPoint) Encode(encoder scale.Encoder) (err error) {
	if !m.HasValue {
		err = encoder.PushByte(0)
		if err != nil {
			return
		}
	} else {
		err = encoder.PushByte(1)
		if err != nil {
			return
		}
		err = encoder.Encode(m.Value)
		if err != nil {
			return
		}
	}
	return
}

func (m OptionReadPoint) Decode(decoder scale.Decoder) (err error) {
	b, _ := decoder.ReadOneByte()
	switch b {
	case 0:
		m.HasValue = false
	case 1:
		m.HasValue = true
		err = decoder.Decode(&m.Value)
		if err != nil {
			return
		}
	default:
		return fmt.Errorf("Unknown byte prefix for encoded OptionReadPoint: %d", b)
	}
	return
}

type ReadPoint struct {
	Latitude  string
	Longitude string
}

func (m ReadPoint) Encode(encoder scale.Encoder) (err error) {
	err = encoder.Encode(m.Latitude)
	if err != nil {
		return
	}

	err = encoder.Encode(m.Longitude)
	if err != nil {
		return
	}
	return
}

func (m ReadPoint) Decode(decoder scale.Decoder) (err error) {
	err = decoder.Decode(&m.Latitude)
	if err != nil {
		return
	}

	err = decoder.Decode(&m.Longitude)
	if err != nil {
		return
	}
	return
}

type OptionReadings struct {
	HasValue bool
	Value    []Reading
}

func (m OptionReadings) Encode(encoder scale.Encoder) (err error) {
	if !m.HasValue {
		err = encoder.PushByte(0)
		if err != nil {
			return
		}
	} else {
		err = encoder.PushByte(1)
		if err != nil {
			return
		}
		err = encoder.Encode(m.Value)
		if err != nil {
			return
		}
	}
	return
}

func (m OptionReadings) Decode(decoder scale.Decoder) (err error) {
	b, _ := decoder.ReadOneByte()
	switch b {
	case 0:
		m.HasValue = false
	case 1:
		m.HasValue = true
		err = decoder.Decode(&m.Value)
		if err != nil {
			return
		}
	default:
		return fmt.Errorf("Unknown byte prefix for encoded OptionReadings: %d", b)
	}
	return
}

type Reading struct {
	DeviceId    string
	ReadingType EnumReadingType
	Timestamp   CompactMoment
	SensorValue string
}

func (m Reading) Encode(encoder scale.Encoder) (err error) {
	err = encoder.Encode(m.DeviceId)
	if err != nil {
		return
	}
	err = encoder.Encode(m.ReadingType)
	if err != nil {
		return
	}
	err = encoder.Encode(m.Timestamp)
	if err != nil {
		return
	}
	err = encoder.Encode(m.SensorValue)
	if err != nil {
		return
	}
	return
}

func (m Reading) Decode(decoder scale.Decoder) (err error) {
	err = decoder.Decode(&m.DeviceId)
	if err != nil {
		return
	}
	err = decoder.Decode(&m.ReadingType)
	if err != nil {
		return
	}
	err = decoder.Decode(&m.Timestamp)
	if err != nil {
		return
	}
	err = decoder.Decode(&m.SensorValue)
	if err != nil {
		return
	}
	return
}

type EnumReadingType struct {
	IsHumidity    bool
	IsPressure    bool
	IsShock       bool
	IsTilt        bool
	IsTemperature bool
	IsVibration   bool
}

func (e EnumReadingType) Encode(encoder scale.Encoder) (err error) {
	if e.IsHumidity {
		err = encoder.PushByte(0)
	} else if e.IsPressure {
		err = encoder.PushByte(1)
	} else if e.IsShock {
		err = encoder.PushByte(2)
	} else if e.IsTilt {
		err = encoder.PushByte(3)
	} else if e.IsTemperature {
		err = encoder.PushByte(4)
	} else if e.IsVibration {
		err = encoder.PushByte(5)
	}
	return
}

func (e *EnumReadingType) Decode(decoder scale.Decoder) (err error) {
	b, err := decoder.ReadOneByte()
	switch b {
	case 0:
		e.IsHumidity = true
	case 1:
		e.IsPressure = true
	case 3:
		e.IsShock = true
	case 4:
		e.IsTilt = true
	case 5:
		e.IsTemperature = true
	case 6:
		e.IsVibration = true
	default:
		return fmt.Errorf("unknown EnumReadingType enum: %v", b)
	}
	return
}
