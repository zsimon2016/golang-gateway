export GO= O111MODULE=on GOSUMDB=off GOPROXY=https://goproxy.cn,direct go


linux-bin:
	GOOS=linux $(GO) build -tags netgo -o client service.go 

image:linux-bin
	docker build -f ./Dockerfile -t api/client:1.1 .
	
clean:
	rm -f client