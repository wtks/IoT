#include <PubSubClient.h>
#include <M5Stack.h>
#include <Wire.h>
#include <WiFi.h>
#include <WiFiClient.h>
#include "BME280.h"

const char* ssid = "";
const char* password = "";

const char* mqttEndpoint = "";
const int mqttPort = 1883;
const char* deviceID = "M5Stack";
const char* pubTopicTemp = "/home/temperature";
const char* pubTopicHumidity = "/home/humidity";
const char* pubTopicPressure = "/home/pressure";
const char* pubTopic = "/home/metrics";

long messageSentAt = 0;
char pubMessage[128];

BME280 bme280;
WiFiClient wifi;
PubSubClient mqttClient(wifi);

void setup(){
    M5.begin();
    Wire.begin();
    bme280.begin();
    
    WiFi.mode(WIFI_STA);
    WiFi.disconnect();
    //Initialize serial and wait for port to open:
    Serial.begin(115200);
    delay(100);
    
    Serial.print("Attempting to connect to SSID: ");
    Serial.println(ssid);
    // attempt to connect to Wifi network:
    WiFi.begin(ssid, password);
    
    while (WiFi.status() != WL_CONNECTED) {
      Serial.print(".");
      // wait 1 second for re-trying
      delay(1000);
    }
    
    Serial.print("Connected to ");
    Serial.println(ssid);

    mqttClient.setServer(mqttEndpoint, mqttPort);
}

void connectMQTT() {
    while (!mqttClient.connected()) {
        if (mqttClient.connect(deviceID)) {
            Serial.println("Connected.");
        } else {
            Serial.print("Failed. Error state=");
            Serial.print(mqttClient.state());
            // Wait 5 seconds before retrying
            delay(5000);
        }
    }
}

void loop() {
    if (!mqttClient.connected()) {
        connectMQTT();
    }
    mqttClient.loop();

    long now = millis();
    if (now - messageSentAt > 5000) {
        float temp, humid, pressure;
        temp = (float)bme280.readTemperature();
        humid = (float)bme280.readHumidity();
        pressure = (float)bme280.readPressure();

        Serial.printf("temp: %2.2f humid: %0.2f pressure: %f\n", temp, humid, pressure);
        
        messageSentAt = now;
        sprintf(pubMessage, "%2.2f", temp);
        mqttClient.publish(pubTopicTemp, pubMessage);
        sprintf(pubMessage, "%0.2f", humid);
        mqttClient.publish(pubTopicHumidity, pubMessage);
        sprintf(pubMessage, "%f", pressure);
        mqttClient.publish(pubTopicPressure, pubMessage);
        sprintf(pubMessage, "{\"t\":%2.2f,\"h\":%0.2f,\"p\":%f}", temp, humid, pressure);
        mqttClient.publish(pubTopic, pubMessage);
    }
}
