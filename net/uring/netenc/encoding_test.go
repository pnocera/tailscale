package netenc

import (
	"encoding/binary"
	"testing"
)

func TestNtohs(t *testing.T) {
	v := [2]byte{0xCD, 0xAB}
	big := binary.BigEndian.Uint16(v[:])
	if 0xABCD != Ntohs(big) {
		t.Errorf("ntohs failed, want %v, got %v", 0xABCD, Ntohs(big))
	}
}

func toBytes32(v uint32) [4]byte {
	out := [4]byte{}
	out[0] = byte(v & 0xFF)
	out[1] = byte((v >> 8) & 0xFF)
	out[2] = byte((v >> 16) & 0xFF)
	out[3] = byte((v >> 24) & 0xFF)
	return out
}

func TestHtonl(t *testing.T) {
	raw := uint32(0xDEADBEEF)

	networkOrder := Htonl(raw)
	bytes := toBytes32(networkOrder)
	fromBig := binary.BigEndian.Uint32(bytes[:])

	if fromBig != raw {
		t.Errorf("htonl failed, want %v, got %v", raw, fromBig)
	}
}

func toBytes16(v uint16) [2]byte {
	out := [2]byte{}
	out[0] = byte(v & 0xFF)
	out[1] = byte((v >> 8) & 0xFF)
	return out
}

func TestHtons(t *testing.T) {
	raw := uint16(0xBEEF)

	networkOrder := Htons(raw)
	bytes := toBytes16(networkOrder)
	fromBig := binary.BigEndian.Uint16(bytes[:])

	if fromBig != raw {
		t.Errorf("htonl failed, want %v, got %v", raw, fromBig)
	}
}
