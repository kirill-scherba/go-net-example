#include "command.h"
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// marshalClients convert binary client list data to json
char *marshalClients(void *data, size_t *data_len) {
  teonet_client_data_ar *client_data_ar = data;
  int data_str_len = KSN_BUFFER_SIZE + sizeof(client_data_ar->client_data[0]) *
                                           2 * client_data_ar->length;
  char *data_str = malloc(data_str_len);
  int ptr = sprintf(data_str, "{ \"length\": %d, \"client_data_ar\": [ ",
                    client_data_ar->length);
  int i = 0;
  for (i = 0; i < client_data_ar->length; i++) {
    ptr += sprintf(data_str + ptr,
                   "%s{ "
                   "\"name\": \"%s\""
                   " }",
                   i ? ", " : "", client_data_ar->client_data[i].name);
  }
  sprintf(data_str + ptr, " ] }");
  *data_len = strlen(data_str);
  return data_str;
}
