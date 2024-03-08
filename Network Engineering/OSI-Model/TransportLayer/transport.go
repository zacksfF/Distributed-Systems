package transportlayer

import "errors"

// TCPHeader represents TCP header
type TCPHeader struct {
	SrcPort    int
	DstPort    int
	DataOffset int
	Reserved   int
	Flags      int
}

// UDPHeader represents UDP header
type UDPHeader struct {
	SrcPort int
	DstPort int
}

var (
	errShortTCPHeaderLength = errors.New("short TCP header length")
	errShortUDPHeaderLength = errors.New("short UDP header length")
)

func decodeTCP(b []byte) (TCPHeader, error) {
	if len(b) < 20 {
		return TCPHeader{}, errShortTCPHeaderLength
	}

	return TCPHeader{
		SrcPort:    int(b[0])<<8 | int(b[1]),
		DstPort:    int(b[2])<<8 | int(b[3]),
		DataOffset: int(b[12]) >> 4,
		Reserved:   0,
		Flags:      ((int(b[12])<<8 | int(b[13])) & 0x01ff),
	}, nil
}

func decodeUDP(b []byte) (UDPHeader, error) {
	if len(b) < 8 {
		return UDPHeader{}, errShortUDPHeaderLength
	}

	return UDPHeader{
		SrcPort: int(b[0])<<8 | int(b[1]),
		DstPort: int(b[2])<<8 | int(b[3]),
	}, nil
}
