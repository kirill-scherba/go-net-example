#include "libgotst.h"
#include <stdio.h>
#include <stdlib.h>

int main(int argc, char **argv) {
  GoString name = {"Kirill", 6};
  char *str = GoTst(name);
  printf("%s\n", str);
  free(str);
  return 0;
}
