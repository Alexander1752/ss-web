import paho.mqtt.client as mqtt
import os
import json
import base64
from PIL import Image
import pytesseract
import io

MQTT_HOST = os.getenv("MQTT_HOST", "broker")
MQTT_PORT = int(os.getenv("MQTT_PORT", 8883))

CA_CERTS = "/certs/ca.crt"
CERTFILE = "/certs/client.crt"
KEYFILE = "/certs/client.key"

def on_connect(client, userdata, flags, rc):
    print(f"Connected to MQTT Broker with result code {rc}")
    client.subscribe("ssproject/images/#")

def on_message(client, userdata, msg):
    try:
        data = json.loads(msg.payload.decode('utf-8'))
        photo_id = data.get("photo_id")
        image_base64 = data.get("image_base64")
        
        if not photo_id or not image_base64:
            print("Invalid message format: missing photo_id or image_base64")
            return
            
        print(f"Processing OCR for photo {photo_id}")
        
        image_bytes = base64.b64decode(image_base64)
        image = Image.open(io.BytesIO(image_bytes))
        
        # Run OCR with English and Romanian
        text = pytesseract.image_to_string(image, lang='eng+ron')
        
        result_payload = {
            "photo_id": photo_id,
            "text": text.strip() if text else "OCR failed"
        }
        
        res = client.publish("ssproject/ocr/results", json.dumps(result_payload))
        res.wait_for_publish()
        print(f"Published OCR result for photo {photo_id}")
        
    except Exception as e:
        print(f"Error processing message: {e}")

client = mqtt.Client(client_id=f"ocr-service-{os.getenv('HOSTNAME')}")

tls_available = os.path.exists(CA_CERTS) and os.path.exists(CERTFILE) and os.path.exists(KEYFILE)

if tls_available:
    client.tls_set(ca_certs=CA_CERTS, certfile=CERTFILE, keyfile=KEYFILE)
else:
    if MQTT_PORT == 8883:
        print("Error: TLS certificates not found but MQTT_PORT is 8883 (requires TLS). "
              "Provide certificates or set MQTT_PORT=1883 for an unencrypted connection.")
        exit(1)
    print("Warning: TLS certificates not found, connecting without TLS")

client.on_connect = on_connect
client.on_message = on_message

client.connect(MQTT_HOST, MQTT_PORT, 60)
client.loop_forever()