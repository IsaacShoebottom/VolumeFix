.PHONY: icon syso build clean

# Export CGO_ENABLED=1
CGO_ENABLED=1
# Export C compiler
CC=gcc

icon:
	magick convert ico/VolumeFix.png ico/VolumeFix.ico
	2goarray Icon ico < ico/VolumeFix.ico > ico/VolumeFix.go

syso:
	rsrc -ico ico/VolumeFix.ico -o bin/VolumeFix.syso

build: icon syso
	go build -ldflags "-H windowsgui" -o bin/

build-debug: icon syso
	go build -o bin/VolumeFix

clean:
	rm -f "ico/VolumeFix.ico"
	rm -f "ico/VolumeFix.go"
	rm -f "bin/VolumeFix.syso"
	rm -f "bin/VolumeFix.exe"
	rm -f "bin/VolumeFix"