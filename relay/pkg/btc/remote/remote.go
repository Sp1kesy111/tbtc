package remote

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/ipfs/go-log"
	"github.com/keep-network/tbtc/relay/pkg/btc"
)

var logger = log.Logger("relay-btc-remote")

const connectionTimeout = 3 * time.Second

// remoteChain represents a remote Bitcoin chain.
type remoteChain struct {
	client *rpcclient.Client
}

// Connect connects to the Bitcoin chain and returns a chain handle.
func Connect(
	ctx context.Context,
	config *btc.Config,
) (btc.Handle, error) {
	logger.Infof("connecting remote Bitcoin chain")

	connCfg := &rpcclient.ConnConfig{
		User:         config.Username,
		Pass:         config.Password,
		Host:         config.URL,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create rpc client at [%s]: [%v]",
			config.URL,
			err,
		)
	}

	err = testConnection(client, connectionTimeout)
	if err != nil {
		return nil, fmt.Errorf(
			"error while connecting to [%s]: [%v]; check if the Bitcoin node "+
				"is running and you provided correct credentials and url",
			config.URL,
			err,
		)
	}

	// When the context is done, cancel all requests from the RPC client
	// and disconnect it.
	go func() {
		<-ctx.Done()
		logger.Info("disconnecting from remote Bitcoin chain")
		client.Shutdown()
	}()

	return &remoteChain{client: client}, nil
}

// GetHeaderByHeight returns the block header for the given block height.
func (rc *remoteChain) GetHeaderByHeight(height *big.Int) (*btc.Header, error) {
	blockHash, err := rc.client.GetBlockHash(height.Int64())
	if err != nil {
		return nil, fmt.Errorf(
			"could not get block hash for height [%d]: [%v]",
			height.Int64(),
			err,
		)
	}

	blockHeader, err := rc.client.GetBlockHeader(blockHash)
	if err != nil {
		return nil, fmt.Errorf(
			"could not get block header for hash [%s]: [%v]",
			blockHash.String(),
			err,
		)
	}

	rawHeader, err := serializeHeader(blockHeader)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to serialize header for block with hash [%s]: [%v]",
			blockHash.String(),
			err,
		)
	}

	// treat all digests as little-endian representation
	relayHeader := &btc.Header{
		Hash:       btc.NewLittleEndianDigest(blockHeader.BlockHash()),
		PrevHash:   btc.NewLittleEndianDigest(blockHeader.PrevBlock),
		MerkleRoot: btc.NewLittleEndianDigest(blockHeader.MerkleRoot),
		Raw:        rawHeader,
		Height:     height.Int64(),
	}

	return relayHeader, nil
}

// GetHeaderByDigest returns the block header for given digest (hash).
func (rc *remoteChain) GetHeaderByDigest(
	digest *btc.Digest,
) (*btc.Header, error) {
	// TODO: implementation
	return nil, nil
}

func testConnection(client *rpcclient.Client, timeout time.Duration) error {
	errChan := make(chan error, 1)

	go func() {
		_, err := client.GetBlockCount()
		errChan <- err
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(timeout):
		return fmt.Errorf(
			"connection timed out after [%f] seconds",
			timeout.Seconds(),
		)
	}
}

// serializeHeader serializes header fields using little-endian representation
// and returns concatenation of them.
func serializeHeader(header *wire.BlockHeader) ([]byte, error) {
	var buffer bytes.Buffer

	err := header.Serialize(&buffer)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
