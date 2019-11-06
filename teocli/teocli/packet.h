#include <stdint.h>
#include "teonet_l0_client.h"

uint8_t packetGetCommand(void *packetPtr);
int packetGetPeerNameLength(teoLNullCPacket *packet);
int packetGetDataLength(void *packetPtr);
int packetGetLength(void *packetPtr);
char* packetGetPeerName(void *packetPtr);
char* packetGetData(void *packetPtr);