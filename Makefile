
.PHONY: default test install clean format

default:
	gomake -C pkg

test:
	gomake -C pkg test

install:
	gomake -C pkg install

clean:
	gomake -C pkg clean

format:
	gomake -C pkg format