/**
 * This module contain part of functions definition and data structures from
 * original teocli teonet_lo_client.h file
 * author Kirill Scherba <kirill@scherba.ru>
 * Created on October 12, 2015, 12:32 PM
 *
 * This function definitions and structures used to create original
 * teocli(teonet l0 client) packets
 */
#ifndef TEONET_L0_CLIENT_H
#define TEONET_L0_CLIENT_H

#include <stdint.h>
#include <stdlib.h>

#pragma pack(push)
#pragma pack(1)

/**
 * L0 client packet data structure
 *
 */
typedef struct teoLNullCPacket {

  uint8_t cmd;              ///< Command
  uint8_t peer_name_length; ///< To peer name length (include leading zero)
  uint16_t data_length;     ///< Packet data length
  uint8_t reserved_1;       ///< Reserved 1
  uint8_t reserved_2;       ///< Reserved 2
  uint8_t checksum;         ///< Whole checksum
  uint8_t header_checksum;  ///< Header checksum
  char peer_name[]; ///< To/From peer name (include leading zero) + packet data

} teoLNullCPacket;

/**
 * L0 System commands
 */
enum CMD_L {

  CMD_L_ECHO = 65,              ///< #65 Echo command
  CMD_L_ECHO_ANSWER,            ///< #66 Answer to echo command
  CMD_L_PEERS = 72,             ///< #72 Get peers command
  CMD_L_PEERS_ANSWER,           ///< #73 Answer to get peers command
  CMD_L_AUTH = 77,              ///< #77 Auth command
  CMD_L_AUTH_ANSWER,            ///< #78 Auth answer command
  CMD_L_L0_CLIENTS,             ///< #79 Get clients list command
  CMD_L_L0_CLIENTS_ANSWER,      ///< #80 Clients list answer command
  CMD_L_SUBSCRIBE_ANSWER = 83,  ///< #83 Subscribe answer
  CMD_L_AUTH_LOGIN_ANSWER = 96, ///< #96 Auth server login answer

  CMD_L_END = 127
};

// Reset compiler warnings to previous state.
#if defined(TEONET_COMPILER_MSVC)
#pragma warning(pop)
#endif

#pragma pack(pop)

size_t teoLNullHeaderSize();

size_t teoLNullPacketCreate(void *buffer, size_t buffer_length, uint8_t command,
                            const char *peer, const void *data,
                            size_t data_length);
size_t teoLNullPacketCreateEcho(void *buf, size_t buf_len,
                                const char *peer_name, const char *msg);
int64_t teoLNullProccessEchoAnswer(const char *msg);

#endif /* TEONET_L0_CLIENT_H */
