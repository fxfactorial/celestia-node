package swiftbinapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"

	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
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

func SetupListen(enableLogging bool, errorCB func(string)) {
	CommunicationIn = make(chan string)
	loggingMsgIn.Store(enableLogging)

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
				// cfg := nodebuilder.DefaultConfig(node.Light)
				nd, err := nodebuilder.NewStripped(node.Light, network, nil, nil)
				if err != nil {
					errorCB(err.Error())
					continue
				}

				// go func() {
				// 	if err := nd.Start(context.Background()); err != nil {
				// 		if err != nil {
				// 			errorCB(err.Error())
				// 			return
				// 		}
				// 	}
				// }()

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
