package tunnel

import "github.com/songgao/packets/ethernet"

// Ethertype redifined to add method Print
type Ethertype struct {
	ethernet.Ethertype
}

// String Ethertype representation
func (t Ethertype) String() (s string) {

	switch t.Ethertype {
	case ethernet.IPv4:
		s = "IPv4"
	case ethernet.ARP:
		s = "ARP"
	case ethernet.WakeOnLAN:
		s = "WakeOnLAN"
	case ethernet.TRILL:
		s = "TRILL"
	case ethernet.DECnetPhase4:
		s = "DECnetPhase4"
	case ethernet.RARP:
		s = "RARP"
	case ethernet.AppleTalk:
		s = "AppleTalk"
	case ethernet.AARP:
		s = "AARP"
	case ethernet.IPX1:
		s = "IPX1"
	case ethernet.IPX2:
		s = "IPX2"
	case ethernet.QNXQnet:
		s = "QNXQnet"
	case ethernet.IPv6:
		s = "IPv6"
	case ethernet.EthernetFlowControl:
		s = "EthernetFlowControl"
	case ethernet.IEEE802_3:
		s = "IEEE802_3"
	case ethernet.CobraNet:
		s = "CobraNet"
	case ethernet.MPLSUnicast:
		s = "MPLSUnicast"
	case ethernet.MPLSMulticast:
		s = "MPLSMulticast"
	case ethernet.PPPoEDiscovery:
		s = "MPLSMulticast"
	case ethernet.PPPoESession:
		s = "PPPoESession"
	case ethernet.JumboFrames:
		s = "JumboFrames"
	case ethernet.HomePlug1_0MME:
		s = "HomePlug1_0MME"
	case ethernet.IEEE802_1X:
		s = "IEEE802_1X"
	case ethernet.PROFINET:
		s = "PROFINET"
	case ethernet.HyperSCSI:
		s = "HyperSCSI"
	case ethernet.AoE:
		s = "AoE"
	case ethernet.EtherCAT:
		s = "EtherCAT"
	case ethernet.EthernetPowerlink:
		s = "EthernetPowerlink"
	case ethernet.LLDP:
		s = "LLDP"
	case ethernet.SERCOS3:
		s = "SERCOS3"
	case ethernet.WSMP:
		s = "WSMP"
	case ethernet.HomePlugAVMME:
		s = "HomePlugAVMME"
	case ethernet.MRP:
		s = "MRP"
	case ethernet.IEEE802_1AE:
		s = "IEEE802_1AE"
	case ethernet.IEEE1588:
		s = "IEEE802_1AE"
	case ethernet.IEEE802_1ag:
		s = "IEEE802_1ag"
	case ethernet.FCoE:
		s = "FCoE"
	case ethernet.FCoEInit:
		s = "FCoEInit"
	case ethernet.RoCE:
		s = "RoCE"
	case ethernet.CTP:
		s = "CTP"
	case ethernet.VeritasLLT:
		s = "VeritasLLT"
	}
	return
}
