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

func EnsureNodeGroup(
	ctx context.Context,
	client *eks.Client,
	cfg Config,
	nodeRoleARN string,
	subnetIDs []string,
) (*types.Nodegroup, error) {

	// ---------------------------------------------------------
	// Check whether node group already exists
	// ---------------------------------------------------------

	out, err :=
		client.DescribeNodegroup(
			ctx,
			&eks.DescribeNodegroupInput{
				ClusterName: aws.String(
					cfg.ClusterName,
				),

				NodegroupName: aws.String(
					cfg.NodeGroupName,
				),
			},
		)

	if err == nil {

		fmt.Printf(
			"Node group %q already exists. Status: %s\n",
			cfg.NodeGroupName,
			out.Nodegroup.Status,
		)

		if out.Nodegroup.Status !=
			types.NodegroupStatusActive {

			fmt.Printf(
				"Waiting for node group %q to become ACTIVE...\n",
				cfg.NodeGroupName,
			)

			waiter :=
				eks.NewNodegroupActiveWaiter(
					client,
				)

			err = waiter.Wait(
				ctx,
				&eks.DescribeNodegroupInput{
					ClusterName: aws.String(
						cfg.ClusterName,
					),

					NodegroupName: aws.String(
						cfg.NodeGroupName,
					),
				},
				30*time.Minute,
			)

			if err != nil {
				return nil, fmt.Errorf(
					"wait for node group %q: %w",
					cfg.NodeGroupName,
					err,
				)
			}

			out, err =
				client.DescribeNodegroup(
					ctx,
					&eks.DescribeNodegroupInput{
						ClusterName: aws.String(
							cfg.ClusterName,
						),

						NodegroupName: aws.String(
							cfg.NodeGroupName,
						),
					},
				)

			if err != nil {
				return nil, err
			}
		}

		return out.Nodegroup, nil
	}

	// ---------------------------------------------------------
	// Check whether error means "not found"
	// ---------------------------------------------------------

	var notFound *types.ResourceNotFoundException

	if !errors.As(err, &notFound) {
		return nil, fmt.Errorf(
			"describe node group %q: %w",
			cfg.NodeGroupName,
			err,
		)
	}

	// ---------------------------------------------------------
	// Create managed node group
	// ---------------------------------------------------------

	fmt.Printf(
		"Node group %q does not exist. Creating...\n",
		cfg.NodeGroupName,
	)

	_, err =
		client.CreateNodegroup(
			ctx,
			&eks.CreateNodegroupInput{
				ClusterName: aws.String(
					cfg.ClusterName,
				),

				NodegroupName: aws.String(
					cfg.NodeGroupName,
				),

				NodeRole: aws.String(
					nodeRoleARN,
				),

				Subnets: subnetIDs,

				InstanceTypes: cfg.InstanceTypes,

				ScalingConfig: &types.NodegroupScalingConfig{
					DesiredSize: aws.Int32(
						cfg.DesiredNodes,
					),

					MinSize: aws.Int32(
						cfg.MinNodes,
					),

					MaxSize: aws.Int32(
						cfg.MaxNodes,
					),
				},

				Tags: map[string]string{
					"ManagedBy": "snowflake-operator-infra-go",

					"Project": "snowflake-operator",
				},
			},
		)

	if err != nil {
		return nil, fmt.Errorf(
			"create node group %q: %w",
			cfg.NodeGroupName,
			err,
		)
	}

	// ---------------------------------------------------------
	// Wait for ACTIVE
	// ---------------------------------------------------------

	fmt.Printf(
		"Waiting for node group %q to become ACTIVE...\n",
		cfg.NodeGroupName,
	)

	waiter :=
		eks.NewNodegroupActiveWaiter(
			client,
		)

	err = waiter.Wait(
		ctx,
		&eks.DescribeNodegroupInput{
			ClusterName: aws.String(
				cfg.ClusterName,
			),

			NodegroupName: aws.String(
				cfg.NodeGroupName,
			),
		},
		30*time.Minute,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"wait for node group %q: %w",
			cfg.NodeGroupName,
			err,
		)
	}

	// ---------------------------------------------------------
	// Retrieve final node group
	// ---------------------------------------------------------

	out, err =
		client.DescribeNodegroup(
			ctx,
			&eks.DescribeNodegroupInput{
				ClusterName: aws.String(
					cfg.ClusterName,
				),

				NodegroupName: aws.String(
					cfg.NodeGroupName,
				),
			},
		)

	if err != nil {
		return nil, fmt.Errorf(
			"describe newly-created node group %q: %w",
			cfg.NodeGroupName,
			err,
		)
	}

	return out.Nodegroup, nil
}
