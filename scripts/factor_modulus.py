# Recovers p and q primes from N, D, and E components.
# Based on code from https://github.com/ius/rsatool

import math
import random
from base64 import b64decode

def factor_modulus(n, d, e):
    """
    Efficiently recover non-trivial factors of n

    See: Handbook of Applied Cryptography
    8.2.2 Security of RSA -> (i) Relation to factoring (p.287)

    http://www.cacr.math.uwaterloo.ca/hac/
    """
    t = (e * d - 1)
    s = 0

    while True:
        quotient, remainder = divmod(t, 2)

        if remainder != 0:
            break

        s += 1
        t = quotient

    found = False

    while not found:
        i = 1
        a = random.randint(1,n-1)

        while i <= s and not found:
            c1 = pow(a, pow(2, i-1, n) * t, n)
            c2 = pow(a, pow(2, i, n) * t, n)

            found = c1 != 1 and c1 != (-1 % n) and c2 == 1

            i += 1

    p = math.gcd(c1-1, n)
    q = n // p

    return p, q

n_data = b64decode('AihRvNoIbTn85FZRYNZRcT+i6KpU+maCsEqr3Q5q+LDB5tH7Tz2qQ38V')
n = int.from_bytes(n_data[2:], byteorder='big')

d_data = b64decode('AigKVje8mROcR8QixnxUEF5b29Curkq01DNDWCdOG99XBqH79OaCiTCB')
d = int.from_bytes(d_data[2:], byteorder='big')

e = 0x10001

def _hex(x):
    bs = x.to_bytes((x.bit_length() + 7) // 8, byteorder='big')
    return ', '.join(hex(b) for b in bs)

p, q = factor_modulus(n, d, e)
print(f'p={p}')
print(_hex(p))
print(f'q={q}')
print(_hex(q))
