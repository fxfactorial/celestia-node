package main

/*
extern void send_cmd_back(char*);
extern void send_error_back(char*);
*/
import "C"
import (
	"strings"

	swiftbinapi "github.com/celestiaorg/celestia-node/swift-bin-api"
)

//export MakeChannelAndListenThread
func MakeChannelAndListenThread(enableLogging bool) {
	swiftbinapi.SetupListen(enableLogging, func(error string) {
		C.send_error_back(C.CString(error))
	})
}

//export MakeChannelAndReplyThread
func MakeChannelAndReplyThread(enableLogging bool) {
	swiftbinapi.SetupReply(enableLogging, func(reply string) {
		C.send_cmd_back(C.CString(reply))
	})
}

//export UISendCmd
func UISendCmd(msg string) {
	cl := strings.Clone(msg)
	swiftbinapi.CommunicationIn <- cl
}

func main() {
	//
}
