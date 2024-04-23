build:
	go build -o lumus ./src/lumus.go

clean:
	rm -f lumus

.PHONY: build clean

