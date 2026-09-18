package awsinfra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// DestroyInfrastructure removes the infrastructure belonging to
// the configured Snowflake Operator stack.
//
// The operation is intentionally idempotent:
// resources that are already missing are treated as successfully deleted.
func DestroyInfrastructure(
	ctx context.Context,
	awsCfg aws.Config,
	cfg Config,
) error {

	eksClient := eks.NewFromConfig(awsCfg)
	ecrClient := ecr.NewFromConfig(awsCfg)
	iamClient := iam.NewFromConfig(awsCfg)
	ec2Client := ec2.NewFromConfig(awsCfg)

	fmt.Println()
	fmt.Println("Starting infrastructure destroy...")
	fmt.Println()

	// ---------------------------------------------------------
	// 1. Delete managed node group
	// ---------------------------------------------------------

	fmt.Println("[1/8] Deleting managed node group...")

	if err := deleteNodeGroup(
		ctx,
		eksClient,
		cfg.ClusterName,
		cfg.NodeGroupName,
	); err != nil {
		return fmt.Errorf(
			"delete node group: %w",
			err,
		)
	}

	// ---------------------------------------------------------
	// 2. Delete EKS cluster
	// ---------------------------------------------------------

	fmt.Println()
	fmt.Println("[2/8] Deleting EKS cluster...")

	if err := deleteEKSCluster(
		ctx,
		eksClient,
		cfg.ClusterName,
	); err != nil {
		return fmt.Errorf(
			"delete EKS cluster: %w",
			err,
		)
	}

	// ---------------------------------------------------------
	// 3. Delete ECR repository
	// ---------------------------------------------------------

	fmt.Println()
	fmt.Println("[3/8] Deleting ECR repository...")

	if err := deleteECRRepository(
		ctx,
		ecrClient,
		cfg.ECRRepository,
	); err != nil {
		return fmt.Errorf(
			"delete ECR repository: %w",
			err,
		)
	}

	// ---------------------------------------------------------
	// 4. Delete node IAM role
	// ---------------------------------------------------------

	fmt.Println()
	fmt.Println("[4/8] Deleting EKS node IAM role...")

	nodePolicies := []string{
		"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
		"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPullOnly",
		"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
	}

	if err := deleteIAMRole(
		ctx,
		iamClient,
		cfg.NodeRoleName,
		nodePolicies,
	); err != nil {
		return fmt.Errorf(
			"delete node IAM role: %w",
			err,
		)
	}

	// ---------------------------------------------------------
	// 5. Delete cluster IAM role
	// ---------------------------------------------------------

	fmt.Println()
	fmt.Println("[5/8] Deleting EKS cluster IAM role...")

	clusterPolicies := []string{
		"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
	}

	if err := deleteIAMRole(
		ctx,
		iamClient,
		cfg.ClusterRoleName,
		clusterPolicies,
	); err != nil {
		return fmt.Errorf(
			"delete cluster IAM role: %w",
			err,
		)
	}

	// ---------------------------------------------------------
	// 6-8. Delete network
	// ---------------------------------------------------------

	fmt.Println()
	fmt.Println("[6/8] Discovering bootstrap VPC...")

	if err := deleteNetwork(
		ctx,
		ec2Client,
		cfg,
	); err != nil {
		return fmt.Errorf(
			"delete network: %w",
			err,
		)
	}

	fmt.Println()
	fmt.Println("[8/8] Destroy validation complete.")

	return nil
}

// ---------------------------------------------------------
// Managed Node Group
// ---------------------------------------------------------

func deleteNodeGroup(
	ctx context.Context,
	client *eks.Client,
	clusterName string,
	nodeGroupName string,
) error {

	_, err := client.DescribeNodegroup(
		ctx,
		&eks.DescribeNodegroupInput{
			ClusterName:   aws.String(clusterName),
			NodegroupName: aws.String(nodeGroupName),
		},
	)

	if err != nil {

		var notFound *ekstypes.ResourceNotFoundException

		if errors.As(err, &notFound) {
			fmt.Println(
				"Node group does not exist. Skipping.",
			)
			return nil
		}

		return err
	}

	fmt.Printf(
		"Deleting node group %s...\n",
		nodeGroupName,
	)

	_, err = client.DeleteNodegroup(
		ctx,
		&eks.DeleteNodegroupInput{
			ClusterName:   aws.String(clusterName),
			NodegroupName: aws.String(nodeGroupName),
		},
	)

	if err != nil {
		return err
	}

	fmt.Println(
		"Waiting for node group deletion...",
	)

	waiter :=
		eks.NewNodegroupDeletedWaiter(client)

	err = waiter.Wait(
		ctx,
		&eks.DescribeNodegroupInput{
			ClusterName:   aws.String(clusterName),
			NodegroupName: aws.String(nodeGroupName),
		},
		30*time.Minute,
	)

	if err != nil {
		return fmt.Errorf(
			"wait for node group deletion: %w",
			err,
		)
	}

	fmt.Println(
		"Node group deleted.",
	)

	return nil
}

// ---------------------------------------------------------
// EKS Cluster
// ---------------------------------------------------------

func deleteEKSCluster(
	ctx context.Context,
	client *eks.Client,
	clusterName string,
) error {

	_, err := client.DescribeCluster(
		ctx,
		&eks.DescribeClusterInput{
			Name: aws.String(clusterName),
		},
	)

	if err != nil {

		var notFound *ekstypes.ResourceNotFoundException

		if errors.As(err, &notFound) {
			fmt.Println(
				"EKS cluster does not exist. Skipping.",
			)
			return nil
		}

		return err
	}

	fmt.Printf(
		"Deleting EKS cluster %s...\n",
		clusterName,
	)

	_, err = client.DeleteCluster(
		ctx,
		&eks.DeleteClusterInput{
			Name: aws.String(clusterName),
		},
	)

	if err != nil {
		return err
	}

	fmt.Println(
		"Waiting for EKS cluster deletion...",
	)

	waiter :=
		eks.NewClusterDeletedWaiter(client)

	err = waiter.Wait(
		ctx,
		&eks.DescribeClusterInput{
			Name: aws.String(clusterName),
		},
		30*time.Minute,
	)

	if err != nil {
		return fmt.Errorf(
			"wait for cluster deletion: %w",
			err,
		)
	}

	fmt.Println(
		"EKS cluster deleted.",
	)

	return nil
}

// ---------------------------------------------------------
// ECR Repository
// ---------------------------------------------------------

func deleteECRRepository(
	ctx context.Context,
	client *ecr.Client,
	repositoryName string,
) error {

	_, err := client.DescribeRepositories(
		ctx,
		&ecr.DescribeRepositoriesInput{
			RepositoryNames: []string{
				repositoryName,
			},
		},
	)

	if err != nil {

		var notFound *ecrtypes.RepositoryNotFoundException

		if errors.As(err, &notFound) {
			fmt.Println(
				"ECR repository does not exist. Skipping.",
			)
			return nil
		}

		return err
	}

	fmt.Printf(
		"Deleting ECR repository %s...\n",
		repositoryName,
	)

	_, err = client.DeleteRepository(
		ctx,
		&ecr.DeleteRepositoryInput{
			RepositoryName: aws.String(
				repositoryName,
			),

			Force: true,
		},
	)

	if err != nil {
		return err
	}

	fmt.Println(
		"ECR repository deleted.",
	)

	return nil
}

// ---------------------------------------------------------
// IAM Role
// ---------------------------------------------------------

func deleteIAMRole(
	ctx context.Context,
	client *iam.Client,
	roleName string,
	expectedPolicies []string,
) error {

	_, err := client.GetRole(
		ctx,
		&iam.GetRoleInput{
			RoleName: aws.String(roleName),
		},
	)

	if err != nil {

		var notFound *iamtypes.NoSuchEntityException

		if errors.As(err, &notFound) {
			fmt.Printf(
				"IAM role %s does not exist. Skipping.\n",
				roleName,
			)
			return nil
		}

		return err
	}

	fmt.Printf(
		"Cleaning IAM role %s...\n",
		roleName,
	)

	// Detach the policies expected from our bootstrap.
	for _, policyARN := range expectedPolicies {

		_, err :=
			client.DetachRolePolicy(
				ctx,
				&iam.DetachRolePolicyInput{
					RoleName: aws.String(
						roleName,
					),

					PolicyArn: aws.String(
						policyARN,
					),
				},
			)

		if err != nil {

			var notFound *iamtypes.NoSuchEntityException

			if !errors.As(
				err,
				&notFound,
			) {
				return fmt.Errorf(
					"detach policy %s from %s: %w",
					policyARN,
					roleName,
					err,
				)
			}
		}
	}

	// Delete any inline policies if present.
	inlinePolicies, err :=
		client.ListRolePolicies(
			ctx,
			&iam.ListRolePoliciesInput{
				RoleName: aws.String(roleName),
			},
		)

	if err != nil {
		return fmt.Errorf(
			"list inline policies for %s: %w",
			roleName,
			err,
		)
	}

	for _, policyName := range inlinePolicies.PolicyNames {

		_, err :=
			client.DeleteRolePolicy(
				ctx,
				&iam.DeleteRolePolicyInput{
					RoleName: aws.String(
						roleName,
					),

					PolicyName: aws.String(
						policyName,
					),
				},
			)

		if err != nil {
			return fmt.Errorf(
				"delete inline policy %s: %w",
				policyName,
				err,
			)
		}
	}

	_, err = client.DeleteRole(
		ctx,
		&iam.DeleteRoleInput{
			RoleName: aws.String(roleName),
		},
	)

	if err != nil {
		return err
	}

	fmt.Printf(
		"IAM role %s deleted.\n",
		roleName,
	)

	return nil
}

// ---------------------------------------------------------
// Network
// ---------------------------------------------------------

func deleteNetwork(
	ctx context.Context,
	client *ec2.Client,
	cfg Config,
) error {

	// ---------------------------------------------------------
	// Find only the VPC created/managed by this stack.
	// ---------------------------------------------------------

	vpcs, err :=
		client.DescribeVpcs(
			ctx,
			&ec2.DescribeVpcsInput{
				Filters: []ec2types.Filter{
					{
						Name: aws.String(
							"tag:Cluster",
						),

						Values: []string{
							cfg.ClusterName,
						},
					},
					{
						Name: aws.String(
							"tag:ManagedBy",
						),

						Values: []string{
							"snowflake-operator-infra-go",
						},
					},
				},
			},
		)

	if err != nil {
		return fmt.Errorf(
			"describe bootstrap VPC: %w",
			err,
		)
	}

	if len(vpcs.Vpcs) == 0 {

		fmt.Println(
			"Bootstrap VPC does not exist. Skipping network deletion.",
		)

		fmt.Println()
		fmt.Println(
			"[7/8] Network already absent.",
		)

		return nil
	}

	if len(vpcs.Vpcs) > 1 {
		return fmt.Errorf(
			"multiple bootstrap VPCs found for cluster %q; refusing automatic deletion",
			cfg.ClusterName,
		)
	}

	vpcID :=
		aws.ToString(
			vpcs.Vpcs[0].VpcId,
		)

	fmt.Printf(
		"Found managed VPC: %s\n",
		vpcID,
	)

	fmt.Println()
	fmt.Println(
		"[7/8] Deleting subnets, route table, Internet Gateway and VPC...",
	)

	// ---------------------------------------------------------
	// Find cluster-specific public route table.
	// ---------------------------------------------------------

	routeTableName :=
		cfg.ClusterName +
			"-public-rt"

	routeTables, err :=
		client.DescribeRouteTables(
			ctx,
			&ec2.DescribeRouteTablesInput{
				Filters: []ec2types.Filter{
					{
						Name: aws.String(
							"vpc-id",
						),

						Values: []string{
							vpcID,
						},
					},
					{
						Name: aws.String(
							"tag:Name",
						),

						Values: []string{
							routeTableName,
						},
					},
				},
			},
		)

	if err != nil {
		return fmt.Errorf(
			"describe route table: %w",
			err,
		)
	}

	// ---------------------------------------------------------
	// Disassociate and delete our public route table.
	// ---------------------------------------------------------

	for _, rt := range routeTables.RouteTables {

		routeTableID :=
			aws.ToString(
				rt.RouteTableId,
			)

		fmt.Printf(
			"Cleaning route table: %s\n",
			routeTableID,
		)

		for _, association := range rt.Associations {

			if aws.ToBool(
				association.Main,
			) {
				continue
			}

			associationID :=
				aws.ToString(
					association.RouteTableAssociationId,
				)

			if associationID == "" {
				continue
			}

			_, err :=
				client.DisassociateRouteTable(
					ctx,
					&ec2.DisassociateRouteTableInput{
						AssociationId: aws.String(
							associationID,
						),
					},
				)

			if err != nil {
				return fmt.Errorf(
					"disassociate route table %s: %w",
					routeTableID,
					err,
				)
			}
		}

		_, err :=
			client.DeleteRouteTable(
				ctx,
				&ec2.DeleteRouteTableInput{
					RouteTableId: aws.String(
						routeTableID,
					),
				},
			)

		if err != nil {
			return fmt.Errorf(
				"delete route table %s: %w",
				routeTableID,
				err,
			)
		}

		fmt.Printf(
			"Route table deleted: %s\n",
			routeTableID,
		)
	}

	// ---------------------------------------------------------
	// Delete only the two cluster-named public subnets.
	// ---------------------------------------------------------

	for i := 1; i <= 2; i++ {

		subnetName :=
			fmt.Sprintf(
				"%s-public-%d",
				cfg.ClusterName,
				i,
			)

		subnets, err :=
			client.DescribeSubnets(
				ctx,
				&ec2.DescribeSubnetsInput{
					Filters: []ec2types.Filter{
						{
							Name: aws.String(
								"vpc-id",
							),

							Values: []string{
								vpcID,
							},
						},
						{
							Name: aws.String(
								"tag:Name",
							),

							Values: []string{
								subnetName,
							},
						},
					},
				},
			)

		if err != nil {
			return fmt.Errorf(
				"describe subnet %s: %w",
				subnetName,
				err,
			)
		}

		for _, subnet := range subnets.Subnets {

			subnetID :=
				aws.ToString(
					subnet.SubnetId,
				)

			fmt.Printf(
				"Deleting subnet: %s (%s)\n",
				subnetID,
				subnetName,
			)

			_, err :=
				client.DeleteSubnet(
					ctx,
					&ec2.DeleteSubnetInput{
						SubnetId: aws.String(
							subnetID,
						),
					},
				)

			if err != nil {
				return fmt.Errorf(
					"delete subnet %s: %w",
					subnetID,
					err,
				)
			}
		}
	}

	// ---------------------------------------------------------
	// Find Internet Gateway attached to this VPC.
	// ---------------------------------------------------------

	igws, err :=
		client.DescribeInternetGateways(
			ctx,
			&ec2.DescribeInternetGatewaysInput{
				Filters: []ec2types.Filter{
					{
						Name: aws.String(
							"attachment.vpc-id",
						),

						Values: []string{
							vpcID,
						},
					},
				},
			},
		)

	if err != nil {
		return fmt.Errorf(
			"describe Internet Gateway: %w",
			err,
		)
	}

	for _, igw := range igws.InternetGateways {

		igwID :=
			aws.ToString(
				igw.InternetGatewayId,
			)

		fmt.Printf(
			"Detaching Internet Gateway: %s\n",
			igwID,
		)

		_, err :=
			client.DetachInternetGateway(
				ctx,
				&ec2.DetachInternetGatewayInput{
					InternetGatewayId: aws.String(
						igwID,
					),

					VpcId: aws.String(
						vpcID,
					),
				},
			)

		if err != nil {
			return fmt.Errorf(
				"detach Internet Gateway %s: %w",
				igwID,
				err,
			)
		}

		_, err =
			client.DeleteInternetGateway(
				ctx,
				&ec2.DeleteInternetGatewayInput{
					InternetGatewayId: aws.String(
						igwID,
					),
				},
			)

		if err != nil {
			return fmt.Errorf(
				"delete Internet Gateway %s: %w",
				igwID,
				err,
			)
		}

		fmt.Printf(
			"Internet Gateway deleted: %s\n",
			igwID,
		)
	}

	// ---------------------------------------------------------
	// Safety check before deleting VPC.
	//
	// At this point our explicit stack resources should be gone.
	// AWS will refuse DeleteVpc if dependencies still remain.
	// ---------------------------------------------------------

	fmt.Printf(
		"Deleting VPC: %s\n",
		vpcID,
	)

	_, err =
		client.DeleteVpc(
			ctx,
			&ec2.DeleteVpcInput{
				VpcId: aws.String(
					vpcID,
				),
			},
		)

	if err != nil {
		return fmt.Errorf(
			"delete VPC %s: %w",
			vpcID,
			err,
		)
	}

	fmt.Printf(
		"VPC deleted: %s\n",
		vpcID,
	)

	return nil
}
