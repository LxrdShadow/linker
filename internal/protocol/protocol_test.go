package protocol

import (
	"os"
	"reflect"
	"testing"

	"github.com/LxrdShadow/linker/internal/config"
)

// Guys, I have no idea how to test this T_T

func TestPrepareFileHeader(t *testing.T) {
	filename := "hello.txt"
	file, _ := os.Create(filename)
	file.Write([]byte("Hello world!"))
	defer file.Close()
	defer os.Remove(filename)

	t.Run("prepare the headers for a test file", func(t *testing.T) {
		header, _ := PrepareFileHeader(file)

		if header.Version != config.PROTOCOL_VERSION {
			t.Errorf("Version mismatch: got %d want %d", header.Version, config.PROTOCOL_VERSION)
		}

		if header.FileNameLength != 9 {
			t.Errorf("Wrong length: got %d want %d", header.FileNameLength, 9)
		}

		if header.FileName != filename {
			t.Errorf("Wrong filename: got %s want %s", header.FileName, filename)
		}

		info, _ := file.Stat()
		if header.FileSize != uint64(info.Size()) {
			t.Errorf("Wrong size: got %d want %d", header.FileSize, uint64(info.Size()))
		}

		if header.ChunkSize != config.CHUNK_SIZE {
			t.Errorf("Wrong chunk size: got %d want %d", header.ChunkSize, config.CHUNK_SIZE)
		}

		if header.Reps != 1 {
			t.Errorf("Wrong Chunk Number: got %d want %d", header.Reps, 1)
		}
	})
}

func TestSerializeHeader(t *testing.T) {
	header := &Header{
		Version:        config.PROTOCOL_VERSION,
		ChunkSize:      1024,
		Reps:           10,
		FileSize:       1024 * 10,
		FileNameLength: 9,
		FileName:       "hello.txt",
	}

	buff, _ := header.Serialize()
	got, _ := DeserializeHeader(buff)

	assertEqual(t, got, header)
}

func TestSerializeChunk(t *testing.T) {
	chunk := &Chunk{
		SequenceNumber: 2,
		DataLength:     2,
		Data:           []byte{byte(1), byte(2)},
	}

	buff, _ := chunk.Serialize()
	got, _ := DeserializeChunk(buff)

	assertEqual(t, got, chunk)
}

func assertEqual(t *testing.T, got any, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("value mismatch: got %+v, want %+v", got, want)
	}
}
