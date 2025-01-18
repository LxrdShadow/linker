package transfer

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/LxrdShadow/linker/internal/protocol"
)

const (
	RECEIVED_DIRECTORY = "./received/"
)

type Receiver struct{}

func NewReceiver() *Receiver {
	return &Receiver{}
}

func (s *Receiver) Connect(host, port, network string) error {
	address := fmt.Sprintf("%s:%s", host, port)

	conn, err := net.Dial(network, address)
	if err != nil {
		return fmt.Errorf("Failed to dial the server: %w\n", err)
	}
	defer conn.Close()

	err = handleIncomingRequest(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to handle request: %s\n", err)
		return err
	}

	// response := make([]byte, 100)
	// conn.Read(response)

	// fmt.Println(string(response))

	return nil
}

func handleIncomingRequest(conn net.Conn) error {
	headerBuffer := make([]byte, protocol.HEADER_MAX_SIZE)

	_, err := conn.Read(headerBuffer)
	if err != nil {
		return fmt.Errorf("failed to read header: %w\n", err)
	}

	header, err := protocol.DeserializeHeader(headerBuffer)
	if err != nil {
		return fmt.Errorf("failed to deserialize header: %w\n", err)
	}

	conn.Write([]byte("header received"))

	file, err := CreateDestFile(RECEIVED_DIRECTORY, header.FileName)
	defer file.Close()
	defer conn.Close()

	err = ReceiveFileByChunks(conn, header, file)
	if err != nil {
		return err
	}

	time := time.Now().UTC().Format("Monday, 02-Jan-06 15:04:05 MST")
	conn.Write([]byte(time))

	return nil
}

func CreateDestFile(dir, filename string) (*os.File, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory: %w\n", err)
		}
	}

	var file *os.File
	filePath := dir + filename
	if _, err := os.Stat(RECEIVED_DIRECTORY + filename); os.IsNotExist(err) {
		file, err = os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %w\n", err)
		}
	} else {
		file, err = os.OpenFile(filePath, os.O_WRONLY, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w\n", err)
		}
	}

	return file, nil
}

func ReceiveFileByChunks(conn net.Conn, header *protocol.Header, file *os.File) error {
	dataBuffer := make([]byte, protocol.DATA_MAX_SIZE)
	var chunk *protocol.Chunk

	for i := 0; i < int(header.Reps); i++ {
		_, err := conn.Read(dataBuffer)
		if err != nil {
			return fmt.Errorf("failed to read data chunk: %w\n", err)
		}

		chunk, err = protocol.DeserializeChunk(dataBuffer)
		if err != nil {
			return fmt.Errorf("failed to deserialize chunk: %w\n", err)
		}

		conn.Write([]byte(fmt.Sprintf("chunk %d received", chunk.SequenceNumber)))

		n, err := file.Write(chunk.Data)
		if err != nil {
			return fmt.Errorf("failed to write the data to the file: %w\n", err)
		}
		fmt.Printf("%d bytes written", n)
	}

	return nil
}
