import RPi.GPIO as GPIO
import dht11
import json

# initialize GPIO
GPIO.setwarnings(False)
GPIO.setmode(GPIO.BCM)
GPIO.cleanup()

# read data using pin 19
instance = dht11.DHT11(pin = 19)
result = instance.read()

o = {'error': false, 'error_code': -9999, 'temperature': -9999, 'humidity': -9999}

if result.is_valid():
    o['temperature'] = result.temperature
    o['humidity'] = result.humidity
else:
    o['error'] = true
    o['error_code'] = result.error_code

print(json.dumps(o))
