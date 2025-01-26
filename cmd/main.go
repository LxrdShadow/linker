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
		sender := transfer.NewSender(flagConfig)
		err := sender.Listen()
		if err != nil {
			log.Error(err.Error())
		}

	case "receive":
		receiver := transfer.NewReceiver(flagConfig)
		err := receiver.Connect()
		if err != nil {
			log.Error(err.Error())
		}
	}
}
