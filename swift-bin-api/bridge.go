package swiftbinapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"

	"github.com/celestiaorg/celestia-node/libs/keystore"
	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/ipfs/go-datastore"
)

type BridgeMessaging struct {
	Cmd     string
	Payload json.RawMessage
}

var (
	CommunicationIn  chan string
	CommunicationOut chan BridgeMessaging
	bridgeLog        = slog.New(slog.NewTextHandler(os.Stderr, nil))
	loggingMsgIn     atomic.Bool
	loggingMsgOut    atomic.Bool
)

const (
	CMD_INIT_NODE      = "load_celestia_node"
	CMD_START_NODE     = "start_celestia_node"
	CMD_STOP_NODE      = "stop_celestia_node"
	RUN_RECEIVE_HEADER = "celestia_new_header"
)

type BridgeCmdLoadNode struct {
	StorePath string
}

type stubbedStore struct {
	stubbedKS *stubbedKeystore
}

type stubbedKeystore struct {
	simple map[string]any
}

func (ks *stubbedKeystore) Put(keystore.KeyName, keystore.PrivKey) error {
	return nil
}

// Get reads PrivKey using given KeyName.
func (ks *stubbedKeystore) Get(keystore.KeyName) (keystore.PrivKey, error) {
	return keystore.PrivKey{}, keystore.ErrNotFound
}

// Delete erases PrivKey using given KeyName.
func (ks *stubbedKeystore) Delete(name keystore.KeyName) error {
	return nil
}

// List lists all stored key names.
func (ks *stubbedKeystore) List() ([]keystore.KeyName, error) {
	return nil, nil
}

// Path reports the path of the Keystore.
func (ks *stubbedKeystore) Path() string {
	return "~/.celestia/ks"
}

// Keyring returns the keyring corresponding to the node's
// keystore.
func (ks *stubbedKeystore) Keyring() keyring.Keyring {
	return nil
}

func (s *stubbedStore) Path() string {
	return "~/.celestia"
}

// Keystore provides a Keystore to access keys.
func (s *stubbedStore) Keystore() (keystore.Keystore, error) {
	return s.stubbedKS, nil
}

// Datastore provides a Datastore - a KV store for arbitrary data to be stored on disk.
func (s *stubbedStore) Datastore() (datastore.Batching, error) {
	return nil, nil
}

// Config loads the stored Node config.
func (s *stubbedStore) Config() (*nodebuilder.Config, error) {
	return nil, nil
}

// PutConfig alters the stored Node config.
func (s *stubbedStore) PutConfig(*nodebuilder.Config) error {
	return nil
}

// Close closes the Store freeing up acquired resources and locks.
func (s *stubbedStore) Close() error {
	return nil
}

func SetupListen(enableLogging bool, errorCB func(string)) {
	CommunicationIn = make(chan string)
	loggingMsgIn.Store(enableLogging)
	stubStore := &stubbedStore{
		&stubbedKeystore{
			simple: map[string]any{},
		},
	}

	go func() {

		for cmd := range CommunicationIn {
			var c BridgeMessaging

			if loggingMsgIn.Load() {
				bridgeLog.Info("GOLANG received", "msg", cmd)
			}

			if err := json.Unmarshal([]byte(cmd), &c); err != nil {
				errorCB(err.Error())
				continue
			}

			switch c.Cmd {
			case CMD_INIT_NODE:
				var l BridgeCmdLoadNode
				if err := json.Unmarshal(c.Payload, &l); err != nil {
					errorCB(err.Error())
					continue
				}

				// cfg := nodebuilder.DefaultConfig(node.Light)
				// // get it from documents?
				// storePath := l.StorePath

				// if err := nodebuilder.Init(
				// 	*cfg, storePath, node.Light,
				// ); err != nil {
				// 	errorCB(err.Error())
				// 	continue
				// }

				CommunicationOut <- BridgeMessaging{Cmd: c.Cmd}
			case CMD_START_NODE:
				network := p2p.DefaultNetwork
				cfg := nodebuilder.DefaultConfig(node.Light)
				nd, err := nodebuilder.NewStripped(node.Light, network, cfg, stubStore)
				if err != nil {
					errorCB(err.Error())
					continue
				}

				go func() {
					if err := nd.Start(context.Background()); err != nil {
						if err != nil {
							errorCB(err.Error())
							return
						}
					}
				}()

				CommunicationOut <- BridgeMessaging{Cmd: c.Cmd}
			case CMD_STOP_NODE:
				CommunicationOut <- BridgeMessaging{Cmd: c.Cmd}
			default:
				errorCB(fmt.Sprintf("Unknown command %s", c.Cmd))
			}
		}
	}()
}

func SetupReply(enableLogging bool, cb func(string)) {
	CommunicationOut = make(chan BridgeMessaging)
	loggingMsgOut.Store(enableLogging)

	go func() {

		for doReply := range CommunicationOut {
			if loggingMsgOut.Load() {
				bridgeLog.Info(
					"GOLANG sending",
					"msg cmd", doReply.Cmd,
					"msg payload", string(doReply.Payload),
				)
			}

			encode, _ := json.Marshal(doReply)
			cb(string(encode))
		}
	}()
}
