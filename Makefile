build: build-windows build-linux

build-windows:
	set GOOS=windows&&set GOARCH=amd64&&cd src&&go build -ldflags "-w -s" -o ..\bin\docker-swarm-visualiser.exe .
	
build-linux:
	set GOOS=linux&&set GOARCH=amd64&&cd src&&go build -ldflags "-w -s" -o ..\bin\docker-swarm-visualiser

run:
	cd src && go run .

test:
	godog

sonar: test

clean:
  del bin\*