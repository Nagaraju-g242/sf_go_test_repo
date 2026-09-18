package awsinfra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func EnsureClusterRole(
	ctx context.Context,
	client *iam.Client,
	roleName string,
) (string, error) {

	return ensureRole(
		ctx,
		client,
		roleName,
		"eks.amazonaws.com",
		[]string{
			"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
		},
	)
}

func EnsureNodeRole(
	ctx context.Context,
	client *iam.Client,
	roleName string,
) (string, error) {

	return ensureRole(
		ctx,
		client,
		roleName,
		"ec2.amazonaws.com",
		[]string{
			"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
			"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPullOnly",
			"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
		},
	)
}

func ensureRole(
	ctx context.Context,
	client *iam.Client,
	roleName string,
	servicePrincipal string,
	policyARNs []string,
) (string, error) {

	fmt.Printf(
		"Checking IAM role %q...\n",
		roleName,
	)

	result, err :=
		client.GetRole(
			ctx,
			&iam.GetRoleInput{
				RoleName: aws.String(roleName),
			},
		)

	if err != nil {

		var notFound *types.NoSuchEntityException

		if !errors.As(err, &notFound) {
			return "", fmt.Errorf(
				"get IAM role %q: %w",
				roleName,
				err,
			)
		}

		fmt.Printf(
			"IAM role %q does not exist. Creating...\n",
			roleName,
		)

		trustPolicy := map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []map[string]interface{}{
				{
					"Effect": "Allow",

					"Principal": map[string]string{
						"Service": servicePrincipal,
					},

					"Action": "sts:AssumeRole",
				},
			},
		}

		policyJSON, err :=
			json.Marshal(trustPolicy)

		if err != nil {
			return "", fmt.Errorf(
				"marshal trust policy: %w",
				err,
			)
		}

		createResult, err :=
			client.CreateRole(
				ctx,
				&iam.CreateRoleInput{
					RoleName: aws.String(roleName),

					AssumeRolePolicyDocument: aws.String(
						string(policyJSON),
					),

					Description: aws.String(
						"Managed by Snowflake Operator infrastructure bootstrap",
					),
				},
			)

		if err != nil {
			return "", fmt.Errorf(
				"create IAM role %q: %w",
				roleName,
				err,
			)
		}

		result =
			&iam.GetRoleOutput{
				Role: createResult.Role,
			}
	}

	if result.Role == nil ||
		result.Role.Arn == nil {

		return "", fmt.Errorf(
			"IAM role %q has no ARN",
			roleName,
		)
	}

	// Attach required AWS managed policies.
	for _, policyARN := range policyARNs {

		fmt.Printf(
			"Ensuring policy %s on %s...\n",
			policyARN,
			roleName,
		)

		_, err :=
			client.AttachRolePolicy(
				ctx,
				&iam.AttachRolePolicyInput{
					RoleName: aws.String(roleName),

					PolicyArn: aws.String(policyARN),
				},
			)

		if err != nil {
			return "", fmt.Errorf(
				"attach policy %q to role %q: %w",
				policyARN,
				roleName,
				err,
			)
		}
	}

	fmt.Printf(
		"IAM role ready: %s\n",
		aws.ToString(result.Role.Arn),
	)

	return aws.ToString(
		result.Role.Arn,
	), nil
}
