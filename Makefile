.PHONY: all demetra

all: demetra

demetra:
	cd src && CGO_ENABLED=0 go build -o ../demetra
