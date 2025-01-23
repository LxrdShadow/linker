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
	"github.com/LxrdShadow/linker/pkg/util"
)

type Sender struct {
	Host, Port, Network string
	File                string
}

func NewSender(host, port, network string, file string) *Sender {
	sender := &Sender{
		Host:    host,
		Port:    port,
		Network: network,
		File:    file,
	}

	return sender
}

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

	_, err = conn.Read(response)
	if err != nil {
		log.Errorf("failed to read response: %s\n", err.Error())
	}
	fmt.Println(string(response))

	err = SendFileByChunks(conn, file, header)
	if err != nil {
		log.Errorf("Error: %s\n", err.Error())
		return
	}

	_, err = conn.Read(response)
	if err != nil && errors.Is(err, io.EOF) {
		log.Errorf("failed to read response: %s\n", err.Error())
	}
	log.Success(string(response))
}

func SendFileByChunks(conn net.Conn, file *os.File, header *protocol.Header) error {
	chunk := new(protocol.Chunk)
	dataBuffer := make([]byte, protocol.DATA_MAX_SIZE)

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
		fmt.Printf("Sent chunk %d. Waiting for response...\t", chunk.SequenceNumber)

		ack := make([]byte, 1)
		_, err = conn.Read(ack)
		if err != nil && errors.Is(err, io.EOF) {
			return fmt.Errorf("failed to receive acknowledgment: %w", err)
		}

		if ack[0] != 1 {
			return fmt.Errorf("invalid acknowledgment received")
		}

		fmt.Printf("Chunk %d received\n", chunk.SequenceNumber)
	}

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
