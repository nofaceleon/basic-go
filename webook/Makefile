.PHONY: docker

docker:
	#把上次的东西删掉
	@rm webook-linux || true
	@docker rmi -f webook-linux:v0.0.1
	@GOOS=linux GOARCH=amd64 go build -tags=k8s -o webook-linux main.go
	@docker build -t sjs/webook:v0.0.1 .


