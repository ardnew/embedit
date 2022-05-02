// =============================================================================
//
//  file: history.h
//  auth: andrew@ardnew.com
//  info: Public API for class History
//
// =============================================================================
#ifndef HISTORY_H
#define HISTORY_H

#include "line.h"

template<size_t maxLineBytes, size_t maxHistLines>
class History: public Fifo<Line<maxLineBytes>, maxHistLines> {
public:
  History(void):
    Fifo<Line<maxLineBytes>, maxHistLines>(FifoDiscardMode::First) {}
  ~History(void) {}
};

#endif // HISTORY_H
