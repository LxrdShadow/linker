package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LxrdShadow/linker/internal/config"
	"github.com/LxrdShadow/linker/internal/errors"
)

type FileHeader struct {
	ChunkSize      uint32
	Reps           uint32
	FileSize       uint64
	FileNameLength uint16
	FileName       string
}

// Prepare the header with the informations about the file
func PrepareFileHeader(file *os.File, baseDir string) (*FileHeader, error) {
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

	header := &FileHeader{
		ChunkSize:      config.CHUNK_SIZE,
		Reps:           uint32(size/config.CHUNK_SIZE) + 1,
		FileSize:       uint64(size),
		FileNameLength: uint16(len(name)),
		FileName:       name,
	}

	return header, nil
}

// Encode the header to byte representation
func (h *FileHeader) Serialize() ([]byte, error) {
	if len(h.FileName) > config.MAX_FILENAME_LENGTH {
		return nil, fmt.Errorf("filename exceeds maximum length of %d bytes\n", config.MAX_FILENAME_LENGTH)
	}

	buff := new(bytes.Buffer)

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
func DeserializeHeader(data []byte) (*FileHeader, error) {
	if len(data) < config.FILE_HEADER_MIN_SIZE || len(data) > config.FILE_HEADER_MAX_SIZE {
		return nil, errors.InvalidHeaderSize
	}

	reader := bytes.NewReader(data)
	var header FileHeader

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
