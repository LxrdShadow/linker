package transfer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/LxrdShadow/linker/internal/config"
	"github.com/LxrdShadow/linker/internal/protocol"
	"github.com/LxrdShadow/linker/pkg/log"
	"github.com/LxrdShadow/linker/pkg/progress"
	"github.com/LxrdShadow/linker/pkg/util"
)

type Receiver struct {
	*Connection
	ReceiveDir string
}

// Creates a new receiver
func NewReceiver(config *util.FlagConfig) *Receiver {
	return &Receiver{
		Connection: &Connection{
			Host:    config.Host,
			Port:    config.Port,
			Network: config.Network,
			Addr:    config.Addr,
		},
		ReceiveDir: config.ReceiveDir,
	}
}

// Connect to a send server
func (r *Receiver) Connect() error {
	conn, err := net.Dial(r.Network, r.Addr)
	if err != nil {
		return fmt.Errorf("Failed to dial the server: %w\n", err)
	}
	defer conn.Close()

	numEntries, err := r.readNumEntries(conn)
	if err != nil {
		return err
	}

	fmt.Println()
	// Loop over the number of entries sent by the server
	for range numEntries {
		err = r.handleIncomingData(conn, r.ReceiveDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to handle request: %s\n", err)
			continue
		}
	}

	time := time.Now().UTC().Format("Monday, 02-Jan-06 15:04:05 MST")
	log.Success(time)
	conn.Write([]byte(time))

	return nil
}

func (r *Receiver) readNumEntries(conn net.Conn) (uint8, error) {
	numEntriesBuff := make([]byte, 1)
	_, err := conn.Read(numEntriesBuff)
	if err != nil {
		return 0, fmt.Errorf("Failed to read the number of files: %w\n", err)
	}

	numEntries, _ := binary.Uvarint(numEntriesBuff)

	if _, err := conn.Write([]byte{1}); err != nil {
		return 0, fmt.Errorf("failed to send acknowledgment: %w", err)
	}

	return uint8(numEntries), nil
}

func (r *Receiver) handleIncomingData(conn net.Conn, receiveDir string) error {
	header, err := r.getHeader(conn)
	if err != nil {
		return err
	}

	file, err := r.createDestFile(receiveDir, header.FileName)
	defer file.Close()

	err = r.receiveFileByChunks(conn, file, header)
	if err != nil {
		return err
	}

	return nil
}

func (r *Receiver) createDestFile(dir, filename string) (*os.File, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory: %w\n", err)
		}
	}

	var file *os.File
	filePath := filepath.Join(dir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err = os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s: %w\n", filePath, err)
		}
	} else {
		file, err = os.OpenFile(filePath, os.O_WRONLY, 0755) // 0755 is the file permission in octal
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w\n", err)
		}
	}

	return file, nil
}

func (r *Receiver) receiveFileByChunks(conn net.Conn, file *os.File, header *protocol.Header) error {
	unit, denom := util.ByteDecodeUnit(header.FileSize)

	bar := progress.NewProgressBar(header.FileSize, '=', denom, header.FileName, unit)
	bar.Render()

	for i := 0; i < int(header.Reps); i++ {
		chunk, n, err := r.getChunk(conn, header.ChunkSize)
		if err != nil {
			return err
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

func (r *Receiver) getHeader(conn net.Conn) (*protocol.Header, error) {
	headerBuffer := make([]byte, config.HEADER_MAX_SIZE)

	_, err := conn.Read(headerBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w\n", err)
	}

	header, err := protocol.DeserializeHeader(headerBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize header: %w\n", err)
	}

	if header.Version != config.PROTOCOL_VERSION {
		return nil, fmt.Errorf("protocol version mismatch: got v%d protocol while using v%d protocol\n", header.Version, config.PROTOCOL_VERSION)
	}

	if _, err := conn.Write([]byte{1}); err != nil {
		return nil, fmt.Errorf("failed to send acknowledgment: %w", err)
	}

	return header, nil
}

func (r *Receiver) getChunk(conn net.Conn, size uint32) (*protocol.Chunk, int, error) {
	chunkBuffer := make([]byte, size)
	n, err := io.ReadFull(conn, chunkBuffer)
	if err != nil && errors.Is(err, io.EOF) {
		return nil, 0, fmt.Errorf("failed to read data chunk: %w\n", err)
	}

	chunk, err := protocol.DeserializeChunk(chunkBuffer[:n])
	if err != nil {
		return nil, 0, fmt.Errorf("failed to deserialize chunk: %w\n", err)
	}

	if _, err := conn.Write([]byte{1}); err != nil {
		return nil, 0, fmt.Errorf("failed to send acknowledgment: %w", err)
	}

	return chunk, n, nil
}
