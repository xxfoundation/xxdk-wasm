#include "textflag.h"

// Throw enables throwing of Javascript exceptions.
TEXT ·Throw(SB), NOSPLIT, $0
  CallImport
  RET
