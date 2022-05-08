# 1. enter source_code dir
import os

os.chdir("../source_code")

dirs = os.listdir(".")

for root, _, files in os.walk("."):
    for file in files:
        last_two_char = file[len(file) - 2:]
        if last_two_char == ".1" or last_two_char == ".2":
            os.remove(os.path.join(root, file))
