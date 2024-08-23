
void doupdate() {
  Serial.print("Connecting to ");
  Serial.print(server);

  // if you get a connection, report back via serial:
  if (!client.connect(server, port)) {
    Serial.println("Connection to server failed");
    return;
  }
  Serial.println(" OK");

  //  // Make an HTTP request
//  client.print(String("GET ") + resource + " HTTP/1.0\r\n");
//  client.print(String("Host: ") + server + "\r\n");
////  client.print("Connection: close\r\n\r\n");
//
//
  client.print(String("GET ") + resource + "/" + fversion + " HTTP/1.1\r\n");
  client.print(String("Host: ") + server + "\r\n");
  client.print("Connection: close\r\n\r\n");




  long timeout = millis();
  while (client.available() == 0)
  {
    if (millis() - timeout > 60000L)
    {
      Serial.println(">>> Client Timeout !");
      client.stop();
      delay(10000L);
      return;
    }
  }

  Serial.println("Reading header");
  // Read the response from the server
  String response = client.readStringUntil('\n');
  response.trim();
  Serial.println(response);

  

  uint32_t contentLength = knownFileSize;

  File file = SPIFFS.open("/update.bin", FILE_APPEND);

Serial.println("Printitng percentage");
  while (client.available())
  {
    String line = client.readStringUntil('\n');
    line.trim();
    Serial.println(line);    // Uncomment this to show response header
    line.toLowerCase();
    if (line.startsWith("content-length:"))
    {
      contentLength = line.substring(line.lastIndexOf(':') + 1).toInt();
    }
    else if (line.length() == 0)
    {
      break;
    }
  }


  timeout = millis();
  uint32_t readLength = 0;
  CRC32 crc;

  unsigned long timeElapsed = millis();
  printPercent(readLength, contentLength);


  while (readLength < contentLength && client.connected() && millis() - timeout < 10000L)
  {
    int i = 0;
    while (client.available())
    { 
      // read file data to spiffs
      if (!file.print(char(client.read())))
      {
        Serial.println("Appending file");
      }
      //Serial.print((char)c);       // Uncomment this to show data
      //crc.update(c);
      readLength++;

      if (readLength % (contentLength / 13) == 0)
      {
        printPercent(readLength, contentLength);
      }
      timeout = millis();
    }
  }

  file.close();

  printPercent(readLength, contentLength);
  timeElapsed = millis() - timeElapsed;
  Serial.println();

  client.stop();
  Serial.println("stop client");

  modem.gprsDisconnect();
  Serial.println("gprs disconnect");
  Serial.println();

  float duration = float(timeElapsed) / 1000;
  /*
    Serial.print("Tamaño de Archivo: ");
    Serial.println(contentLength);
    Serial.print("Leido:  ");
    Serial.println(readLength);
    Serial.print("Calculado. CRC32:    0x");
    Serial.println(crc.finalize(), HEX);
    Serial.print("Conocido CRC32:    0x");
    Serial.println(knownCRC32, HEX);
    Serial.print("Bajado en:       ");
    Serial.print(duration);
    Serial.println("s");

    Serial.println("Se genera una espera de 3 segundos");
    for (int i = 0; i < 3; i++)
    {
        Serial.print(String(i) + "...");
        delay(1000);
    }
  */
  //readFile(SPIFFS, "/update.bin");

  updateFromFS();

  // Do nothing forevermore
  //  while (true)
  //  {
  //    delay(1000);
  //  }
}



void int_GSM_modem() {

  // Set GSM module baud rate and UART pins
  SerialAT.begin(115200, SERIAL_8N1, RXD2, TXD2);
  gsmreset() ;
  delay(7000);

  // To skip it, call init() instead of restart()
  SerialMon.println("Initializing modem...");
  //  modem.restart();
  modem.init();

  if (!modem.waitForNetwork()) {
    Serial.println(" fail");
    //    while (true);
  }
  SerialMon.print("Connecting to APN: ");
  SerialMon.print(apn);
  if (!modem.gprsConnect(apn, gprsUser, gprsPass)) {
    SerialMon.println(" fail");
    //    ESP.restart();
  }
  else {
    SerialMon.println(" OK");
  }

  if (modem.isGprsConnected()) {
    SerialMon.println("GPRS connected");
  }
}
