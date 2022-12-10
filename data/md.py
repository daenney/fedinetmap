#!/usr/bin/env python3

# Generates GFM markdown table

import json

f = open('networks.json')
data = json.load(f)
for i in data:
    print("| {} | {} | {} |".format(i["name"], i["count"], ", ".join([str(j) for j in i["asNumbers"]])))
f.close()
