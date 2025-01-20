package util

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

type FlagConfig struct {
	Mode, FilePath, Address, Host, Port string
}

const (
	SEND_FLAG    = "send"
	RECEIVE_FLAG = "receive"
)

func ParseFlags(args []string) (*FlagConfig, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("expected '%s' or '%s' subcommands\n", SEND_FLAG, RECEIVE_FLAG)
	}

	sendCmd := flag.NewFlagSet(SEND_FLAG, flag.ExitOnError)
	sendFile := sendCmd.String("file", "", "Path of the file to send")
	sendAddr := sendCmd.String("addr", "", "Address for the server (host:port)")
	sendHost := sendCmd.String("host", "", "Host IP for the server")
	sendPort := sendCmd.String("port", "", "Port for the server")

	receiveCmd := flag.NewFlagSet(RECEIVE_FLAG, flag.ExitOnError)
	receiveAddr := receiveCmd.String("addr", "", "Address of the server (host:port)")
	receiveHost := receiveCmd.String("host", "", "Host IP of the server")
	receivePort := receiveCmd.String("port", "", "Port of the server")

	var config *FlagConfig
	var err error

	switch args[1] {
	case SEND_FLAG:
		sendCmd.Parse(args[2:])
		config, err = getSendConfig(sendFile, sendAddr, sendHost, sendPort)

	case RECEIVE_FLAG:
		receiveCmd.Parse(args[2:])
		config, err = getReceiveConfig(receiveAddr, receiveHost, receivePort)
	}

	if err != nil {
		return nil, err
	}

	return config, nil
}

func getSendConfig(file, addr, host, port *string) (*FlagConfig, error) {
	if *file == "" {
		return nil, fmt.Errorf("'%s' have to come with a file\n", SEND_FLAG)
	}

	if _, err := os.Stat(*file); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s: no such file or directory", *file)
	}

	var hostConf string
	var portConf string
	var err error

	if (*host == "" || *port == "") && *addr == "" {
		return nil, fmt.Errorf("'%s' have to come with an address (host:port)\n", RECEIVE_FLAG)
	} else if (*host != "" || *port != "") && *addr != "" {
		return nil, fmt.Errorf("'%s' have to only come with 'addr' (host:port) or 'host' and 'port' \n", RECEIVE_FLAG)
	} else if *addr != "" {
		hostConf, portConf, err = getHostPortFromAddr(*addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse address: %w", err)
		}
	} else if *host != "" && *port != "" {
		hostConf = *host
		portConf = *port
	} else if *host == "" && *port == "" {
		conf, err := getLocalHostAddress()
		if err != nil {
			return nil, err
		}
		hostConf = conf
		portConf = "6969"
	}

	return &FlagConfig{
		Mode:     SEND_FLAG,
		FilePath: *file,
		Address:  getAddrFromHostPort(*host, *port),
		Host:     hostConf,
		Port:     portConf,
	}, nil
}

func getReceiveConfig(addr, host, port *string) (*FlagConfig, error) {
	var hostConf string
	var portConf string
	var err error

	if (*host == "" || *port == "") && *addr == "" {
		return nil, fmt.Errorf("'%s' have to come with an address (host:port)\n", RECEIVE_FLAG)
	} else if (*host != "" || *port != "") && *addr != "" {
		return nil, fmt.Errorf("'%s' have to only come with 'addr' (host:port) or 'host' and 'port' \n", RECEIVE_FLAG)
	} else if *addr != "" {
		hostConf, portConf, err = getHostPortFromAddr(*addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse address: %w", err)
		}
	}

	return &FlagConfig{
		Mode:    RECEIVE_FLAG,
		Address: getAddrFromHostPort(*host, *port),
		Host:    hostConf,
		Port:    portConf,
	}, err
}

func getLocalHostAddress() (string, error) {
	var host string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("failed to get host IP address: %w\n", err)
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			host = ipNet.IP.String()
		}
	}

	return host, nil
}

func getAddrFromHostPort(host, port string) string {
	return fmt.Sprintf("%s:%s", host, port)
}

func getHostPortFromAddr(addr string) (string, string, error) {
	if len(strings.Split(addr, ":")) != 2 {
		return "", "", fmt.Errorf("%s: wrong address format, it should be host:port\n", addr)
	}

	host := strings.Split(addr, ":")[0]
	port := strings.Split(addr, ":")[1]

	return host, port, nil
}
