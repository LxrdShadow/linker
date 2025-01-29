package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/LxrdShadow/linker/internal/config"
	"github.com/LxrdShadow/linker/internal/errors"
)

type TransferHeader struct {
	Version byte
	Reps    uint16
	IsDir   []bool
}

// Prepare the header with the informations about the file and the protocol
func PrepareTransferHeader(entries []string) (*TransferHeader, error) {
	isDir := make([]bool, len(entries))

	for i, entry := range entries {
		info, err := os.Stat(entry)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			isDir[i] = true
		} else {
			isDir[i] = false
		}
	}

	header := &TransferHeader{
		Version: config.PROTOCOL_VERSION,
		Reps:    uint16(len(entries)),
		IsDir:   isDir,
	}

	return header, nil
}

// Encode the header to byte representation
func (th *TransferHeader) Serialize() ([]byte, error) {
	buff := new(bytes.Buffer)

	// Version
	if err := binary.Write(buff, binary.BigEndian, th.Version); err != nil {
		return nil, fmt.Errorf("failed to write version: %w\n", err)
	}

	// Number of entries to process
	if err := binary.Write(buff, binary.BigEndian, th.Reps); err != nil {
		return nil, fmt.Errorf("failed to write reps: %w\n", err)
	}

	// To check wether an entry is a directory or not
	isDirBytes := encodeBooleans(th.IsDir)
	buff.Write(isDirBytes)

	return buff.Bytes(), nil
}

// Decode a byte representation of a header to a TransferHeader struct
func DeserializeTransferHeader(data []byte) (*TransferHeader, error) {
	if len(data) < config.FILE_HEADER_MIN_SIZE {
		return nil, errors.InvalidHeaderSize
	}

	reader := bytes.NewReader(data)
	var header TransferHeader

	// Version
	if err := binary.Read(reader, binary.BigEndian, &header.Version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w\n", err)
	}

	if header.Version != config.PROTOCOL_VERSION {
		return nil, fmt.Errorf("protocol version mismatch: got v%d protocol while using v%d protocol\n", header.Version, config.PROTOCOL_VERSION)
	}

	// Number of entries to process
	if err := binary.Read(reader, binary.BigEndian, &header.Reps); err != nil {
		return nil, fmt.Errorf("failed to read reps: %w\n", err)
	}

	// To check wether an entry is a directory or not
	byteCount := (header.Reps + 7) / 8
	tmp := make([]byte, byteCount)
	reader.Read(tmp)
	// if err := binary.Read(reader, binary.BigEndian, &header.IsDir); err != nil {
	// 	return nil, fmt.Errorf("failed to read IsDir confirmations: %w\n", err)
	// }
	header.IsDir = decodeBooleans(tmp, int(header.Reps))

	if len(header.IsDir) != int(header.Reps) {
		return nil, fmt.Errorf("malformed transfer header\n")
	}

	return &header, nil
}

// Encode an array of bools to binary
func encodeBooleans(booleans []bool) []byte {
	// Pack 8 booleans into one byte
	byteCount := (len(booleans) + 7) / 8
	encoded := make([]byte, byteCount)

	for i, bool := range booleans {
		if bool {
			encoded[i/8] |= 1 << (7 - (i % 8))
		}
	}

	return encoded
}

// Decode binary to get an array of bools
func decodeBooleans(encoded []byte, length int) []bool {
	booleans := make([]bool, length)

	for i := 0; i < length; i++ {
		byteIndex := i / 8
		bitIndex := 7 - (i % 8)
		booleans[i] = (encoded[byteIndex] & (1 << bitIndex)) != 0
	}

	return booleans
}
