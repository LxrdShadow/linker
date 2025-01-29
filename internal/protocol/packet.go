package protocol

type Packet interface {
	Serialize() ([]byte, error)
}
