// =============================================================================
//
//  file: line.h
//  auth: andrew@ardnew.com
//  info: Public API for class Line
//
// =============================================================================
#ifndef LINE_H
#define LINE_H

#ifndef ARDUINO
#include <string>
#endif

#include "fifo.h"

template<size_t maxLineBytes>
class Line: public Fifo<char, maxLineBytes> {
protected:
  size_t _pos; // cursor position in this line

  size_t append(const char *s) {
    size_t more = strnlen(s, this->_size);
    if (more == 0) {
      return 0;
    }
    size_t used = this->_tail - this->_head;
    if (used == this->_size) {
      return 0;
    }
    if (used+more > this->_size) {
      more = this->_size - used;
    }
    for (int i = 0; i < more; ++i) {
      this->_elem[this->_tail%this->_size] = s[i];
      ++this->_tail;
    }
    return more;
  }

public:
  Line(void):
    Fifo<char, maxLineBytes>(FifoDiscardMode::Last),
    _pos(0) {
    memset(this->_elem, 0, maxLineBytes);
  }
  Line(const std::string &str):
    Fifo<char, maxLineBytes>(FifoDiscardMode::Last) {
    memset(this->_elem, 0, maxLineBytes);
    size_t n = append(str.c_str());
    if (n == maxLineBytes) {
      this->set(-1, 0); // truncate
    }
  }
  Line(const char *s):
    Fifo<char, maxLineBytes>(FifoDiscardMode::Last) {
    memset(this->_elem, 0, maxLineBytes);
    size_t n = append(s);
    if (n == maxLineBytes) {
      this->set(-1, 0); // truncate
    }
  }
  Line(const char c):
    Fifo<char, maxLineBytes>(FifoDiscardMode::Last),
    _pos(0) {
    memset(this->_elem, 0, maxLineBytes);
    if (c != 0 && this->enq(c)) {
      _pos = 1;
    }
  }

  ~Line(void) {}
};

#endif // LINE_H
