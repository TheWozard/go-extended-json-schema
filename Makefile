.PHONY: test clean

test:
	 go test -v ./...

clean:
	git clean -fXd