.PHONY: run
## run: runs the app locally
run: 
	cd http-server ; \
		go run . 

.PHONY: dist-http
dist-http:
	rm -rf dist/http && mkdir -p dist/http
	cp function/install/manifest.json dist/http
	cp -r static dist/http
	

	cd http-server ; \
	 	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ../dist/http/nextcloud . 

	cd dist/http ; \
		zip -rm ../bundle-http.zip static config nextcloud manifest.json
	rm -r dist/http

.PHONY: dist-aws
## dist-aws: creates the bundle file for AWS Lambda deployments
dist-aws:
	rm -rf dist/aws && mkdir -p dist/aws
	cp function/install/manifest.json dist/aws
	cp -r static dist/aws

	cd aws ; \
    	 	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../dist/aws/nextcloud .
	cd dist/aws ; \
				zip -m nextcloud.zip nextcloud ; \
		zip -rm ../bundle-aws.zip  static config nextcloud.zip manifest.json
	rm -r dist/aws