#include "packet.h"

#define packet(packetPtr) ((teoLNullCPacket *)packetPtr)

uint8_t packetGetCommand(void *packetPtr) {
  return packet(packetPtr)->cmd;
}

int packetGetPeerNameLength(teoLNullCPacket *packet) {
  return packet->peer_name_length;
}

int packetGetDataLength(void *packetPtr) {
  return packet(packetPtr)->data_length;
}

int packetGetLength(void *packetPtr) {
	return teoLNullHeaderSize() + packetGetPeerNameLength(packetPtr) +
	  packetGetDataLength(packetPtr);
}

char* packetGetPeerName(void *packetPtr) {
  return packet(packetPtr)->peer_name;
}

char* packetGetData(void *packetPtr) {
  return packet(packetPtr)->peer_name + packet(packetPtr)->peer_name_length;
}
