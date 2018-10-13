#include <M5Stack.h>

#define READ_PIN 21
#define IRDATA_SIZE 512

void setup() {
  M5.begin();
  Serial.begin(115200);
  pinMode(READ_PIN, INPUT);
}

void loop() {
  unsigned long usec, irOffTime;
  int i, j;
  unsigned int irdata[IRDATA_SIZE];

  while (digitalRead(READ_PIN) == HIGH);
  for (i=0; i < IRDATA_SIZE;) {
    usec = micros();
    while (digitalRead(READ_PIN) == LOW);
    irdata[i] = micros() - usec;
    irdata[++i] = 0;
    usec = micros();
    while (digitalRead(READ_PIN) == HIGH) {
      irOffTime = micros() - usec;
      if (irOffTime > 65000) goto ir_exit;
    }
    irdata[i++] = irOffTime;
  }
  
ir_exit:

  for (j=0; j < i; j++) {
    Serial.print(j % 2 == 0 ? "ON  " : "OFF ");
    Serial.print(irdata[j], DEC);
    Serial.print("\n");
  }
  Serial.print("\n");
}
