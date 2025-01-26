package config

const (
	RECEIVE_DIRECTORY = "./received/"
)

const (
	PROTOCOL_VERSION    = 1
	CHUNK_MIN_SIZE      = 4 + 8 // 4 bytes for SequenceNumber, 8 bytes for DataLength
	CHUNK_SIZE          = 65536 // 64 KB
	DATA_MAX_SIZE       = CHUNK_SIZE - CHUNK_MIN_SIZE
	MAX_FILENAME_LENGTH = 255
	HEADER_MIN_SIZE     = 19                                    // 19 bytes without filename
	HEADER_MAX_SIZE     = HEADER_MIN_SIZE + MAX_FILENAME_LENGTH // 274 bytes
)

