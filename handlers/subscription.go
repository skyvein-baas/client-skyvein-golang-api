package handlers

import (
	"context"
	"sync"

	"github.com/centrifuge/go-substrate-rpc-client/v2/client"
	"github.com/centrifuge/go-substrate-rpc-client/v2/config"
	gethrpc "github.com/centrifuge/go-substrate-rpc-client/v2/gethrpc"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"

	"github.com/skyvein-baas/client-skyvein-golang-api/models"
)

// xt: types.Extrinsic
func AuthorSubmitAndWatchExtrinsic(cli client.Client, xt models.MyExtrinsic) (*MyExtrinsicStatusSubscription, error) { //nolint:lll
	ctx, cancel := context.WithTimeout(context.Background(), config.Default().SubscribeTimeout)
	defer cancel()

	c := make(chan types.ExtrinsicStatus)

	enc, err := types.EncodeToHexString(xt)
	if err != nil {
		return nil, err
	}

	sub, err := cli.Subscribe(ctx, "author", "submitAndWatchExtrinsic", "unwatchExtrinsic", "extrinsicUpdate",
		c, enc)
	if err != nil {
		return nil, err
	}

	return &MyExtrinsicStatusSubscription{Sub: sub, Channel: c}, nil
}

// ExtrinsicStatusSubscription is a subscription established through one of the Client's subscribe methods.
type MyExtrinsicStatusSubscription struct {
	Sub      *gethrpc.ClientSubscription
	Channel  chan types.ExtrinsicStatus
	QuitOnce sync.Once // ensures quit is closed once
}

// Chan returns the subscription channel.
//
// The channel is closed when Unsubscribe is called on the subscription.
func (s *MyExtrinsicStatusSubscription) Chan() <-chan types.ExtrinsicStatus {
	return s.Channel
}

// Err returns the subscription error channel. The intended use of Err is to schedule
// resubscription when the client connection is closed unexpectedly.
//
// The error channel receives a value when the subscription has ended due
// to an error. The received error is nil if Close has been called
// on the underlying client and no other error has occurred.
//
// The error channel is closed when Unsubscribe is called on the subscription.
func (s *MyExtrinsicStatusSubscription) Err() <-chan error {
	return s.Sub.Err()
}

// Unsubscribe unsubscribes the notification and closes the error channel.
// It can safely be called more than once.
func (s *MyExtrinsicStatusSubscription) Unsubscribe() {
	s.Sub.Unsubscribe()
	s.QuitOnce.Do(func() {
		close(s.Channel)
	})
}
