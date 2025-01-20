package transfer

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/LxrdShadow/linker/internal/protocol"
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
	address := fmt.Sprintf("%s:%s", s.Host, s.Port)
	listener, err := net.Listen(s.Network, address)
	if err != nil {
		return fmt.Errorf("Failed to listen on %s: %w", address, err)
	}

	fmt.Printf("Listening on %s\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection: %s", err.Error())
		}
		defer conn.Close()

		go s.SendSingleFile(conn)
	}
}

func (s *Sender) SendSingleFile(conn net.Conn) {
	response := make([]byte, 100)
	// conn.Read(response)

	// fmt.Println(string(response))
	fmt.Println("Connected with", conn.RemoteAddr().String())

	file, err := os.OpenFile(s.File, os.O_RDONLY, 0755)
	if err != nil {
		fmt.Printf("Error: failed to open file: %v\n", err.Error())
		return
	}
	defer file.Close()

	header, err := protocol.PrepareFileHeader(file)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	headerBuffer, err := header.Serialize()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		return
	}

	conn.Write(headerBuffer)

	_, err = conn.Read(response)
	if err != nil {
		fmt.Printf("Error: failed to read response: %v\n", err.Error())
	}
	fmt.Println(string(response))

	err = SendFileByChunks(conn, file, header)
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		return
	}

	_, err = conn.Read(response)
	if err != nil && err != io.EOF {
		fmt.Printf("Error: failed to read response: %v\n", err.Error())
	}
	fmt.Println(string(response))
}

func SendFileByChunks(conn net.Conn, file *os.File, header *protocol.Header) error {
	chunk := new(protocol.Chunk)
	dataBuffer := make([]byte, protocol.DATA_MAX_SIZE)
	// response := make([]byte, 100)

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

		// n, err = conn.Read(response)
		// if err != nil && err != io.EOF && n != 0 {
		// 	return fmt.Errorf("failed to read response: %w", err)
		// }
		// fmt.Println(string(response[:n]))
		// fmt.Println(n)
		ack := make([]byte, 1)
		if _, err := conn.Read(ack); err != nil && err != io.EOF {
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
