import os

os.chdir("../source_code")

dirs = os.listdir(".")

count = 0

for root, _, files in os.walk("."):
    if len(files) == 0:
        print(root)
        count = count + 1

print(count)