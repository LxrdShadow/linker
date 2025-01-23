package main

import (
	"os"

	"github.com/LxrdShadow/linker/pkg/log"
	"github.com/LxrdShadow/linker/pkg/transfer"
	"github.com/LxrdShadow/linker/pkg/util"
)

func main() {
	flagConfig, err := util.ParseFlags(os.Args)
	if err != nil {
		log.Error(err.Error())
		return
	}

	switch flagConfig.Mode {
	case "send":
		sender := transfer.NewSender(flagConfig.Host, flagConfig.Port, "tcp", flagConfig.FilePath)
		err := sender.Listen()
		if err != nil {
			log.Error(err.Error())
		}

	case "receive":
		receiver := transfer.NewReceiver()
		err := receiver.Connect(flagConfig.Host, flagConfig.Port, "tcp")
		if err != nil {
			log.Error(err.Error())
		}
	}
}
