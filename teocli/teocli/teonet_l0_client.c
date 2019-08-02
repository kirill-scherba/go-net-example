/**
 * This module contain part of functions from
 * original reocli teonet_lo_client.c file
 * Author: Kirill Scherba <kirill@scherba.ru>
 * Created on October 12, 2015, 12:32 PM
 *
 * This function used to create original teocli(teonet l0 client) packets
 */

#include <string.h>
#include "teonet_l0_client.h"

/**
 * Calculate checksum
 *
 * Calculate byte checksum in data buffer
 *
 * @param data Pointer to data buffer
 * @param data_length Length of the data buffer to calculate checksum
 *
 * @return Byte checksum of the input buffer
 */
uint8_t get_byte_checksum(void *data, size_t data_length)
{
    int i;
    uint8_t *ch, checksum = 0;
    for(i = 0; i < (int)data_length; i++) {

        ch = (uint8_t*)((char*)data + i);
        checksum += *ch;
    }

    return checksum;
}

/**
 * Return L0 header size
 */
size_t teoLNullHeaderSize() {
  return sizeof(teoLNullCPacket);
}


/**
 * Create L0 client packet
 *
 * @param buffer Buffer to create packet in
 * @param buffer_length Buffer length
 * @param command Command to peer
 * @param peer Teonet peer
 * @param data Command data
 * @param data_length Command data length
 *
 * @return Length of created packet or zero if buffer to less
 */
size_t teoLNullPacketCreate(void* buffer, size_t buffer_length, uint8_t command, const char * peer,
        const void* data, size_t data_length)
{
    size_t peer_name_length = strlen(peer) + 1;

    // Check buffer length
    if (buffer_length < sizeof(teoLNullCPacket) + peer_name_length + data_length) {
        return 0;
    }

    teoLNullCPacket* pkg = (teoLNullCPacket*) buffer;
    memset(buffer, 0, sizeof(teoLNullCPacket));

    pkg->cmd = command;
    pkg->data_length = (uint16_t)data_length;
    pkg->peer_name_length = (uint8_t)peer_name_length;
    memcpy(pkg->peer_name, peer, pkg->peer_name_length);
    memcpy(pkg->peer_name + pkg->peer_name_length, data, pkg->data_length);
    pkg->checksum = get_byte_checksum(pkg->peer_name, pkg->peer_name_length +
            pkg->data_length);
    pkg->header_checksum = get_byte_checksum(pkg, sizeof(teoLNullCPacket) -
            sizeof(pkg->header_checksum));

    return sizeof(teoLNullCPacket) + pkg->peer_name_length + pkg->data_length;
}
