package config

const (
	RECEIVE_DIRECTORY = "."
)

const (
	PROTOCOL_VERSION         = 1
	CHUNK_MIN_SIZE           = 4 + 8 // 4 bytes for SequenceNumber, 8 bytes for DataLength
	CHUNK_SIZE               = 65536 // 64 KB
	MAX_ENTRY_COUNT          = 65536
	DATA_MAX_SIZE            = CHUNK_SIZE - CHUNK_MIN_SIZE
	DIR_HEADER_SIZE          = 4
	TRANSFER_HEADER_MIN_SIZE = 1 + 2 // Version + Reps
	TRANSFER_HEADER_MAX_SIZE = TRANSFER_HEADER_MIN_SIZE + MAX_ENTRY_COUNT
	MAX_FILENAME_LENGTH      = 255
	FILE_HEADER_MIN_SIZE     = 4 + 4 + 8 + 2                              // 18 bytes without filename
	FILE_HEADER_MAX_SIZE     = FILE_HEADER_MIN_SIZE + MAX_FILENAME_LENGTH // 274 bytes
)
