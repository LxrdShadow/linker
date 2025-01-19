package util

import (
	"flag"
	"fmt"
	"os"
)

type FlagConfig struct {
	Mode, FilePath string
}

func ParseFlags() (*FlagConfig, error) {
	mode := flag.String("mode", "", "Mode of the operation: 'send' or 'receive'\n")
	file := flag.String("file", "", "Path of the file to send")

	flag.Parse()

	if *mode == "" {
		return nil, fmt.Errorf("--mode is required ('send' or 'receive')\n")
	}

	if *mode != "send" && *mode != "receive" {
		return nil, fmt.Errorf("unknown mode: '%s'\n\t--mode have to be 'send' or 'receive'\n", *mode)
	}

	if *mode == "send" {
		if *file == "" {
			return nil, fmt.Errorf("--mode send have to come with a file\n")
		}

		if _, err := os.Stat(*file); os.IsNotExist(err) {
			return nil, fmt.Errorf("%s: no such file or directory", *file)
		}
	}

	config := &FlagConfig{Mode: *mode}
	if *mode == "send" {
		config.FilePath = *file
	}

	return config, nil
}
