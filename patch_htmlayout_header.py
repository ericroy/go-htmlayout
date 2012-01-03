import sys
import os
import operator
import shutil
import re

BASE = "./htmlayout/include"
REPLACEMENTS = {
	"htmlayout_dom.h": [
		("struct htmlayout_dom_element;", "typedef struct htmlayout_dom_element {} htmlayout_dom_element;", 1)
	],
	"htmlayout_behavior.h": [
		("struct EXCHANGE_PARAMS;", "typedef struct EXCHANGE_PARAMS EXCHANGE_PARAMS;", 1)
	],
}

def apply_replacements(path):
	with open(os.path.join(BASE, path), "r+") as file:
		s = file.read()
		for find, replace, count in REPLACEMENTS[path]:
			s = re.sub(find, replace, s, count)
		file.seek(0)
		file.write(s)

def restore_backups():
	for filename in REPLACEMENTS:
		path = os.path.join(BASE, filename)
		original = path+".original"
		if os.path.exists(original):
			shutil.move(original, path)

def main():
	if not os.path.exists(BASE):
		print("You must install the htmlayout sdk alongside this script in a folder called 'htmlayout'")
		return
	paths = [os.path.join(BASE, filename) for filename in REPLACEMENTS]
	existance = [os.path.exists(path) for path in paths]
	if False in existance:
		print("Could not find the following files to patch: "+missing)
		return
	for path in paths:
		dst = path+".original"
		if os.path.exists(dst):
			print("Patch has already been applied (backup '%s' exists)" % dst)
			return
		shutil.copyfile(path, dst)	
	for filename in REPLACEMENTS:
		apply_replacements(filename)
	print("Patch complete")

if __name__ == "__main__":
	if len(sys.argv) == 2 and sys.argv[1] == "restore":
		restore_backups()
	else:
		main()