#!/usr/bin/env python3
import sys, random, json

def main(pin):
    success = random.randrange(10) != 9

    o = {}

    if success:
        o = {'error': False, 'error_code': -9999, 'temperature': 20.8, 'humidity': 41.3}
    else:
        o = {'error': True, 'error_code': 2, 'temperature': -9999, 'humidity': -9999}

    print(json.dumps(o))

if __name__ == '__main__':
        main(sys.argv[1])
