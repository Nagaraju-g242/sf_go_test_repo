package awsinfra

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Network struct {
	VpcID     string
	SubnetIDs []string
}

func EnsureNetwork(
	ctx context.Context,
	client *ec2.Client,
	cfg Config,
) (Network, error) {

	fmt.Println(
		"Checking VPC...",
	)

	// ---------------------------------------------------------
	// Find VPC belonging to THIS cluster
	// ---------------------------------------------------------

	vpcs, err :=
		client.DescribeVpcs(
			ctx,
			&ec2.DescribeVpcsInput{
				Filters: []types.Filter{
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
		return Network{}, fmt.Errorf(
			"describe VPCs: %w",
			err,
		)
	}

	var vpcID string

	// ---------------------------------------------------------
	// Reuse existing VPC
	// ---------------------------------------------------------

	if len(vpcs.Vpcs) > 0 {

		vpcID =
			aws.ToString(
				vpcs.Vpcs[0].VpcId,
			)

		fmt.Printf(
			"Reusing VPC: %s\n",
			vpcID,
		)

	} else {

		// -----------------------------------------------------
		// Create VPC
		// -----------------------------------------------------

		fmt.Printf(
			"Creating VPC %s...\n",
			cfg.VpcCIDR,
		)

		result, err :=
			client.CreateVpc(
				ctx,
				&ec2.CreateVpcInput{
					CidrBlock: aws.String(
						cfg.VpcCIDR,
					),
				},
			)

		if err != nil {
			return Network{}, fmt.Errorf(
				"create VPC: %w",
				err,
			)
		}

		vpcID =
			aws.ToString(
				result.Vpc.VpcId,
			)

		// Tags

		_, err =
			client.CreateTags(
				ctx,
				&ec2.CreateTagsInput{
					Resources: []string{
						vpcID,
					},

					Tags: []types.Tag{
						{
							Key: aws.String(
								"Name",
							),

							Value: aws.String(
								cfg.ClusterName +
									"-vpc",
							),
						},
						{
							Key: aws.String(
								"Cluster",
							),

							Value: aws.String(
								cfg.ClusterName,
							),
						},
						{
							Key: aws.String(
								"ManagedBy",
							),

							Value: aws.String(
								"snowflake-operator-infra-go",
							),
						},
					},
				},
			)

		if err != nil {
			return Network{}, err
		}

		// Enable DNS support.

		_, err =
			client.ModifyVpcAttribute(
				ctx,
				&ec2.ModifyVpcAttributeInput{
					VpcId: aws.String(vpcID),

					EnableDnsSupport: &types.AttributeBooleanValue{
						Value: aws.Bool(true),
					},
				},
			)

		if err != nil {
			return Network{}, err
		}

		// Enable DNS hostnames.

		_, err =
			client.ModifyVpcAttribute(
				ctx,
				&ec2.ModifyVpcAttributeInput{
					VpcId: aws.String(vpcID),

					EnableDnsHostnames: &types.AttributeBooleanValue{
						Value: aws.Bool(true),
					},
				},
			)

		if err != nil {
			return Network{}, err
		}

		fmt.Printf(
			"Created VPC: %s\n",
			vpcID,
		)
	}

	// ---------------------------------------------------------
	// Find available AZs
	// ---------------------------------------------------------

	azResult, err :=
		client.DescribeAvailabilityZones(
			ctx,
			&ec2.DescribeAvailabilityZonesInput{
				Filters: []types.Filter{
					{
						Name: aws.String(
							"state",
						),

						Values: []string{
							"available",
						},
					},
				},
			},
		)

	if err != nil {
		return Network{}, fmt.Errorf(
			"describe availability zones: %w",
			err,
		)
	}

	var azs []string

	for _, az := range azResult.AvailabilityZones {

		if az.ZoneName != nil {
			azs =
				append(
					azs,
					*az.ZoneName,
				)
		}
	}

	sort.Strings(azs)

	if len(azs) < 2 {
		return Network{}, fmt.Errorf(
			"at least two availability zones are required",
		)
	}

	// ---------------------------------------------------------
	// Internet Gateway
	// ---------------------------------------------------------

	igws, err :=
		client.DescribeInternetGateways(
			ctx,
			&ec2.DescribeInternetGatewaysInput{
				Filters: []types.Filter{
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
		return Network{}, err
	}

	var internetGatewayID string

	if len(igws.InternetGateways) > 0 {

		internetGatewayID =
			aws.ToString(
				igws.InternetGateways[0].
					InternetGatewayId,
			)

	} else {

		igw, err :=
			client.CreateInternetGateway(
				ctx,
				&ec2.CreateInternetGatewayInput{},
			)

		if err != nil {
			return Network{}, err
		}

		internetGatewayID =
			aws.ToString(
				igw.InternetGateway.
					InternetGatewayId,
			)

		_, err =
			client.AttachInternetGateway(
				ctx,
				&ec2.AttachInternetGatewayInput{
					VpcId: aws.String(vpcID),

					InternetGatewayId: aws.String(
						internetGatewayID,
					),
				},
			)

		if err != nil {
			return Network{}, err
		}
	}

	// ---------------------------------------------------------
	// Create/reuse two public subnets
	// ---------------------------------------------------------

	cidrs :=
		[]string{
			cfg.PublicSubnet1CIDR,
			cfg.PublicSubnet2CIDR,
		}

	var subnetIDs []string

	for i := 0; i < 2; i++ {

		subnetName :=
			fmt.Sprintf(
				"%s-public-%d",
				cfg.ClusterName,
				i+1,
			)

		existing, err :=
			client.DescribeSubnets(
				ctx,
				&ec2.DescribeSubnetsInput{
					Filters: []types.Filter{
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
			return Network{}, err
		}

		var subnetID string

		if len(existing.Subnets) > 0 {

			subnetID =
				aws.ToString(
					existing.Subnets[0].
						SubnetId,
				)

			fmt.Printf(
				"Reusing subnet: %s\n",
				subnetID,
			)

		} else {

			subnet, err :=
				client.CreateSubnet(
					ctx,
					&ec2.CreateSubnetInput{
						VpcId: aws.String(
							vpcID,
						),

						CidrBlock: aws.String(
							cidrs[i],
						),

						AvailabilityZone: aws.String(
							azs[i],
						),
					},
				)

			if err != nil {
				return Network{}, fmt.Errorf(
					"create subnet %d: %w",
					i+1,
					err,
				)
			}

			subnetID =
				aws.ToString(
					subnet.Subnet.
						SubnetId,
				)

			_, err =
				client.CreateTags(
					ctx,
					&ec2.CreateTagsInput{
						Resources: []string{
							subnetID,
						},

						Tags: []types.Tag{
							{
								Key: aws.String(
									"Name",
								),

								Value: aws.String(
									subnetName,
								),
							},
							{
								Key: aws.String(
									"kubernetes.io/role/elb",
								),

								Value: aws.String(
									"1",
								),
							},
						},
					},
				)

			if err != nil {
				return Network{}, err
			}

			_, err =
				client.ModifySubnetAttribute(
					ctx,
					&ec2.ModifySubnetAttributeInput{
						SubnetId: aws.String(
							subnetID,
						),

						MapPublicIpOnLaunch: &types.AttributeBooleanValue{
							Value: aws.Bool(true),
						},
					},
				)

			if err != nil {
				return Network{}, err
			}

			fmt.Printf(
				"Created subnet: %s (%s)\n",
				subnetID,
				azs[i],
			)
		}

		subnetIDs =
			append(
				subnetIDs,
				subnetID,
			)
	}

	// ---------------------------------------------------------
	// Public route table
	// ---------------------------------------------------------

	routeTableName :=
		cfg.ClusterName +
			"-public-rt"

	routeTables, err :=
		client.DescribeRouteTables(
			ctx,
			&ec2.DescribeRouteTablesInput{
				Filters: []types.Filter{
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
		return Network{}, err
	}

	var routeTableID string

	if len(routeTables.RouteTables) > 0 {

		routeTableID =
			aws.ToString(
				routeTables.RouteTables[0].
					RouteTableId,
			)

	} else {

		rt, err :=
			client.CreateRouteTable(
				ctx,
				&ec2.CreateRouteTableInput{
					VpcId: aws.String(
						vpcID,
					),
				},
			)

		if err != nil {
			return Network{}, err
		}

		routeTableID =
			aws.ToString(
				rt.RouteTable.
					RouteTableId,
			)

		_, err =
			client.CreateTags(
				ctx,
				&ec2.CreateTagsInput{
					Resources: []string{
						routeTableID,
					},

					Tags: []types.Tag{
						{
							Key: aws.String(
								"Name",
							),

							Value: aws.String(
								routeTableName,
							),
						},
					},
				},
			)

		if err != nil {
			return Network{}, err
		}

		_, err =
			client.CreateRoute(
				ctx,
				&ec2.CreateRouteInput{
					RouteTableId: aws.String(
						routeTableID,
					),

					DestinationCidrBlock: aws.String(
						"0.0.0.0/0",
					),

					GatewayId: aws.String(
						internetGatewayID,
					),
				},
			)

		if err != nil {
			return Network{}, err
		}
	}

	// ---------------------------------------------------------
	// Associate subnets with public route table
	// ---------------------------------------------------------

	for _, subnetID := range subnetIDs {

		// Check whether this subnet is already associated
		// with our route table before creating another
		// association.

		associated, err :=
			client.DescribeRouteTables(
				ctx,
				&ec2.DescribeRouteTablesInput{
					Filters: []types.Filter{
						{
							Name: aws.String(
								"association.subnet-id",
							),

							Values: []string{
								subnetID,
							},
						},
					},
				},
			)

		if err != nil {
			return Network{}, err
		}

		alreadyAssociated := false

		for _, rt := range associated.RouteTables {

			if aws.ToString(
				rt.RouteTableId,
			) == routeTableID {

				alreadyAssociated = true
				break
			}
		}

		if !alreadyAssociated {

			_, err =
				client.AssociateRouteTable(
					ctx,
					&ec2.AssociateRouteTableInput{
						RouteTableId: aws.String(
							routeTableID,
						),

						SubnetId: aws.String(
							subnetID,
						),
					},
				)

			if err != nil {
				return Network{}, fmt.Errorf(
					"associate subnet %s with route table: %w",
					subnetID,
					err,
				)
			}
		}
	}

	return Network{
		VpcID: vpcID,

		SubnetIDs: subnetIDs,
	}, nil
}
