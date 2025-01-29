package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/LxrdShadow/linker/internal/config"
	"github.com/LxrdShadow/linker/internal/errors"
)

type Chunk struct {
	SequenceNumber uint32
	DataLength     uint64
	Data           []byte
}

// Encode the chunk to byte representation
func (ch *Chunk) Serialize() ([]byte, error) {
	buff := new(bytes.Buffer)
	if err := binary.Write(buff, binary.BigEndian, ch.SequenceNumber); err != nil {
		return nil, fmt.Errorf("failed to write chunk sequence number: %w\n", err)
	}

	if err := binary.Write(buff, binary.BigEndian, ch.DataLength); err != nil {
		return nil, fmt.Errorf("failed to write chunk data length: %w\n", err)
	}

	if _, err := buff.Write(ch.Data); err != nil {
		return nil, fmt.Errorf("failed to write chunk data: %w\n", err)
	}

	return buff.Bytes(), nil
}

// Decode a byte representation of a header to a Header struct
func DeserializeChunk(data []byte) (*Chunk, error) {
	if len(data) < config.CHUNK_MIN_SIZE {
		return nil, errors.InvalidChunkSize
	}

	// fmt.Println(binary.BigEndian.Uint64(data[4:12]))

	reader := bytes.NewReader(data)
	var chunk Chunk

	// Sequence Number
	if err := binary.Read(reader, binary.BigEndian, &chunk.SequenceNumber); err != nil {
		return nil, fmt.Errorf("failed to read chunk sequence number: %w\n", err)
	}

	// Data Length
	if err := binary.Read(reader, binary.BigEndian, &chunk.DataLength); err != nil {
		return nil, fmt.Errorf("failed to read chunk data length: %w\n", err)
	}
	// fmt.Println("length:", chunk.DataLength)
	// fmt.Println("data:", len(chunk.Data))

	// Data
	if uint64(len(data)) < uint64(chunk.DataLength+config.CHUNK_MIN_SIZE) {
		return nil, fmt.Errorf("not enough data to read the chunk data: got %d want %d", len(data), chunk.DataLength+config.CHUNK_MIN_SIZE)
	}
	chunk.Data = make([]byte, chunk.DataLength)

	if err := binary.Read(reader, binary.BigEndian, &chunk.Data); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read chunk data: %w\n", err)
	}

	return &chunk, nil
}
