import io
import json
import os
import socket
import ssl
import sys
import time

import paho.mqtt.client as mqtt
from PIL import Image, ImageDraw

# Configuration
BROKER = "127.0.0.1"
PORT = 8883  # MQTT with mTLS

# ===== CHANGE THIS FOR EACH DEVICE =====
DEVICE_ID = "python-sender-1"  # Unique ID for this device
DEVICE_NAME = "Python Test Device"  # Human-readable name
# ========================================

# Topics based on device ID
REGISTER_TOPIC = f"register/{DEVICE_ID}"
# Fixed topic to match server subscription (ssproject/images/#)
PHOTO_TOPIC = f"ssproject/images/{DEVICE_ID}"

# Get absolute paths
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
PROJECT_ROOT = os.path.dirname(SCRIPT_DIR)

SECRETS_DIR = os.path.join(PROJECT_ROOT, "secrets")
CA_CRT = os.path.join(SECRETS_DIR, "ca.pem")
CLIENT_CRT = os.path.join(SECRETS_DIR, "client.crt")
CLIENT_KEY = os.path.join(SECRETS_DIR, "client.key")

def get_local_ip():
    """Get the local IP address of this machine"""
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        ip = s.getsockname()[0]
        s.close()
        return ip
    except:
        return "unknown"

def create_test_image():
    """Create a test image with timestamp and device ID"""
    img = Image.new('RGB', (300, 150), color='white')
    d = ImageDraw.Draw(img)
    d.text((10, 20), f"Device: {DEVICE_ID}", fill='black')
    d.text((10, 50), f"Time: {time.strftime('%H:%M:%S')}", fill='black')
    d.text((10, 80), "HELLO ADMIN", fill='blue')

    img_byte_arr = io.BytesIO()
    img.save(img_byte_arr, format='JPEG')
    return img_byte_arr.getvalue()

def load_image_from_file(path):
    """Load image from file and convert to bytes"""
    try:
        with Image.open(path) as img:
            # Convert to RGB if necessary (e.g. for PNGs with transparency)
            if img.mode in ('RGBA', 'P'):
                img = img.convert('RGB')

            img_byte_arr = io.BytesIO()
            img.save(img_byte_arr, format='JPEG')
            return img_byte_arr.getvalue()
    except Exception as e:
        print(f"Error loading image {path}: {e}")
        sys.exit(1)

def on_connect(client, userdata, flags, rc):
    if rc == 0:
        print("Connected to MQTT Broker!")

        # Step 1: Register the device
        local_ip = get_local_ip()
        registration = json.dumps({
            "name": DEVICE_NAME,
            "ip": local_ip,
            "port": str(PORT)
        })
        print(f"Registering device: {REGISTER_TOPIC}")
        client.publish(REGISTER_TOPIC, registration)
        time.sleep(0.5)  # Wait for registration

        # Step 2: Send the photo
        print(f"Publishing image to: {PHOTO_TOPIC}")

        image_data = None
        if len(sys.argv) > 1:
            file_path = sys.argv[1]
            if os.path.isdir(file_path):
                print(f"Error: {file_path} is a directory. Please specify an image file.")
                sys.exit(1)
            print(f"Sending file: {file_path}")
            image_data = load_image_from_file(file_path)
        else:
            print("Sending generated test image (no file argument provided)")
            image_data = create_test_image()

        msg_info = client.publish(PHOTO_TOPIC, image_data)
        msg_info.wait_for_publish()
    else:
        print(f"Failed to connect, return code {rc}")
        sys.exit(1)

def on_publish(client, userdata, mid):
    # Disconnect after second publish (the photo)
    # Note: connect sends no message, register is mid=1, photo is mid=2
    if mid == 2:
        print("Message published successfully!")
        print(f"\n✅ Device '{DEVICE_ID}' registered and photo sent!")
        print(f"   Topic: {PHOTO_TOPIC}")
        client.disconnect()
        sys.exit(0)

# Create MQTT client
client = mqtt.Client(client_id=DEVICE_ID)
client.on_connect = on_connect
client.on_publish = on_publish

client.tls_set(
    ca_certs=CA_CRT,
    certfile=CLIENT_CRT,
    keyfile=CLIENT_KEY,
    tls_version=ssl.PROTOCOL_TLSv1_2
)
client.tls_insecure_set(True)

print(f"Device ID: {DEVICE_ID}")
print(f"Connecting to {BROKER}:{PORT}...")
try:
    client.connect(BROKER, PORT, 60)
    client.loop_forever()
except Exception as e:
    print(f"Connection failed: {e}")
    sys.exit(1)
