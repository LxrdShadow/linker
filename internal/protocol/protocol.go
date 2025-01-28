package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/LxrdShadow/linker/internal/config"
	"github.com/LxrdShadow/linker/internal/errors"
)

type TransferHeader struct {
	Version byte
	Reps    uint16
	IsDir   []bool
}

type Header struct {
	Version        byte
	ChunkSize      uint32
	Reps           uint32
	FileSize       uint64
	FileNameLength uint16
	FileName       string
}

type Chunk struct {
	SequenceNumber uint32
	DataLength     uint64
	Data           []byte
}

type Packet interface {
	Serialize() ([]byte, error)
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

	fmt.Println(th.IsDir)

	return buff.Bytes(), nil
}

// Decode a byte representation of a header to a Header struct
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
	fmt.Println(header.IsDir)

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

// Prepare the header with the informations about the file and the protocol
func PrepareFileHeader(file *os.File, baseDir string) (*Header, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("Failed to get file info: %w\n", err)
	}
	size := fileInfo.Size()
	var name string

	if baseDir == "" {
		name = filepath.Base(file.Name())
	} else {
		name, err = filepath.Rel(baseDir, file.Name())
		if err != nil {
			return nil, fmt.Errorf("Failed to get file relative path: %w\n", err)
		}
	}

	header := &Header{
		Version:        config.PROTOCOL_VERSION,
		ChunkSize:      config.CHUNK_SIZE,
		Reps:           uint32(size/config.CHUNK_SIZE) + 1,
		FileSize:       uint64(size),
		FileNameLength: uint16(len(name)),
		FileName:       name,
	}

	return header, nil
}

// Encode the header to byte representation
func (h *Header) Serialize() ([]byte, error) {
	if len(h.FileName) > config.MAX_FILENAME_LENGTH {
		return nil, fmt.Errorf("filename exceeds maximum length of %d bytes\n", config.MAX_FILENAME_LENGTH)
	}

	buff := new(bytes.Buffer)

	// Version
	if err := binary.Write(buff, binary.BigEndian, h.Version); err != nil {
		return nil, fmt.Errorf("failed to write version: %w\n", err)
	}

	// Chunk size
	if err := binary.Write(buff, binary.BigEndian, h.ChunkSize); err != nil {
		return nil, fmt.Errorf("failed to write chunk size: %w\n", err)
	}

	// Chunk count (repetitions)
	if err := binary.Write(buff, binary.BigEndian, h.Reps); err != nil {
		return nil, fmt.Errorf("failed to write reps: %w\n", err)
	}

	// File size
	if err := binary.Write(buff, binary.BigEndian, h.FileSize); err != nil {
		return nil, fmt.Errorf("failed to write file size: %w\n", err)
	}

	// Length of the file name
	if err := binary.Write(buff, binary.BigEndian, h.FileNameLength); err != nil {
		return nil, fmt.Errorf("failed to write filename length: %w\n", err)
	}

	// The actual name of the file
	if _, err := buff.WriteString(h.FileName); err != nil {
		return nil, fmt.Errorf("failed to write filename: %w\n", err)
	}

	return buff.Bytes(), nil
}

// Decode a byte representation of a header to a Header struct
func DeserializeHeader(data []byte) (*Header, error) {
	if len(data) < config.FILE_HEADER_MIN_SIZE || len(data) > config.FILE_HEADER_MAX_SIZE {
		return nil, errors.InvalidHeaderSize
	}

	reader := bytes.NewReader(data)
	var header Header

	// Version
	if err := binary.Read(reader, binary.BigEndian, &header.Version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w\n", err)
	}

	// Chunk size
	if err := binary.Read(reader, binary.BigEndian, &header.ChunkSize); err != nil {
		return nil, fmt.Errorf("failed to read chunk size: %w\n", err)
	}

	// Chunk count (repetitions)
	if err := binary.Read(reader, binary.BigEndian, &header.Reps); err != nil {
		return nil, fmt.Errorf("failed to read reps: %w\n", err)
	}

	// File size
	if err := binary.Read(reader, binary.BigEndian, &header.FileSize); err != nil {
		return nil, fmt.Errorf("failed to read file size: %w\n", err)
	}

	// Length of the file name
	if err := binary.Read(reader, binary.BigEndian, &header.FileNameLength); err != nil {
		return nil, fmt.Errorf("failed to read filename length: %w\n", err)
	}

	if header.FileNameLength > config.MAX_FILENAME_LENGTH {
		return nil, fmt.Errorf("filename exceeds maximum length of %d bytes\n", config.MAX_FILENAME_LENGTH)
	}

	// The actual name of the file
	fileNameBytes := make([]byte, header.FileNameLength)
	if _, err := reader.Read(fileNameBytes); err != nil {
		return nil, fmt.Errorf("failed to read filename: %w\n", err)
	}
	header.FileName = string(fileNameBytes)

	return &header, nil
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
