#include <embedit.h>

void putc(const char c) { Serial.print(c); }

Embedit<> edit(putc);

void setup() {
}

void loop() {
}
