/**
 * Teonet core C module
 *
 * Combine from methods of C teonet ver 0.3.xxx for using in teonet-go
 *
 * File: packet.c
 * Author: Kirill Scherba <kirill@scherba.ru>
 *
 * Created on August 8, 2019, 12:52
 */

#include "packet.h"
#include <stdlib.h> /* malloc */
#include <string.h> /* memcpy */

/**
 * Create teonet packet with from field from another host (Resend)
 *
 * This function created to resend messages between networks in multi network
 * application
 *
 * @param cmd Command ID
 * @param from From name
 * @param from_len From name length
 * @param data Pointer to data
 * @param data_len Data length
 * @param [out] packet_len Pointer to packet length
 *
 * @return Pointer to packet. Should be free after use.
 */
// void *ksnCoreCreatePacketFrom(ksnCoreClass *kc, uint8_t cmd,
void *createPacketFrom(uint8_t cmd, char *from, size_t from_len,
                       const void *data, size_t data_len, size_t *packet_len) {

  size_t ptr = 0;
  *packet_len = from_len + data_len + PACKET_HEADER_ADD_SIZE;
  void *packet = malloc(*packet_len);

  // Copy packet data
  *((uint8_t *)packet) = from_len;
  ptr += sizeof(uint8_t); // From name length
  memcpy(packet + ptr, from, from_len);
  ptr += from_len; // From name
  *((uint8_t *)packet + ptr) = cmd;
  ptr += sizeof(uint8_t); // Command
  memcpy(packet + ptr, data, data_len);
  ptr += data_len; // Data

  return packet;
}

/**
 * Parse received data to ksnCorePacketData structure
 *
 * @param packet Received packet
 * @param packet_len Received packet length
 * @param rd Pointer to ksnCoreRecvData structure
 */
// int ksnCoreParsePacket(void *packet, size_t packet_len, ksnCorePacketData
// *rd) {
int parsePacket(void *packet, size_t packet_len, ksnCorePacketData *rd) {

  size_t ptr = 0;
  int packed_valid = 0;

  // Raw Packet buffer and length
  rd->raw_data = packet;
  rd->raw_data_len = packet_len;

  rd->from_len = *((uint8_t *)packet); ptr += sizeof(rd->from_len); // From length
  if (rd->from_len &&
      rd->from_len + PACKET_HEADER_ADD_SIZE <= packet_len &&
      *((char *)(packet + ptr + rd->from_len - 1)) == '\0'
  ) {
    rd->from = (char *)(packet + ptr); ptr += rd->from_len; // From pointer
    if(strlen(rd->from) + 1 == rd->from_len) {

      rd->cmd = *((uint8_t *)(packet + ptr)); ptr += sizeof(rd->cmd); // Command ID
      rd->data = packet + ptr; // Data pointer
      rd->data_len = packet_len - ptr; // Data length

      packed_valid = 1;
    }
  }

  return packed_valid;
}
