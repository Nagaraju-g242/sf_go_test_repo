package awsinfra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
)

func EnsureEKSCluster(
	ctx context.Context,
	client *eks.Client,
	cfg Config,
	clusterRoleARN string,
	subnetIDs []string,
) (*types.Cluster, error) {

	// ---------------------------------------------------------
	// Check whether cluster already exists
	// ---------------------------------------------------------

	out, err := client.DescribeCluster(
		ctx,
		&eks.DescribeClusterInput{
			Name: aws.String(cfg.ClusterName),
		},
	)

	if err == nil {

		fmt.Printf(
			"EKS cluster %q already exists. Status: %s\n",
			cfg.ClusterName,
			out.Cluster.Status,
		)

		// Wait if the cluster exists but is not ACTIVE yet.
		if out.Cluster.Status != types.ClusterStatusActive {

			fmt.Printf(
				"Waiting for EKS cluster %q to become ACTIVE...\n",
				cfg.ClusterName,
			)

			waiter :=
				eks.NewClusterActiveWaiter(client)

			err = waiter.Wait(
				ctx,
				&eks.DescribeClusterInput{
					Name: aws.String(cfg.ClusterName),
				},
				30*time.Minute,
			)

			if err != nil {
				return nil, fmt.Errorf(
					"wait for EKS cluster %q: %w",
					cfg.ClusterName,
					err,
				)
			}

			out, err =
				client.DescribeCluster(
					ctx,
					&eks.DescribeClusterInput{
						Name: aws.String(cfg.ClusterName),
					},
				)

			if err != nil {
				return nil, err
			}
		}

		return out.Cluster, nil
	}

	// ---------------------------------------------------------
	// Determine whether the error means "not found"
	// ---------------------------------------------------------

	var notFound *types.ResourceNotFoundException

	if !errors.As(err, &notFound) {
		return nil, fmt.Errorf(
			"describe EKS cluster %q: %w",
			cfg.ClusterName,
			err,
		)
	}

	// ---------------------------------------------------------
	// Create cluster
	// ---------------------------------------------------------

	fmt.Printf(
		"EKS cluster %q does not exist. Creating...\n",
		cfg.ClusterName,
	)

	_, err =
		client.CreateCluster(
			ctx,
			&eks.CreateClusterInput{
				Name: aws.String(
					cfg.ClusterName,
				),

				Version: aws.String(
					cfg.KubernetesVersion,
				),

				RoleArn: aws.String(
					clusterRoleARN,
				),

				ResourcesVpcConfig: &types.VpcConfigRequest{
					SubnetIds: subnetIDs,

					EndpointPublicAccess: aws.Bool(true),

					EndpointPrivateAccess: aws.Bool(true),
				},

				Tags: map[string]string{
					"ManagedBy": "snowflake-operator-infra-go",
					"Project":   "snowflake-operator",
				},
			},
		)

	if err != nil {
		return nil, fmt.Errorf(
			"create EKS cluster %q: %w",
			cfg.ClusterName,
			err,
		)
	}

	// ---------------------------------------------------------
	// Wait for ACTIVE
	// ---------------------------------------------------------

	fmt.Printf(
		"Waiting for EKS cluster %q to become ACTIVE...\n",
		cfg.ClusterName,
	)

	waiter :=
		eks.NewClusterActiveWaiter(client)

	err = waiter.Wait(
		ctx,
		&eks.DescribeClusterInput{
			Name: aws.String(
				cfg.ClusterName,
			),
		},
		30*time.Minute,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"wait for EKS cluster %q: %w",
			cfg.ClusterName,
			err,
		)
	}

	// ---------------------------------------------------------
	// Retrieve final cluster
	// ---------------------------------------------------------

	out, err =
		client.DescribeCluster(
			ctx,
			&eks.DescribeClusterInput{
				Name: aws.String(
					cfg.ClusterName,
				),
			},
		)

	if err != nil {
		return nil, fmt.Errorf(
			"describe newly-created EKS cluster %q: %w",
			cfg.ClusterName,
			err,
		)
	}

	return out.Cluster, nil
}
