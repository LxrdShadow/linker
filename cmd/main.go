package main

import (
	"fmt"
	"os"

	// "github.com/LxrdShadow/linker/internal/protocol"
	"github.com/LxrdShadow/linker/pkg/transfer"
	"github.com/LxrdShadow/linker/pkg/util"
)

func main() {
	flagConfig, err := util.ParseFlags(os.Args)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	switch flagConfig.Mode {
	case "send":
		sender := transfer.NewSender(flagConfig.Host, flagConfig.Port, "tcp", flagConfig.FilePath)
		err := sender.Listen()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		}

	case "receive":
		receiver := transfer.NewReceiver()
		err := receiver.Connect(flagConfig.Host, flagConfig.Port, "tcp")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		}
	}
}
