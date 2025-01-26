package transfer

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/LxrdShadow/linker/internal/protocol"
	// "github.com/LxrdShadow/linker/pkg/color"
	"github.com/LxrdShadow/linker/pkg/progress"
	"github.com/LxrdShadow/linker/pkg/util"
)

const (
	RECEIVED_DIRECTORY = "./received/"
)

type Receiver struct{}

// Creates a new receiver
func NewReceiver() *Receiver {
	return &Receiver{}
}

// Connect to a send server
func (s *Receiver) Connect(host, port, network string) error {
	address := fmt.Sprintf("%s:%s", host, port)
	// server, err := net.ResolveTCPAddr(network, address)
	// if err != nil {
	// 	return fmt.Errorf("Failed to resolve the address %s: %w\n", color.Sprint(color.RED, address), err)
	// }

	conn, err := net.Dial(network, address)
	if err != nil {
		return fmt.Errorf("Failed to dial the server: %w\n", err)
	}
	defer conn.Close()

	start := time.Now()
	err = handleIncomingData(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to handle request: %s\n", err)
		return err
	}

	fmt.Println(time.Since(start))

	return nil
}

func handleIncomingData(conn net.Conn) error {
	headerBuffer := make([]byte, protocol.HEADER_MAX_SIZE)

	_, err := conn.Read(headerBuffer)
	if err != nil {
		return fmt.Errorf("failed to read header: %w\n", err)
	}

	header, err := protocol.DeserializeHeader(headerBuffer)
	if err != nil {
		return fmt.Errorf("failed to deserialize header: %w\n", err)
	}

	conn.Write([]byte{1})

	if header.Version != protocol.PROTOCOL_VERSION {
		return fmt.Errorf("protocol version mismatch: got v%d protocol while using v%d protocol\n", header.Version, protocol.PROTOCOL_VERSION)
	}

	file, err := CreateDestFile(RECEIVED_DIRECTORY, header.FileName)
	defer file.Close()

	err = ReceiveFileByChunks(conn, file, header)
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

func ReceiveFileByChunks(conn net.Conn, file *os.File, header *protocol.Header) error {
	chunkBuffer := make([]byte, header.ChunkSize)
	var chunk *protocol.Chunk

	unit, denom := util.ByteDecodeUnit(header.FileSize)

	bar := progress.NewProgressBar(header.FileSize, '=', denom, header.FileName, unit)
	fmt.Println()
	bar.Render()

	for i := 0; i < int(header.Reps); i++ {
		n, err := io.ReadFull(conn, chunkBuffer)
		if err != nil && errors.Is(err, io.EOF) {
			return fmt.Errorf("failed to read data chunk: %w\n", err)
		}

		chunk, err = protocol.DeserializeChunk(chunkBuffer[:n])
		if err != nil {
			return fmt.Errorf("failed to deserialize chunk: %w\n", err)
		}

		if _, err := conn.Write([]byte{1}); err != nil {
			return fmt.Errorf("failed to send acknowledgment: %w", err)
		}

		bar.AppendUpdate(uint64(n))
		_, err = file.Write(chunk.Data)
		if err != nil {
			return fmt.Errorf("failed to write the data to the file: %w\n", err)
		}
	}
	bar.Finish()
	fmt.Println()

	return nil
}
