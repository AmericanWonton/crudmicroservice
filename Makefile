run:
	go run *.go
gomod-exp:
	export GO111MODULE=on
gobuild:
	GOOS=linux GOARCH=amd64 go build -o crudoperations
dockerbuild:
	docker build -t crudoperations .
dockerbuildandpush:
	docker build -t crudoperations .
	docker tag crudoperations americanwonton/crudoperations
	docker push americanwonton/crudoperations
dockerrun:
	docker run -it -p 80:80 crudoperations
dockerrundetached:
	docker run -d -p 80:80 crudoperations
dockerrunitvolume:
	docker run -it -p 80:80 -v photo-images:/static/images crudoperations
dockerrundetvolume:
	docker run -d -p 80:80 -v photo-images:/static/images crudoperations
dockertagimage:
	docker tag crudoperations americanwonton/crudoperations
dockerimagepush:
	docker push americanwonton/crudoperations
dockerallpush:
	docker tag crudoperations americanwonton/crudoperations
	docker push americanwonton/crudoperations
dockerseeshell:
	docker run -it crudoperations sh