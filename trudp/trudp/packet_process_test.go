package trudp

import (
	"testing"
)

func BenchmarkRoomPacketDistance(b *testing.B) {
	pac := &packetType{}
	pac.packetDistance(120, 122)
}
