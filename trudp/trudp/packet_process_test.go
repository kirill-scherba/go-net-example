package trudp

import (
	"testing"
)

func BenchmarkPacketType_PacketDistance(b *testing.B) {
	pac := &packetType{}
	for i := 0; i < b.N; i++ {
		pac.packetDistance(uint32(i), 122)
	}
}

func BenchmarkPacketType_PacketDistanceSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if uint32(i) < 122 {
			//
		}
	}
}

func BenchmarkChannelData_GetId(b *testing.B) {
	tcd := &ChannelData{}
	for i := 0; i < b.N; i++ {
		if i != int(tcd.getID()) {
			b.FailNow()
		}
	}
}

func TestChannelData_GetId(t *testing.T) {
	tcd := &ChannelData{}
	for i := 0; i < packetIDlimit/10000; i++ {
		if i != int(tcd.getID()) {
			t.Errorf("wrong id")
		}
	}
}

func TestPacketType_PacketDistance(t *testing.T) {
	type fields struct {
		trudp      *TRUDP
		data       []byte
		sendQueueF bool
		destoryF   bool
	}
	type args struct {
		expectedID uint32
		id         uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
		{"", fields{}, args{2, 1}, -1},
		{"", fields{}, args{2, 2}, 0},
		{"", fields{}, args{2, 3}, 1},
		{"", fields{}, args{220, 520}, 300},
		{"", fields{}, args{520, 220}, -300},
		{"", fields{}, args{packetIDlimit - 10, 1}, 11},
		{"", fields{}, args{1, packetIDlimit - 10}, -11},
		{"", fields{}, args{1, packetIDlimit - 1024}, -1025},
		{"", fields{}, args{1, packetIDlimit / 3}, packetIDlimit/3 - 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pac := &packetType{
				trudp:      tt.fields.trudp,
				data:       tt.fields.data,
				sendQueueF: tt.fields.sendQueueF,
				destoryF:   tt.fields.destoryF,
			}
			if got := pac.packetDistance(tt.args.expectedID, tt.args.id); got != tt.want {
				t.Errorf("packetType.packetDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}
