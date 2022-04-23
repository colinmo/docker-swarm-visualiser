build: build-windows build-linux

build-windows:
	set GOOS=windows&&set GOARCH=amd64&&cd src&&go build -ldflags "-w -s" -o ..\bin\docker-swarm-visualiser.exe .
	
build-linux:
	set GOOS=linux&&set GOARCH=amd64&&cd src&&go build -ldflags "-w -s" -o ..\bin\docker-swarm-visualiser

run:
	cd src && go run .


test:
	cd src && go test "./..." -coverprofile="coverage.out" -v 2>&1 | go-junit-report > junit.xml
	cd src && gosonar --basedir E:\xampp\docker-swarm-visualiser\src\cmd\ --coverage coverage.out --junit junit.xml
	cd src && godog run

sonar: test
	docker run --rm -v E:\xampp\docker-swarm-visualiser:/usr/src sonarsource/sonar-scanner-cli

clean:
	del bin\*