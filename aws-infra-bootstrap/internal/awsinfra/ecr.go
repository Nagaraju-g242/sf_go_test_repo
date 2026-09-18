package awsinfra

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

func EnsureECRRepository(ctx context.Context, client *ecr.Client, name string) (string, error) {
	out, err := client.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{RepositoryNames: []string{name}})
	if err == nil {
		if len(out.Repositories) == 0 || out.Repositories[0].RepositoryUri == nil {
			return "", fmt.Errorf("repository URI missing")
		}
		return aws.ToString(out.Repositories[0].RepositoryUri), nil
	}
	var nf *ecrtypes.RepositoryNotFoundException
	if !errors.As(err, &nf) {
		return "", fmt.Errorf("describe ECR repository %q: %w", name, err)
	}
	created, err := client.CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName:             aws.String(name),
		ImageScanningConfiguration: &ecrtypes.ImageScanningConfiguration{ScanOnPush: true},
	})
	if err != nil {
		return "", fmt.Errorf("create ECR repository %q: %w", name, err)
	}
	if created.Repository == nil || created.Repository.RepositoryUri == nil {
		return "", fmt.Errorf("repository URI missing after creation")
	}
	return aws.ToString(created.Repository.RepositoryUri), nil
}
