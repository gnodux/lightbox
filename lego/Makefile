
OS=`go env GOOS`
ARCH=`go env GOARCH`
INSTALL_PATH=/usr/local/share/lego
TARGET_OS={darwin,linux,windows}
TARGET_ARCH={arm64,amd64}
VERSION ?=1.2

generate:
	cd ../ext && go generate
build: generate
	go build -ldflags "-s -w"  -o lego.${OS}
install: build
	cp lego.${OS} ${INSTALL_PATH}/lego
	[ -f /usr/local/bin/lego ] && rm /usr/local/bin/lego
	ln -s ${INSTALL_PATH}/lego /usr/local/bin/
	cp ../ext/tengo.json  ~/Library/Application\ Support/Code/User/snippets/tengo.json
	cp -r ../docs ${INSTALL_PATH}/
all: generate clean
	cp ../ext/tengo.json ./output/tengo.json
	cp -r ../docs ./output/docs
	for os in ${TARGET_OS}; do \
		for arch in ${TARGET_ARCH}; do \
			GOOS=$$os GOARCH=$$arch go build  -o ./output/lego.$$os-$$arch; \
		done; \
	done
release:all
	if [ -f lego ]; then rm lego; fi && \
	if [ -f "lego.exe" ]; then rm lego.exe; fi
	cd output && for os in ${TARGET_OS}; do \
		for arch in ${TARGET_ARCH}; do \
		  [ $$os = "windows" ]&& ext=".exe"; \
		  cp lego.$$os-$$arch lego$${ext};\
		  zip -r lego.$$os-$$arch-${VERSION}.zip lego$${ext} tengo.json docs; \
		  mv lego.$$os-$$arch-${VERSION}.zip ../../lego-release;\
		done; \
	done
clean:
	if [ -d ./output ]; then  rm -r ./output; fi
	mkdir output