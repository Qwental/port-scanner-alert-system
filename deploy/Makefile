.PHONY: build deploy check clean

build:
	cd .. && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
		CC=x86_64-linux-gnu-gcc \
		go build -ldflags="-s -w" -o bin/scanner-linux-amd64 cmd/app/main.go

check: build
	ansible-playbook -i inventory.ini playbook.yaml --check --diff --ask-vault-pass

deploy: build
	ansible-playbook -i inventory.ini playbook.yaml --diff --ask-vault-pass

clean:
	rm -f ../bin/scanner-linux-amd64