#include <PubSubClient.h>
#include <M5Stack.h>
#include <WiFi.h>
#include <WiFiClient.h>

const char* ssid = "";
const char* password = "";

const char* mqttEndpoint = "";
const int mqttPort = 1883;
const char* deviceID = "M5Stack";
const char* subTopic = "/home/temperature";

WiFiClient wifi;
PubSubClient mqttClient(wifi);

void callback(char* topic, byte* payload, unsigned int length) {
  Serial.print("Message arrived [");
  Serial.print(topic);
  Serial.print("] ");

  //M5.Lcd.fillScreen(0);
  M5.Lcd.setCursor(1, 1);
  for (int i=0;i<length;i++) {
    Serial.print((char)payload[i]);
    M5.Lcd.printf("%c", (char)payload[i]);
  }
  
  Serial.println();
}

void reconnect() {
  // Loop until we're reconnected
  while (!mqttClient.connected()) {
    Serial.print("Attempting MQTT connection...");
    // Attempt to connect
    if (mqttClient.connect(deviceID)) {
      Serial.println("connected");
      mqttClient.subscribe(subTopic);
    } else {
      Serial.print("failed, rc=");
      Serial.print(mqttClient.state());
      Serial.println(" try again in 5 seconds");
      // Wait 5 seconds before retrying
      delay(5000);
    }
  }
}

void setup() {
  M5.begin();

  M5.Lcd.setTextSize(50);

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
  mqttClient.setCallback(callback);
}

void loop() {
  if (!mqttClient.connected()) {
    reconnect();
  }
  mqttClient.loop();
}
