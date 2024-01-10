import sys
import random

"""
This script was used to generate the datasets.
I chose values between 1,1024 so that the inervals would fit nicely for the largest amount of leaves 2^10
number of values given as a command line parameter.
"""
size = int(sys.argv[1])

file_name = f"data{size}.txt"

with open(file_name, 'w') as file:
    for _ in range(size):
        value = random.uniform(1, 1024)
        file.write(f"{value}\n")
