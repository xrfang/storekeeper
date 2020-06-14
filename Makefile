GOMOD=storekeeper
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
HASH=$(shell git log -n1 --pretty=format:%h)
REVS=$(shell git log --oneline|wc -l)
build: release
upx:
	upx -9 bin/storekeeper
debug: setver geneh compdbg pack
release: setver geneh comprel upx pack
geneh: #generate error handler
	@for tpl in `find . -type f |grep errors.tpl`; do \
	    target=`echo $$tpl|sed 's/\.tpl/\.go/'`; \
	    pkg=`basename $$(dirname $$tpl)`; \
		sed "s/package main/package $$pkg/" src/errors.go > $$target; \
    done
setver:
	cp src/verinfo.tpl src/version.go
	sed -i 's/{_BRANCH}/$(BRANCH)/' src/version.go
	sed -i 's/{_G_HASH}/$(HASH)/' src/version.go
	sed -i 's/{_G_REVS}/$(REVS)/' src/version.go
comprel:
	mkdir -p bin && cd src && go build -ldflags="-s -w" . && mv $(GOMOD)* ../bin
compdbg:
	mkdir -p bin && cd src && go build -race -gcflags=all=-d=checkptr=0 . && mv $(GOMOD)* ../bin
pack: export GOOS=
pack: export GOARCH=
pack: export GOARM=
pack:
	cd utils && go build . && ./pack && rm pack
clean:
	rm -fr bin src/version.go src/*/errors.go
	git checkout resources/*
