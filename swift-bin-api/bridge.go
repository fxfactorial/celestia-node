package swiftbinapi

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync/atomic"
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
