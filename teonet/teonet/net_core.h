/**
 * File: net_core.h
 * Author: Kirill Scherba <kirill@scherba.ru>
 *
 * Created on August 8, 2019, 12:52
 */

#ifndef NET_CORE_H
#define NET_CORE_H

#include <stddef.h>
#include <stdint.h>

#define PACKET_HEADER_ADD_SIZE 2    // Sizeof from length + Sizeof command

/**
 * KSNet core received data structure
 */
typedef struct ksnCorePacketData {

    char *addr;             ///< Remote peer IP address
    int port;               ///< Remote peer port
    int mtu;                ///< Remote mtu
    char *from;             ///< Remote peer name
    uint8_t from_len;       ///< Remote peer name length

    uint8_t cmd;            ///< Command ID

    void *data;             ///< Received data
    size_t data_len;        ///< Received data length

    void *raw_data;         ///< Received packet data
    size_t raw_data_len;    ///< Received packet length

//    ksnet_arp_data_ext *arp;///< Pointer to extended ARP Table data

    int l0_f;               ///< L0 command flag (from set to l0 client name)

} ksnCorePacketData;

void *createPacketFrom(uint8_t cmd, char *from, size_t from_len,
       const void *data, size_t data_len, size_t *packet_len);
int parsePacket(void *packet, size_t packet_len, ksnCorePacketData *rd);

#endif // NET_CORE_H
