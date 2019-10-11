#include "command.h"
#include "base64.h"
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// marshalClients convert binary client list data to json
char *marshalClients(void *data, size_t *ret_data_len) {
  teonet_client_data_ar *client_data_ar = data;
  int data_str_len = KSN_BUFFER_SIZE + sizeof(client_data_ar->client_data[0]) *
                                           2 * client_data_ar->length; 
  char *data_str = malloc(data_str_len);
  int ptr = sprintf(data_str, "{\"length\":%d,\"client_data_ar\":[",
                    client_data_ar->length);
  int i = 0;
  for (i = 0; i < client_data_ar->length; i++) {
    ptr += sprintf(data_str + ptr,
                   "%s{"
                   "\"name\":\"%s\""
                   "}",
                   i ? "," : "", client_data_ar->client_data[i].name);
  }
  sprintf(data_str + ptr, "]}");
  *ret_data_len = strlen(data_str);
  return data_str;
}

// marshalSubscribe convert binary subscribe answer data to json
char *marshalSubscribe(void *data, size_t in_data_len, size_t *ret_data_len) {
  size_t data_len = in_data_len - sizeof(teoSScrData);
  teoSScrData *sscr_data = data;
  size_t b64_data_len;
  char *b64_data = ksn_base64_encode((const unsigned char *)sscr_data->data,
                                     data_len, &b64_data_len);
  int data_str_len = KSN_BUFFER_SIZE + b64_data_len;
  char *data_str = malloc(data_str_len);
  sprintf(data_str, "{\"ev\":%d,\"cmd\":%d,\"data\":\"%s\"}", sscr_data->ev,
          sscr_data->cmd, b64_data);
  free(b64_data);
  *ret_data_len = strlen(data_str);
  return data_str;
}
