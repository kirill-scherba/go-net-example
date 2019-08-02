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
#define	TEONET_L0_CLIENT_H

#include <stdint.h>
#include <stdlib.h>

#pragma pack(push)
#pragma pack(1)

/**
 * L0 client packet data structure
 *
 */
typedef struct teoLNullCPacket {

    uint8_t cmd; ///< Command
    uint8_t peer_name_length; ///< To peer name length (include leading zero)
    uint16_t data_length; ///< Packet data length
    uint8_t reserved_1; ///< Reserved 1
    uint8_t reserved_2; ///< Reserved 2
    uint8_t checksum; ///< Whole checksum
    uint8_t header_checksum; ///< Header checksum
    char peer_name[]; ///< To/From peer name (include leading zero) + packet data

} teoLNullCPacket;

// Reset compiler warnings to previous state.
#if defined(TEONET_COMPILER_MSVC)
#pragma warning(pop)
#endif

#pragma pack(pop)

size_t teoLNullHeaderSize();

size_t teoLNullPacketCreate(void* buffer, size_t buffer_length, uint8_t command, const char * peer,
         const void* data, size_t data_length);

#endif	/* TEONET_L0_CLIENT_H */
