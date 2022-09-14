#include "textflag.h"

// Throw enables throwing of Javascript exceptions.
TEXT ·throw(SB), NOSPLIT, $0
  CallImport
  RET
