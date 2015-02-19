
from __future__ import print_function
import pafy, sys

if len(sys.argv) == 2:
	youtube = pafy.new(sys.argv[1])
	print(youtube.title)
	print(youtube.getbestaudio().url)
else:
	print("usage: %s <url>" % sys.argv[0], file=sys.stderr)
	sys.exit(1)

