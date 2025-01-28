package transfer

import (
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

	transferHeader, err := protocol.PrepareTransferHeader(s.Entries)
	if err != nil {
		return fmt.Errorf("failed to prepare transfer header: %w", err)
	}

	// TODO: Send the transfer header instead of the number of entries
	err = s.sendPacket(conn, transferHeader)

	for i, entry := range s.Entries {
		if transferHeader.IsDir[i] {
			err = s.sendDirectory(conn, entry)
		} else {
			err = s.sendSingleFile(conn, entry, "")
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

// Send the single file specified in the app's flags
func (s *Sender) sendDirectory(conn net.Conn, dir string) error {
	baseDir := filepath.Dir(filepath.Clean(dir))

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		s.sendSingleFile(conn, path, baseDir)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to send directory: %s: %w", dir, err)
	}

	return nil
}

// Send one file specified as argument
func (s *Sender) sendSingleFile(conn net.Conn, filepath, baseDir string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w\n", err)
	}
	defer file.Close()

	header, err := protocol.PrepareFileHeader(file, baseDir)
	if err != nil {
		return fmt.Errorf("failed to get file header: %w", err)
	}

	err = s.sendPacket(conn, header)
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

		s.sendPacket(conn, chunk)

		bar.AppendUpdate(uint64(n))
	}
	bar.Finish()
	fmt.Println()

	return nil
}

// Send a packet (it could be a header or a chunk of data)
func (s *Sender) sendPacket(conn net.Conn, packet protocol.Packet) error {
	packetBuffer, err := packet.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize packet: %w", err)
	}

	conn.Write(packetBuffer)

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

func (s *Sender) sendHello(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Connected with:", conn.RemoteAddr().String())
	response := make([]byte, 100)
	conn.Read(response)

	fmt.Println(string(response))

	conn.Write([]byte("Hello world"))
}
