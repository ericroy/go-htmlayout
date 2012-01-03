import os
import zipfile
import urllib2
import shutil
import patch_htmlayout_header

URL = "http://www.terrainformatica.com/htmlayout/HTMLayoutSDK.zip"
PATH = "./HTMLayoutSDK.zip"
TARGET = "./htmlayout"

"""
Borrowed this unzipper from:
http://code.activestate.com/recipes/252508-file-unzip/
"""
class Unzipper:
    def __init__(self, verbose = False, percent = 10):
        self.verbose = verbose
        self.percent = percent
        
    def extract(self, file, dir):
        if not dir.endswith(':') and not os.path.exists(dir):
            os.mkdir(dir)

        zf = zipfile.ZipFile(file)

        # create directory structure to house files
        self._createstructure(file, dir)

        num_files = len(zf.namelist())
        percent = self.percent
        divisions = 100 / percent
        perc = int(num_files / divisions)

        # extract files to directory structure
        for i, name in enumerate(zf.namelist()):
            if self.verbose == True:
                print "Extracting %s" % name
            elif perc > 0 and (i % perc) == 0 and i > 0:
                complete = int (i / perc) * percent
                print "%s%% complete" % complete
            if not name.endswith('/'):
                outfile = open(os.path.join(dir, name), 'wb')
                outfile.write(zf.read(name))
                outfile.flush()
                outfile.close()

    def _createstructure(self, file, dir):
        self._makedirs(self._listdirs(file), dir)

    def _makedirs(self, directories, basedir):
        """ Create any directories that don't currently exist """
        for dir in directories:
            curdir = os.path.join(basedir, dir)
            if not os.path.exists(curdir):
                os.mkdir(curdir)

    def _listdirs(self, file):
        """ Grabs all the directories in the zip structure
        This is necessary to create the structure before trying
        to extract the file to it. """
        zf = zipfile.ZipFile(file)
        dirs = []
        for name in zf.namelist():
            if name.endswith('/'):
                dirs.append(name)
        dirs.sort()
        return dirs



def main():
	if os.path.exists(PATH):
		os.remove(PATH)
	if os.path.exists(TARGET):
		shutil.rmtree(TARGET)
	print("Downloading sdk...")
	with open(PATH, "wb") as dest:
		source = urllib2.urlopen(URL)
		dest.write(source.read())
		source.close()
	print("Download complete, extracting...")
	os.mkdir(TARGET)
	z = Unzipper()
	z.extract(PATH, TARGET)
	print("Extraction complete, patching...")
	patch_htmlayout_header.main()
	print("Finished.")

if __name__ == "__main__":
	main()