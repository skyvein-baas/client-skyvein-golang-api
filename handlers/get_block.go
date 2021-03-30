package handlers

import (
	"github.com/centrifuge/go-substrate-rpc-client/v2/client"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"

	"github.com/skyvein-baas/client-skyvein-golang-api/models"
)

func ChainGetBlock(cli client.Client, blockHash *types.Hash) (*models.MySignedBlock, error) {
	var SignedBlock models.MySignedBlock
	err := client.CallWithBlockHash(cli, &SignedBlock, "chain_getBlock", blockHash)
	if err != nil {
		return nil, err
	}
	return &SignedBlock, err
}
