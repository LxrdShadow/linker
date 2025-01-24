package transfer

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/LxrdShadow/linker/internal/protocol"
	"github.com/LxrdShadow/linker/pkg/color"
	"github.com/LxrdShadow/linker/pkg/log"
	"github.com/LxrdShadow/linker/pkg/progress"
	"github.com/LxrdShadow/linker/pkg/util"
)

type Sender struct {
	Host, Port, Network string
	File                string
}

// Creates a new sender
func NewSender(host, port, network string, file string) *Sender {
	sender := &Sender{
		Host:    host,
		Port:    port,
		Network: network,
		File:    file,
	}

	return sender
}

// Listens on the sender's host IP and port
func (s *Sender) Listen() error {
	address := util.GetAddrFromHostPort(s.Host, s.Port)
	listener, err := net.Listen(s.Network, address)
	if err != nil {
		return fmt.Errorf("Failed to listen on %s: %w", color.Sprint(color.RED, address), err)
	}

	fmt.Printf("Listening on: %s\n", color.Sprint(color.GREEN, address))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("Failed to accept connection: %s", err.Error())
		}
		defer conn.Close()

		go s.SendSingleFile(conn)
	}
}

// Send the single file specified in the app's flags
func (s *Sender) SendSingleFile(conn net.Conn) {
	response := make([]byte, 100)

	fmt.Println("Connected with", conn.RemoteAddr().String())

	file, err := os.OpenFile(s.File, os.O_RDONLY, 0755)
	if err != nil {
		log.Errorf("failed to open file: %s\n", err.Error())
		return
	}
	defer file.Close()

	header, err := protocol.PrepareFileHeader(file)
	if err != nil {
		log.Error(err.Error())
		return
	}

	headerBuffer, err := header.Serialize()
	if err != nil {
		log.Error(err.Error())
		return
	}

	conn.Write(headerBuffer)

	ack := make([]byte, 1)
	_, err = conn.Read(ack)
	if err != nil && errors.Is(err, io.EOF) {
		log.Errorf("failed to receive acknowledgment: %w", err)
	}

	if ack[0] != 1 {
		log.Error("invalid acknowledgment received")
	}

	err = SendFileByChunks(conn, file, header)
	if err != nil {
		log.Error(err.Error())
		return
	}

	_, err = conn.Read(response)
	if err != nil && errors.Is(err, io.EOF) {
		log.Errorf("failed to read response: %s\n", err.Error())
	}

	log.Successf("%s\n", string(response))
}

func SendFileByChunks(conn net.Conn, file *os.File, header *protocol.Header) error {
	chunk := new(protocol.Chunk)
	dataBuffer := make([]byte, protocol.DATA_MAX_SIZE)

	unit, denom := util.ByteDecodeUnit(header.FileSize)

	bar := progress.NewProgressBar(header.FileSize, '=', denom, header.FileName, unit)
	fmt.Println()
	bar.Render()

	ack := make([]byte, 1)

	for i := 0; i < int(header.Reps); i++ {
		n, _ := file.ReadAt(dataBuffer, int64(i*len(dataBuffer)))

		chunk.SequenceNumber = uint32(i)
		chunk.DataLength = uint64(n)
		chunk.Data = dataBuffer

		chunkBuffer, err := chunk.Serialize()
		if err != nil {
			return fmt.Errorf("failed to serialize chunk %d: %w", chunk.SequenceNumber, err)
		}

		_, err = conn.Write(chunkBuffer)
		if err != nil {
			return fmt.Errorf("failed to write chunk %d: %w", chunk.SequenceNumber, err)
		}
		// fmt.Printf("Sent chunk %d. Waiting for response...\t", chunk.SequenceNumber)

		_, err = conn.Read(ack)
		if err != nil && errors.Is(err, io.EOF) {
			return fmt.Errorf("failed to receive acknowledgment: %w", err)
		}

		if ack[0] != 1 {
			return fmt.Errorf("invalid acknowledgment received")
		}

		bar.AppendUpdate(uint64(n))

		// fmt.Printf("Chunk %d received\n", chunk.SequenceNumber)
	}
	bar.Finish()
	fmt.Println()

	return nil
}

func (s *Sender) SendHello(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Connected with:", conn.RemoteAddr().String())
	response := make([]byte, 100)
	conn.Read(response)

	fmt.Println(string(response))

	conn.Write([]byte("Hello world"))
}
