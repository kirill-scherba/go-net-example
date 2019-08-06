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
#include <stdio.h>

#define ARP_TABLE_IP_SIZE 48    // INET6_ADDRSTRLEN = 46

/**
 * KSNet ARP table data structure
 */
typedef struct ksnet_arp_data {

    int16_t mode;                   ///< Peers mode: -1 - This host, -2 undefined host, 0 - peer , 1 - r-host, 2 - TCP Proxy peer
    char addr[ARP_TABLE_IP_SIZE];   ///< Peer IP address
// \todo test is it possible to change this structure for running peers
//    char addr[48];      ///< Peer IP address
    int16_t port;                   ///< Peer port

    double last_activity;           ///< Last time received data from peer
    double last_triptime_send;      ///< Last time when triptime request send
    double last_triptime_got;       ///< Last time when triptime received

    double last_triptime;           ///< Last triptime
    double triptime;                ///< Middle triptime

    double monitor_time;            ///< Monitor ping time

    double connected_time;          ///< Time when peer was connected to this peer

//    char *type;                     ///< Peer type

} ksnet_arp_data;

#pragma pack(push)
#pragma pack(1)

/**
 * KSNet ARP table whole data array
 */
typedef struct ksnet_arp_data_ar {

    uint32_t length;
    struct _arp_data {

        char name[ARP_TABLE_IP_SIZE];
        ksnet_arp_data data;

    } arp_data[];

} ksnet_arp_data_ar;

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


uint8_t get_byte_checksum(void *data, size_t data_length);
char *arp_data_print(ksnet_arp_data_ar *arp_data_ar);
int packetCheck(void *packetPtr, size_t packetLen);

size_t teoLNullHeaderSize();

size_t teoLNullPacketCreate(void *buffer, size_t buffer_length, uint8_t command,
                            const char *peer, const void *data,
                            size_t data_length);
size_t teoLNullPacketCreateEcho(void *buf, size_t buf_len,
                                const char *peer_name, const char *msg);
int64_t teoLNullProccessEchoAnswer(const char *msg);

#endif /* TEONET_L0_CLIENT_H */
