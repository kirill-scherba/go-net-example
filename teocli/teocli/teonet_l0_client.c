/**
 * This module contain part of functions from
 * original reocli teonet_lo_client.c file
 * Author: Kirill Scherba <kirill@scherba.ru>
 * Created on October 12, 2015, 12:32 PM
 *
 * This function used to create original teocli(teonet l0 client) packets
 */

#include "teonet_l0_client.h"
#include "teonet_time.h"
#include <string.h>

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
uint8_t get_byte_checksum(void *data, size_t data_length) {
  int i;
  uint8_t *ch, checksum = 0;
  for (i = 0; i < (int)data_length; i++) {
    ch = (uint8_t *)((char *)data + i);
    checksum += *ch;
  }
  return checksum;
}

char *arp_data_print(ksnet_arp_data_ar *arp_data_ar) {
  int i;
  char *buf;
  size_t ptr = 0, len = arp_data_ar->length;
  if (len) {
    buf = malloc(len * 100);
    for (i = 0; i < (int)arp_data_ar->length; i++) {
      ptr += sprintf(buf + ptr, "%3d %-12s(%2d)   %-15s   %d %8.3f ms\n", i+1,
                     arp_data_ar->arp_data[i].name,
                     arp_data_ar->arp_data[i].data.mode,
                     arp_data_ar->arp_data[i].data.addr,
                     arp_data_ar->arp_data[i].data.port,
                     arp_data_ar->arp_data[i].data.last_triptime);
    }
  } else {
    buf = malloc(1);
    buf[0] = 0;
  }

  return buf;
}

/**
 * Check packet
 *
 * @param data packetPtr to data buffer
 * @param packetLen Length of the data buffer to calculate checksum
 *
 * Check packet length and checksum
 * @return 0 - valid packet;
 */
int packetCheck(void *packetPtr, size_t packetLen) {
  if (!packetPtr || !packetLen) {
    return -2;
  }
  teoLNullCPacket *packet = (teoLNullCPacket *)packetPtr;
  size_t header_length = teoLNullHeaderSize();
  if (packetLen <= header_length) {
    return -2; // Packet less than header (it may be first or next part of
               // splitted packet)
  }
  uint8_t header_checksum = get_byte_checksum(
      packet, sizeof(teoLNullCPacket) - sizeof(packet->header_checksum));
  if (packet->header_checksum != header_checksum) {
    return -3; // Wrong header checksum (it may be next part of splitted packet)
  }
  if (packetLen <
      header_length + packet->peer_name_length + packet->data_length) {
    return -1; // Wrong packet size (or first part of splitted packet)
  }
  uint8_t checksum = get_byte_checksum(
      packet->peer_name, packet->peer_name_length + packet->data_length);
  // printf("checksum %d\n", checksum);
  return (packet->checksum != checksum); // 1 - wrong packet checksum
}

/**
 * Return L0 header size
 */
size_t teoLNullHeaderSize() { return sizeof(teoLNullCPacket); }

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
size_t teoLNullPacketCreate(void *buffer, size_t buffer_length, uint8_t command,
                            const char *peer, const void *data,
                            size_t data_length) {
  size_t peer_name_length = strlen(peer) + 1;

  // Check buffer length
  if (buffer_length <
      sizeof(teoLNullCPacket) + peer_name_length + data_length) {
    return 0;
  }

  teoLNullCPacket *pkg = (teoLNullCPacket *)buffer;
  memset(buffer, 0, sizeof(teoLNullCPacket));

  pkg->cmd = command;
  pkg->data_length = (uint16_t)data_length;
  pkg->peer_name_length = (uint8_t)peer_name_length;
  memcpy(pkg->peer_name, peer, pkg->peer_name_length);
  memcpy(pkg->peer_name + pkg->peer_name_length, data, pkg->data_length);
  pkg->checksum = get_byte_checksum(pkg->peer_name,
                                    pkg->peer_name_length + pkg->data_length);
  pkg->header_checksum = get_byte_checksum(
      pkg, sizeof(teoLNullCPacket) - sizeof(pkg->header_checksum));

  return sizeof(teoLNullCPacket) + pkg->peer_name_length + pkg->data_length;
}

/**
 * Create package for Echo command
 * @param buf Buffer to create packet in
 * @param buf_len Buffer length
 * @param peer_name Peer name to send to
 * @param msg Echo message
 * @return
 */
size_t teoLNullPacketCreateEcho(void *buf, size_t buf_len,
                                const char *peer_name, const char *msg) {
  int64_t current_time = teotimeGetCurrentTime();

  unsigned int time_length = sizeof(current_time);

  const size_t msg_len = strlen(msg) + 1;
  const size_t msg_buf_len = msg_len + time_length;
  void *msg_buf = malloc(msg_buf_len);

  // Fill message buffer
  memcpy(msg_buf, msg, msg_len);
  memcpy((char *)msg_buf + msg_len, &current_time, time_length);
  size_t package_len = teoLNullPacketCreate(buf, buf_len, CMD_L_ECHO, peer_name,
                                            msg_buf, msg_buf_len);

  free(msg_buf);

  return package_len;
}

/**
 * Process ECHO_ANSWER request.(Get time from answers data and calculate trip
 * time)
 *
 * @param msg Echo answers command data
 * @return Trip time in ms
 */
int64_t teoLNullProccessEchoAnswer(const char *msg) {
  size_t time_ptr = strlen(msg) + 1;

  const int64_t *time_pointer = (const int64_t *)(msg + time_ptr);
  int64_t time_value = *time_pointer;

  int64_t trip_time = teotimeGetTimePassed(time_value);

  return trip_time;
}
