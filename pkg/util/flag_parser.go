package util

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"

	"github.com/LxrdShadow/linker/internal/config"
)

type FlagConfig struct {
	Mode, Addr, Host, Port, Network, ReceiveDir string
	Entries                                        []string
}

const (
	HOST_COMMAND    = "send"
	CONNECT_COMMAND = "receive"
)

// Parse the flags given by the user
func ParseFlags(args []string) (*FlagConfig, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("expected '%s' or '%s' subcommands\n", HOST_COMMAND, CONNECT_COMMAND)
	}

	flag.Usage = appUsage
	flag.Parse()

	sendCmd := flag.NewFlagSet(HOST_COMMAND, flag.ExitOnError)
	// sendFile := sendCmd.String("file", "", "Path of the file to send")
	sendAddr := sendCmd.String("addr", "", "Address for the server (host:port)")
	sendHost := sendCmd.String("host", "", "Host IP for the server")
	sendPort := sendCmd.String("port", "", "Port for the server")

	receiveCmd := flag.NewFlagSet(CONNECT_COMMAND, flag.ExitOnError)
	receiveAddr := receiveCmd.String("addr", "", "Address of the server (host:port)")
	receiveHost := receiveCmd.String("host", "", "Host IP of the server")
	receivePort := receiveCmd.String("port", "", "Port of the server")
	receiveDir := receiveCmd.String("receive-dir", config.RECEIVE_DIRECTORY, "Directory to store the received files")

	var config *FlagConfig
	var err error

	switch args[1] {
	case HOST_COMMAND:
		sendCmd.Parse(args[2:])
		config, err = getSendConfig(sendCmd, sendAddr, sendHost, sendPort)

	case CONNECT_COMMAND:
		receiveCmd.Parse(args[2:])
		config, err = getReceiveConfig(receiveAddr, receiveHost, receivePort, receiveDir)
	}

	if err != nil {
		return nil, err
	}

	return config, nil
}

// Get the configurations for a send command
func getSendConfig(sendCmd *flag.FlagSet, addr, host, port *string) (*FlagConfig, error) {
	entries := sendCmd.Args()

	if len(entries) == 0 {
		return nil, fmt.Errorf("'%s' have to come with a file\n", HOST_COMMAND)
	}

	for _, entry := range entries {
		if _, err := os.Stat(entry); os.IsNotExist(err) {
			return nil, fmt.Errorf("%s: no such file or directory", entry)
		}
	}

	var hostConf string
	var portConf string
	var addrConf string
	var err error

	if (!isEmptyString(*host) || !isEmptyString(*port)) && !isEmptyString(*addr) {
		return nil, fmt.Errorf("'%s' have to only come with 'addr' (host:port) or 'host' and 'port' \n", CONNECT_COMMAND)
	} else if !isEmptyString(*addr) {
		hostConf, portConf, err = GetHostPortFromAddr(*addr)
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

		addrConf = GetAddrFromHostPort(hostConf, portConf)
	}

	return &FlagConfig{
		Network: "tcp",
		Mode:    HOST_COMMAND,
		Entries: entries,
		Addr: addrConf,
		Host:    hostConf,
		Port:    portConf,
	}, nil
}

// Get the configurations for a receive command
func getReceiveConfig(addr, host, port, receiveDir *string) (*FlagConfig, error) {
	var hostConf, portConf, addrConf, receiveDirConf string
	var err error

	if (isEmptyString(*host) || isEmptyString(*port)) && isEmptyString(*addr) {
		return nil, fmt.Errorf("'%s' have to come with an address (-addr host:port)\n", CONNECT_COMMAND)
	} else if (!isEmptyString(*host) || !isEmptyString(*port)) && !isEmptyString(*addr) {
		return nil, fmt.Errorf("'%s' have to only come with '-addr' (host:port) or '-host' and '-port' \n", CONNECT_COMMAND)
	} else if !isEmptyString(*addr) {
		addrConf = *addr
		hostConf, portConf, err = GetHostPortFromAddr(*addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse address: %w", err)
		}
	} else {
		hostConf = *host
		portConf = *port
		addrConf = GetAddrFromHostPort(hostConf, portConf)
	}

	if !isEmptyString(*receiveDir) {
		receiveDirConf = *receiveDir
	}

	return &FlagConfig{
		Network:    "tcp",
		Mode:       CONNECT_COMMAND,
		Addr:    addrConf,
		Host:       hostConf,
		Port:       portConf,
		ReceiveDir: receiveDirConf,
	}, err
}

// Get the local interface address for the current computer
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

// check if a string is empty (for readability)
func isEmptyString(str string) bool {
	return str == ""
}

// Custom usage message for the app
func appUsage() {
	intro := `lnkr (linker) is a simple file transfer program.

Usage:
	lnkr <command> [command flags] <FILES>`

	fmt.Fprintln(os.Stderr, intro)
	fmt.Fprintln(os.Stderr, "\nCommands:")
	fmt.Fprintf(os.Stderr, "\t%s\n", HOST_COMMAND)
	fmt.Fprintln(os.Stderr, "\t\tcreates a server to send files")
	fmt.Fprintf(os.Stderr, "\t%s\n", CONNECT_COMMAND)
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
