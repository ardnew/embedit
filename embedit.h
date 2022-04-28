// =============================================================================
//
//  file: embedit.h
//  auth: andrew@ardnew.com
//  info: Public API for class Embedit
//
// =============================================================================
#ifndef EMBEDIT_H
#define EMBEDIT_H

#ifdef ARDUINO
#include <Arduino.h>
#endif

template<typename T, size_t N>
class Queue {
private:
  size_t _size;    // capacity of FIFO
  size_t _head;    // oldest elem in FIFO is at index _head
  size_t _tail;    // new elems are enqueued at index _tail
  T      _elem[N]; // statically-sized FIFO
public:
  Queue(): 
    _size(N), 
    _head(0), 
    _tail(0) {
  }
  ~Queue() {}
};

template<size_t maxLineBytes = 128, size_t maxHistLines = 64>
class Embedit {
private:
  Queue<Queue<char, maxLineBytes>, maxHistLines> _history;
public:
  Embedit(void) {
  }
  ~Embedit(void) {
  }

};


#endif // EMBEDIT_H
