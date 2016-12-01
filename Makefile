dist:
	mkdir dist

dist/resizer.linux.x86: dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/resizer.linux.x86 github.com/marshallbrekka/image-resizer-lambda/src/go

dist/lambda.zip: dist dist/resizer.linux.x86
	zip -j dist/lambda.zip dist/resizer.linux.x86 src/js/index.js

build: dist/lambda.zip

clean:
	rm -r dist
