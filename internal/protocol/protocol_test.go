package protocol

import (
	"os"
	"testing"
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

		if header.Version != PROTOCOL_VERSION {
			t.Errorf("Version mismatch: got %d want %d", header.Version, PROTOCOL_VERSION)
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

		if header.ChunkSize != CHUNK_SIZE {
			t.Errorf("Wrong chunk size: got %d want %d", header.ChunkSize, CHUNK_SIZE)
		}

		if header.Reps != 1 {
			t.Errorf("Wrong Chunk Number: got %d want %d", header.Reps, 1)
		}
	})
}
