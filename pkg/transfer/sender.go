package transfer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"

	"github.com/LxrdShadow/linker/internal/config"
	"github.com/LxrdShadow/linker/internal/protocol"
	"github.com/LxrdShadow/linker/pkg/color"
	"github.com/LxrdShadow/linker/pkg/log"
	"github.com/LxrdShadow/linker/pkg/progress"
	"github.com/LxrdShadow/linker/pkg/util"
)

type Sender struct {
	*Connection
	Entries []string
}

// Creates a new sender object
func NewSender(config *util.FlagConfig) *Sender {
	sender := &Sender{
		Connection: &Connection{
			Host:    config.Host,
			Port:    config.Port,
			Network: config.Network,
			Addr:    config.Addr,
		},
		Entries: config.Entries,
	}

	return sender
}

// Listens on the sender's host IP and port
func (s *Sender) Listen() error {
	listener, err := net.Listen(s.Network, s.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", color.Sprint(color.RED, s.Addr), err)
	}

	fmt.Printf("Listening on: %s\n", color.Sprint(color.GREEN, s.Addr))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("failed to accept connection: %s", err.Error())
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Sender) handleConnection(conn net.Conn) error {
	fmt.Println("Connected with", conn.RemoteAddr().String())
	fmt.Println()
	defer conn.Close()

	err := s.sendNumEntries(conn)

	for _, entry := range s.Entries {
		info, err := os.Stat(entry)
		if err != nil {
			log.Errorf("failed to send %s: %s", err.Error())
			continue
		}

		if info.IsDir() {
			// TODO: Handle Sending Directories
			err = s.sendDirectory(conn, entry)
		} else {
			err = s.sendSingleFile(conn, entry)
		}

		if err != nil {
			log.Errorf("failed to send %s: %s", err.Error())
			continue
		}
	}

	response := make([]byte, 50)
	_, err = conn.Read(response)
	if err != nil && errors.Is(err, io.EOF) {
		log.Errorf("failed to read response: %s\n", err.Error())
	}

	log.Successf("%s\n", string(response))
	fmt.Println()
	fmt.Printf("Listening on: %s\n", color.Sprint(color.GREEN, s.Addr))

	return nil
}

// Send the number of files for the client to receive
func (s *Sender) sendNumEntries(conn net.Conn) error {
	numEntriesBuff := make([]byte, 1)
	binary.PutUvarint(numEntriesBuff, uint64(len(s.Entries)))
	conn.Write(numEntriesBuff)

	ack := make([]byte, 1)
	_, err := conn.Read(ack)
	if err != nil && errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to receive acknowledgment: %w", err)
	}

	if ack[0] != 1 {
		return fmt.Errorf("invalid acknowledgment received")
	}

	return nil
}

// Send the single file specified in the app's flags
func (s *Sender) sendDirectory(conn net.Conn, dir string) error {
	baseDir := filepath.Dir(filepath.Clean(dir))
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", filepath.Base(path), err)
		}

		header, err := protocol.PrepareFileHeader(file, baseDir)
		if err != nil {
			return fmt.Errorf("failed to get file header: %w", err)
		}

		fmt.Printf("Header: %+v\n", header)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to send directory: %s: %w", dir, err)
	}

	return nil
}

// Send one file specified as argument
func (s *Sender) sendSingleFile(conn net.Conn, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w\n", err)
	}
	defer file.Close()

	header, err := protocol.PrepareFileHeader(file, "")
	if err != nil {
		return fmt.Errorf("failed to get file header: %w", err)
	}

	err = s.sendHeader(conn, header)
	if err != nil {
		return fmt.Errorf("failed to send header: %w", err)
	}

	err = s.sendFileByChunks(conn, file, header)
	if err != nil {
		return fmt.Errorf("failed to send file: %w", err)
	}

	return nil
}

func (s *Sender) sendFileByChunks(conn net.Conn, file *os.File, header *protocol.Header) error {
	chunk := new(protocol.Chunk)
	dataBuffer := make([]byte, config.DATA_MAX_SIZE)

	unit, denom := util.ByteDecodeUnit(header.FileSize)

	bar := progress.NewProgressBar(header.FileSize, '=', denom, header.FileName, unit)
	bar.Render()

	for i := 0; i < int(header.Reps); i++ {
		n, _ := file.ReadAt(dataBuffer, int64(i*len(dataBuffer)))

		chunk.SequenceNumber = uint32(i)
		chunk.DataLength = uint64(n)
		chunk.Data = dataBuffer

		s.sendChunk(conn, chunk)

		bar.AppendUpdate(uint64(n))
	}
	bar.Finish()
	fmt.Println()

	return nil
}

// Send the header for one file
func (s *Sender) sendHeader(conn net.Conn, header *protocol.Header) error {
	headerBuffer, err := header.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize header: %w", err)
	}

	conn.Write(headerBuffer)

	ack := make([]byte, 1)
	_, err = conn.Read(ack)
	if err != nil && errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to receive acknowledgment: %w", err)
	}

	if ack[0] != 1 {
		return fmt.Errorf("invalid acknowledgment received")
	}

	return nil
}

// Send one chunk of data
func (s *Sender) sendChunk(conn net.Conn, chunk *protocol.Chunk) error {
	ack := make([]byte, 1)
	chunkBuffer, err := chunk.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize chunk %d: %w", chunk.SequenceNumber, err)
	}

	_, err = conn.Write(chunkBuffer)
	if err != nil {
		return fmt.Errorf("failed to write chunk %d: %w", chunk.SequenceNumber, err)
	}

	_, err = conn.Read(ack)
	if err != nil && errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to receive acknowledgment: %w", err)
	}

	if ack[0] != 1 {
		return fmt.Errorf("invalid acknowledgment received")
	}

	return nil
}

func (s *Sender) sendHello(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Connected with:", conn.RemoteAddr().String())
	response := make([]byte, 100)
	conn.Read(response)

	fmt.Println(string(response))

	conn.Write([]byte("Hello world"))
}
