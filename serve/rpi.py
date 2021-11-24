#!/usr/bin/env python3
import sys, json
import RPi.GPIO as GPIO
import dht11

def main(pin):
    # initialize GPIO
    GPIO.setwarnings(False)
    GPIO.setmode(GPIO.BCM)
    GPIO.cleanup()
    
    # read data using provided pin
    instance = dht11.DHT11(pin = pin)
    result = instance.read()
    
    o = {'error': False, 'error_code': -9999, 'temperature': -9999, 'humidity': -9999}

    if result.is_valid():
        o['temperature'] = result.temperature
        o['humidity'] = result.humidity
    else:
        o['error'] = True
        o['error_code'] = result.error_code

    print(json.dumps(o))

if __name__ == '__main__':
        main(int(sys.argv[1]))
