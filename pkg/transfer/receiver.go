package transfer

import (
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

	transferHeader, err := r.getTransferHeader(conn)
	if err != nil {
		return err
	}

	fmt.Println()
	// Loop over the number of entries sent by the server
	for i := range transferHeader.Reps {
		if transferHeader.IsDir[i] {
			err = r.receiveDirectory(conn, r.ReceiveDir)
		} else {
			err = r.receiveSingleFile(conn, r.ReceiveDir)
		}

		if err != nil {
			log.Errorf("failed to handle request: %v\n", err)
			continue
		}
	}

	time := time.Now().UTC().Format("Monday, 02-Jan-06 15:04:05 MST")
	log.Success(time)
	conn.Write([]byte(time))

	return nil
}

func (r *Receiver) receiveDirectory(conn net.Conn, receiveDir string) error {
	header, err := r.getDirHeader(conn)
	if err != nil {
		return err
	}

	for range header.Reps {
		err = r.receiveSingleFile(conn, receiveDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Receiver) receiveSingleFile(conn net.Conn, receiveDir string) error {
	header, err := r.getFileHeader(conn)
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
	path := filepath.Join(dir, filepath.Dir(filename))

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory: %w\n", err)
		}
	}

	var file *os.File
	filePath := filepath.Join(path, filepath.Base(filename))

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

func (r *Receiver) receiveFileByChunks(conn net.Conn, file *os.File, header *protocol.FileHeader) error {
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

func (r *Receiver) getTransferHeader(conn net.Conn) (*protocol.TransferHeader, error) {
	headerBuffer := make([]byte, config.FILE_HEADER_MAX_SIZE)

	_, err := conn.Read(headerBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w\n", err)
	}

	header, err := protocol.DeserializeTransferHeader(headerBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize header: %w\n", err)
	}

	if _, err := conn.Write([]byte{1}); err != nil {
		return nil, fmt.Errorf("failed to send acknowledgment: %w", err)
	}

	return header, nil
}

func (r *Receiver) getDirHeader(conn net.Conn) (*protocol.DirHeader, error) {
	headerBuffer := make([]byte, config.DIR_HEADER_SIZE)

	_, err := conn.Read(headerBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w\n", err)
	}

	header, err := protocol.DeserializeDirHeader(headerBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize header: %w\n", err)
	}

	if _, err := conn.Write([]byte{1}); err != nil {
		return nil, fmt.Errorf("failed to send acknowledgment: %w", err)
	}

	return header, nil
}

func (r *Receiver) getFileHeader(conn net.Conn) (*protocol.FileHeader, error) {
	headerBuffer := make([]byte, config.FILE_HEADER_MAX_SIZE)

	_, err := conn.Read(headerBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w\n", err)
	}

	header, err := protocol.DeserializeHeader(headerBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize header: %w\n", err)
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
