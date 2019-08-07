all: build

hello:
		echo "Hello"

build:
		go build -o libgotst.so -buildmode=c-shared libgotst.go
		clang -o gotst libgotst.c -L. -lgotst

clean:
		rm gotst libgotst.so libgotst.h

run:
		LD_LIBRARY_PATH=./ ./gotst
