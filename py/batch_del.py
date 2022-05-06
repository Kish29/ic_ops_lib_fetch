# 1. enter source_code dir
import os

os.chdir("../source_code")

dirs = os.listdir(".")

for root, _, files in os.walk("."):
    for file in files:
        if ".1" == file[len(file) - 2:]:
            os.remove(os.path.join(root, file))
