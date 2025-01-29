package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type DirHeader struct {
	Reps uint32
}

// Prepare the header with the informations about the directory
func PrepareDirHeader(reps int) *DirHeader {
	header := &DirHeader{
		Reps: uint32(reps),
	}

	return header
}

// Encode a DirHeader struct to its byte representation
func (dh *DirHeader) Serialize() ([]byte, error) {
	buff := new(bytes.Buffer)

	// Number of files in the directory
	if err := binary.Write(buff, binary.BigEndian, dh.Reps); err != nil {
		return nil, fmt.Errorf("failed to write reps: %w\n", err)
	}

	return buff.Bytes(), nil
}

// Decode a byte representation of a header to a DirHeader struct
func DeserializeDirHeader(data []byte) (*DirHeader, error) {
	reader := bytes.NewReader(data)
	var header DirHeader

	// Number of files in the directory
	if err := binary.Read(reader, binary.BigEndian, &header.Reps); err != nil {
		return nil, fmt.Errorf("failed to read reps: %w\n", err)
	}

	return &header, nil
}
