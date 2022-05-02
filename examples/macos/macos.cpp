#include <iostream>

#include <embedit.h>

void putc(const char c) { fputc(c, stdout); }

Embedit<> edit(putc);

int main(int argc, char *argv[]) {
  std::cout << "size = " << edit.history.len() << std::endl;
  return 0;
}
