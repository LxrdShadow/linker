package main

import (
	"fmt"
	"os"
	// "os"

	// "github.com/LxrdShadow/linker/internal/protocol"
	"github.com/LxrdShadow/linker/pkg/transfer"
	"github.com/LxrdShadow/linker/pkg/util"
)

func main() {
	flagConfig, err := util.ParseFlags()
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	switch flagConfig.Mode {
	case "send":
		sender := transfer.NewSender("localhost", "3000", "tcp", flagConfig.FilePath)
		err := sender.Listen()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		}

	case "receive":
		receiver := transfer.NewReceiver()
		receiver.Connect("localhost", "3000", "tcp")
	}
}
