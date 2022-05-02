// =============================================================================
//
//  file: fifo.h
//  auth: andrew@ardnew.com
//  info: Public API for class Fifo
//
// =============================================================================
#ifndef FIFO_H
#define FIFO_H

#ifdef ARDUINO
#include <Arduino.h>
#endif

enum class FifoDiscardMode { Last, First };

template<typename T, size_t N>
class Fifo {
public:

protected:
  FifoDiscardMode _mode;    // enqueue behavior if FIFO full
  volatile size_t _size;    // capacity of FIFO
  volatile size_t _head;    // oldest elem in FIFO is at index _head
  volatile size_t _tail;    // new elems are enqueued at index _tail
  T               _elem[N]; // statically-sized FIFO

public:
  Fifo(const FifoDiscardMode mode):
    _mode(mode),
    _size(N), 
    _head(0), 
    _tail(0) {}

  ~Fifo() {}

  void reset(int size) {
    if (size < 0 || (size_t)size > N) {
      size = (int)N;
    }
    _size = (size_t)size;
    _head = 0;
    _tail = 0;
  }

  inline int cap(void) { return (int)_size; }
  inline int len(void) { return (int)(_tail-_head); }
  inline int rem(void) { return cap() - len(); }

  bool deq(T &e) {
    if (_head == _tail) {
      return false; // empty FIFO
    }
    e = _elem[_head%_size];
    ++_head;
    return true;
  }

  bool enq(const T &e) {
    if (_tail-_head == _size) {
      // FIFO is full
      switch (_mode) {
      case FifoDiscardMode::Last:
        return false; // drop incoming data
      case FifoDiscardMode::First:
        ++_head;      // drop outgoing data
      }
    }
    _elem[_tail%_size] = e;
    ++_tail;
    return true;
  }

  bool front(T &e) {
    if (_head == _tail) {
      return false; // empty FIFO
    }
    e = _elem[_head%_size];
    return true;
  }

  bool back(T &e) {
    if (_head == _tail) {
      return false; // empty FIFO
    }
    e = _elem[(_tail-1)%_size];
    return true;
  }

  bool index(int &i) {
    int n = len();
    if (i < 0) {
      if (-i <= n) {
        i = ((int)_tail+i) % int(_size);
        return true;
      }
    } else {
      if (i < n) {
        i = ((int)_head+i) % int(_size);
        return true;
      }
    }
    return false;
  }

  bool get(const int i, T &e) {
    int n = i;
    if (index(n)) {
      e = _elem[n];
      return true;
    }
    return false;
  }

  bool set(const int i, const T &e) {
    int n = i;
    if (index(n)) {
      _elem[n] = e;
      return true;
    }
    return false;
  }
};

#endif // FIFO_H
