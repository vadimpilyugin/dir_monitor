import string
import random
import os
import sys
import time
from os import path
def id_generator(size, chars=string.ascii_lowercase):
  return ''.join(random.choice(chars) for _ in range(size))

def usage():
  print("Usage: create.py [DIRECTORY]")
  sys.exit(1)

if len(sys.argv) < 2:
  usage()
dir_path = sys.argv[1]

for i in range(100):
  fn = "somefile"+str(i)+".txt"
  with open(path.join(dir_path, fn), "w") as f:
    f.write(id_generator(10**5))
    print("Written", fn)
  time.sleep(5)
