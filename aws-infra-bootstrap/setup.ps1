$ErrorActionPreference = "Stop"
go get github.com/aws/aws-sdk-go-v2@latest
go get github.com/aws/aws-sdk-go-v2/config@latest
go get github.com/aws/aws-sdk-go-v2/service/ecr@latest
go get github.com/aws/aws-sdk-go-v2/service/eks@latest
go get github.com/aws/aws-sdk-go-v2/service/sts@latest
go mod tidy
go fmt ./...
