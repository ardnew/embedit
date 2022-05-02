// =============================================================================
//
//  file: embedit.h
//  auth: andrew@ardnew.com
//  info: Public API for class Embedit
//
// =============================================================================
#ifndef EMBEDIT_H
#define EMBEDIT_H

#include "history.h"

typedef void (*putcFunc)(const char c);

template<size_t maxLineBytes = 128, size_t maxHistLines = 64>
class Embedit {
protected:
  putcFunc _putc;

public:
  History<maxLineBytes, maxHistLines> history;

  Embedit(const putcFunc putc):
    _putc(putc) {}

  ~Embedit(void) {}

  void putc(const char c)  { if (_putc) { _putc(c); } }
  void puts(const char *s) { while (*s) { putc(*s++); } }

  void puts(const std::string &str) { puts(str.c_str()); }
};

#endif // EMBEDIT_H
