.PHONY: run
## run: runs the app locally
run: 
	cd http-server ; \
		go run . 

dist-http:
	rm -rf dist/http && mkdir -p dist/http/config
	cp  config/app.env dist/http/config
	cp -r static dist/http
	

	cd http-server ; \
	 	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ../dist/http/nextcloud . 

	cd dist/http ; \
		zip -rm ../bundle-http.zip static config nextcloud
	rm -r dist/http