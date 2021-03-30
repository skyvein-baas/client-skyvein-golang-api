package models

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v2/scale"
)

const (
	NanosInSecond  = 1e9
	MillisInSecond = 1e3
)

type CompactMoment struct {
	time.Time
}

// NewMoment creates a new Moment type
func NewCompactMoment(t time.Time) CompactMoment {
	return CompactMoment{t}
}

func (m *CompactMoment) Decode(decoder scale.Decoder) error {
	compactT, err := decoder.DecodeUintCompact()
	if err != nil {
		return err
	}
	u := compactT.Uint64()

	// Error in case of overflow
	if u > math.MaxInt64 {
		return fmt.Errorf("cannot decode a uint64 into a CompactMoment if it overflows int64")
	}

	secs := u / MillisInSecond
	nanos := (u % uint64(MillisInSecond)) * uint64(NanosInSecond)

	*m = NewCompactMoment(time.Unix(int64(secs), int64(nanos)))

	return nil
}

func (m CompactMoment) Encode(encoder scale.Encoder) error {
	err := encoder.EncodeUintCompact(*big.NewInt(0).SetUint64(uint64(m.UnixNano() / (NanosInSecond / MillisInSecond))))
	if err != nil {
		return err
	}

	return nil
}
