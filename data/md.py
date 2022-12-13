#!/usr/bin/env python3

# Generates GFM markdown table

import json

def endUserEmoji(val):
    if val:
        return "✅"
    return ""

f = open('networks.json')
data = json.load(f)
print("| Name | Instances | AS Number | End user |")
print("| --- | --- | --- | --- |")
previous_had_children = False
for count, i in enumerate(data):
    if i.get("children") is None:
        print("| {} | {} | {} | {} |".format(i["name"], i["count"], i["asn"], endUserEmoji(i.get("endUser", False))))
        previous_had_children = False
    else:
        if count != 0 and not previous_had_children:
            print("| &nbsp; | | | |")
        print("| **{}** | **{}** | | |".format(i["name"], i["count"]))
        for j in i["children"].values():
            print("| {} | {} | {} | {} |".format(j["name"], j["count"], j["asn"], endUserEmoji(j.get("endUser", False))))
        print("| &nbsp; | | | |")
        previous_had_children = True
f.close()
