LOCALPATH := /usr/local/bin/
ARTIFACT_PATH := out
SERVICE := setup

build:
	mkdir -p $ARTIFACT_PATH
	go build -o ${ARTIFACT_PATH}/${SERVICE} cmd/main.go

clean:
	rm -rf ${ARTIFACT_PATH}

move:
	mv ${ARTIFACT_PATH}/${SERVICE} ${LOCALPATH}

install: build move clean
