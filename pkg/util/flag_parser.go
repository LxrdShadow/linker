package util

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
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

	flag.Usage = appUsage
	flag.Parse()

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
	var addrConf string
	var err error

	if (!isEmptyString(*host) || !isEmptyString(*port)) && !isEmptyString(*addr) {
		return nil, fmt.Errorf("'%s' have to only come with 'addr' (host:port) or 'host' and 'port' \n", RECEIVE_FLAG)
	} else if !isEmptyString(*addr) {
		hostConf, portConf, err = getHostPortFromAddr(*addr)
		addrConf = *addr
		if err != nil {
			return nil, fmt.Errorf("failed to parse address: %w", err)
		}
	} else if isEmptyString(*addr) {
		if *host != "" {
			hostConf = *host
		} else {
			conf, err := getLocalHostAddress()
			if err != nil {
				return nil, err
			}
			hostConf = conf

		}

		if *port != "" {
			portConf = *port
		} else {
			portConf = strconv.Itoa(rand.Intn(64000) + 1000)
		}

		addrConf = getAddrFromHostPort(hostConf, portConf)
	}

	return &FlagConfig{
		Mode:     SEND_FLAG,
		FilePath: *file,
		Address:  addrConf,
		Host:     hostConf,
		Port:     portConf,
	}, nil
}

func getReceiveConfig(addr, host, port *string) (*FlagConfig, error) {
	var hostConf string
	var portConf string
	var addrConf string
	var err error

	if (isEmptyString(*host) || isEmptyString(*port)) && isEmptyString(*addr) {
		return nil, fmt.Errorf("'%s' have to come with an address (-addr host:port)\n", RECEIVE_FLAG)
	} else if (!isEmptyString(*host) || !isEmptyString(*port)) && !isEmptyString(*addr) {
		return nil, fmt.Errorf("'%s' have to only come with '-addr' (host:port) or '-host' and '-port' \n", RECEIVE_FLAG)
	} else if !isEmptyString(*addr) {
		addrConf = *addr
		hostConf, portConf, err = getHostPortFromAddr(*addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse address: %w", err)
		}
	} else {
		hostConf = *host
		portConf = *port
		addrConf = getAddrFromHostPort(hostConf, portConf)
	}

	return &FlagConfig{
		Mode:    RECEIVE_FLAG,
		Address: addrConf,
		Host:    hostConf,
		Port:    portConf,
	}, err
}

func getLocalHostAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("failed to get host IP address: %w\n", err)
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}

	return "", nil
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

func isEmptyString(str string) bool {
	return str == ""
}

func appUsage() {
	intro := `lnkr (linker) is a simple file transfer program.

Usage:
	lnkr <command> [command flags]`

	fmt.Fprintln(os.Stderr, intro)
	fmt.Fprintln(os.Stderr, "\nCommands:")
	fmt.Fprintf(os.Stderr, "\t%s\n", SEND_FLAG)
	fmt.Fprintln(os.Stderr, "\t\tcreates a server to send files")
	fmt.Fprintf(os.Stderr, "\t%s\n", RECEIVE_FLAG)
	fmt.Fprintln(os.Stderr, "\t\tjoin a send server to receive the files")

	// fmt.Fprintln(os.Stderr, "\nCommand Flags:")
	// fmt.Fprintf(os.Stderr, "\t--file  -file\n")
	// fmt.Fprintln(os.Stderr, "\t\tpath of the file to send")
	// fmt.Fprintf(os.Stderr, "\t--addr  -addr\n")
	// fmt.Fprintln(os.Stderr, "\t\taddres for the server (host:port)")
	// fmt.Fprintf(os.Stderr, "\t--host  -host\n")
	// fmt.Fprintln(os.Stderr, "\t\thost IP for the server")
	// fmt.Fprintf(os.Stderr, "\t--port  -port\n")
	// fmt.Fprintln(os.Stderr, "\t\tport for the server")

	// fmt.Fprintln(os.Stderr, "\nExample:")
	// fmt.Fprintln(os.Stderr, "\tlnkr send -file test.txt -addr 192.168.1.1:9090")
	// fmt.Fprintln(os.Stderr, "\tlnkr send --file=test.txt --host=192.168.1.1 --port=9090")
	// fmt.Fprintln(os.Stderr, "\tlnkr receive -host 192.168.1.1 -port 9090")
	// fmt.Fprintln(os.Stderr, "\tlnkr receive --addr=192.168.1.1:9090")

	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Run `lnkr <command> -h` to get help for a specific command\n\n")
}
