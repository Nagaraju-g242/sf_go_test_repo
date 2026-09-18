package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"example.com/snowflake-operator-infra/internal/awsinfra"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func main() {

	ctx := context.Background()

	// ---------------------------------------------------------
	// 1. Load configuration
	// ---------------------------------------------------------

	cfg, err := awsinfra.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf(
			"configuration error: %v",
			err,
		)
	}

	// ---------------------------------------------------------
	// 2. Load AWS configuration
	// ---------------------------------------------------------

	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.Region),
	)

	if err != nil {
		log.Fatalf(
			"load AWS configuration: %v",
			err,
		)
	}

	// ---------------------------------------------------------
	// 3. Detect AWS account
	// ---------------------------------------------------------

	stsClient := sts.NewFromConfig(awsCfg)

	identity, err :=
		stsClient.GetCallerIdentity(
			ctx,
			&sts.GetCallerIdentityInput{},
		)

	if err != nil {
		log.Fatalf(
			"get AWS caller identity: %v",
			err,
		)
	}

	// ---------------------------------------------------------
	// 4. Determine requested action
	//
	// Supported:
	//
	// go run .\cmd\infrastructure create
	// go run .\cmd\infrastructure destroy
	//
	// Default action = create
	// ---------------------------------------------------------

	action := "create"

	if len(os.Args) > 1 {
		action = strings.ToLower(
			strings.TrimSpace(os.Args[1]),
		)
	}

	// ---------------------------------------------------------
	// 5. DESTROY MODE
	// ---------------------------------------------------------

	if action == "destroy" {

		fmt.Println()
		fmt.Println(
			"============================================================",
		)
		fmt.Println(
			"Snowflake Operator - AWS Infrastructure Destroy",
		)
		fmt.Println(
			"============================================================",
		)

		fmt.Printf(
			"AWS Account        : %s\n",
			aws.ToString(identity.Account),
		)

		fmt.Printf(
			"AWS Region         : %s\n",
			cfg.Region,
		)

		fmt.Printf(
			"ECR Repository     : %s\n",
			cfg.ECRRepository,
		)

		fmt.Printf(
			"EKS Cluster        : %s\n",
			cfg.ClusterName,
		)

		fmt.Printf(
			"Node Group         : %s\n",
			cfg.NodeGroupName,
		)

		fmt.Printf(
			"Cluster Role       : %s\n",
			cfg.ClusterRoleName,
		)

		fmt.Printf(
			"Node Role          : %s\n",
			cfg.NodeRoleName,
		)

		fmt.Println(
			"============================================================",
		)

		// -----------------------------------------------------
		// Destroy safety check
		// -----------------------------------------------------

		if os.Getenv("ALLOW_INFRA_DESTROY") != "true" {

			log.Fatal(
				"set ALLOW_INFRA_DESTROY=true before destroying infrastructure",
			)
		}

		fmt.Println()
		fmt.Println(
			"WARNING: Infrastructure destroy has been enabled.",
		)

		fmt.Printf(
			"Target cluster: %s\n",
			cfg.ClusterName,
		)

		fmt.Println()

		// -----------------------------------------------------
		// Destroy infrastructure
		// -----------------------------------------------------

		if err := awsinfra.DestroyInfrastructure(
			ctx,
			awsCfg,
			cfg,
		); err != nil {

			log.Fatalf(
				"destroy infrastructure: %v",
				err,
			)
		}

		fmt.Println()
		fmt.Println(
			"============================================================",
		)
		fmt.Println(
			"Infrastructure destroyed successfully",
		)
		fmt.Println(
			"============================================================",
		)

		return
	}

	// ---------------------------------------------------------
	// 6. Validate action
	// ---------------------------------------------------------

	if action != "create" {

		log.Fatalf(
			"invalid action %q; supported actions are create and destroy",
			action,
		)
	}

	// ---------------------------------------------------------
	// 7. CREATE MODE - Display effective configuration
	// ---------------------------------------------------------

	fmt.Println()
	fmt.Println(
		"============================================================",
	)
	fmt.Println(
		"Snowflake Operator - AWS Infrastructure Bootstrap",
	)
	fmt.Println(
		"============================================================",
	)

	fmt.Printf(
		"AWS Account        : %s\n",
		aws.ToString(identity.Account),
	)

	fmt.Printf(
		"AWS Region         : %s\n",
		cfg.Region,
	)

	fmt.Printf(
		"ECR Repository     : %s\n",
		cfg.ECRRepository,
	)

	fmt.Printf(
		"EKS Cluster        : %s\n",
		cfg.ClusterName,
	)

	fmt.Printf(
		"Kubernetes Version : %s\n",
		cfg.KubernetesVersion,
	)

	fmt.Printf(
		"Node Group         : %s\n",
		cfg.NodeGroupName,
	)

	fmt.Printf(
		"Instance Types     : %v\n",
		cfg.InstanceTypes,
	)

	fmt.Printf(
		"Desired Nodes      : %d\n",
		cfg.DesiredNodes,
	)

	fmt.Printf(
		"Minimum Nodes      : %d\n",
		cfg.MinNodes,
	)

	fmt.Printf(
		"Maximum Nodes      : %d\n",
		cfg.MaxNodes,
	)

	fmt.Println(
		"============================================================",
	)

	// ---------------------------------------------------------
	// 8. Create safety check
	// ---------------------------------------------------------

	if os.Getenv("ALLOW_INFRA_CREATE") != "true" {

		log.Fatal(
			"set ALLOW_INFRA_CREATE=true before provisioning",
		)
	}

	// ---------------------------------------------------------
	// 9. Create AWS clients
	// ---------------------------------------------------------

	ec2Client :=
		ec2.NewFromConfig(awsCfg)

	iamClient :=
		iam.NewFromConfig(awsCfg)

	ecrClient :=
		ecr.NewFromConfig(awsCfg)

	eksClient :=
		eks.NewFromConfig(awsCfg)

	// ---------------------------------------------------------
	// 10. Create / reuse network
	// ---------------------------------------------------------

	fmt.Println()
	fmt.Println(
		"[1/6] Creating/reusing VPC and subnets...",
	)

	network, err :=
		awsinfra.EnsureNetwork(
			ctx,
			ec2Client,
			cfg,
		)

	if err != nil {
		log.Fatalf(
			"network: %v",
			err,
		)
	}

	fmt.Printf(
		"VPC: %s\n",
		network.VpcID,
	)

	fmt.Printf(
		"Subnets: %v\n",
		network.SubnetIDs,
	)

	// ---------------------------------------------------------
	// 11. Create / reuse EKS cluster IAM role
	// ---------------------------------------------------------

	fmt.Println()
	fmt.Println(
		"[2/6] Creating/reusing EKS IAM roles...",
	)

	clusterRoleARN, err :=
		awsinfra.EnsureClusterRole(
			ctx,
			iamClient,
			cfg.ClusterRoleName,
		)

	if err != nil {
		log.Fatalf(
			"EKS cluster IAM role: %v",
			err,
		)
	}

	fmt.Printf(
		"Cluster Role: %s\n",
		clusterRoleARN,
	)

	// ---------------------------------------------------------
	// 12. Create / reuse node IAM role
	// ---------------------------------------------------------

	nodeRoleARN, err :=
		awsinfra.EnsureNodeRole(
			ctx,
			iamClient,
			cfg.NodeRoleName,
		)

	if err != nil {
		log.Fatalf(
			"EKS node IAM role: %v",
			err,
		)
	}

	fmt.Printf(
		"Node Role: %s\n",
		nodeRoleARN,
	)

	// ---------------------------------------------------------
	// 13. Create / reuse ECR
	// ---------------------------------------------------------

	fmt.Println()
	fmt.Println(
		"[3/6] Creating/reusing ECR repository...",
	)

	repoURI, err :=
		awsinfra.EnsureECRRepository(
			ctx,
			ecrClient,
			cfg.ECRRepository,
		)

	if err != nil {
		log.Fatalf(
			"ECR: %v",
			err,
		)
	}

	fmt.Printf(
		"ECR: %s\n",
		repoURI,
	)

	// ---------------------------------------------------------
	// 14. Create / reuse EKS cluster
	// ---------------------------------------------------------

	fmt.Println()
	fmt.Println(
		"[4/6] Creating/reusing EKS cluster...",
	)

	cluster, err :=
		awsinfra.EnsureEKSCluster(
			ctx,
			eksClient,
			cfg,
			clusterRoleARN,
			network.SubnetIDs,
		)

	if err != nil {
		log.Fatalf(
			"EKS: %v",
			err,
		)
	}

	fmt.Printf(
		"EKS Cluster: %s\n",
		aws.ToString(cluster.Name),
	)

	fmt.Printf(
		"EKS Endpoint: %s\n",
		aws.ToString(cluster.Endpoint),
	)

	// ---------------------------------------------------------
	// 15. Fetch EKS OIDC issuer
	// ---------------------------------------------------------

	var oidcIssuer string

	if cluster.Identity != nil &&
		cluster.Identity.Oidc != nil {

		oidcIssuer =
			aws.ToString(
				cluster.Identity.Oidc.Issuer,
			)
	}

	if oidcIssuer == "" {

		log.Fatal(
			"EKS cluster did not return an OIDC issuer",
		)
	}

	fmt.Printf(
		"OIDC Issuer: %s\n",
		oidcIssuer,
	)

	// ---------------------------------------------------------
	// 16. Create / reuse managed node group
	// ---------------------------------------------------------

	fmt.Println()
	fmt.Println(
		"[5/6] Creating/reusing managed node group...",
	)

	nodeGroup, err :=
		awsinfra.EnsureNodeGroup(
			ctx,
			eksClient,
			cfg,
			nodeRoleARN,
			network.SubnetIDs,
		)

	if err != nil {
		log.Fatalf(
			"node group: %v",
			err,
		)
	}

	// ---------------------------------------------------------
	// 17. Final result
	// ---------------------------------------------------------

	fmt.Println()
	fmt.Println(
		"[6/6] Infrastructure validation complete.",
	)

	fmt.Println()
	fmt.Println(
		"============================================================",
	)

	fmt.Println(
		"Infrastructure ready",
	)

	fmt.Println(
		"============================================================",
	)

	fmt.Printf(
		"AWS Account      : %s\n",
		aws.ToString(identity.Account),
	)

	fmt.Printf(
		"Region           : %s\n",
		cfg.Region,
	)

	fmt.Printf(
		"VPC              : %s\n",
		network.VpcID,
	)

	fmt.Printf(
		"Subnets          : %v\n",
		network.SubnetIDs,
	)

	fmt.Printf(
		"Cluster Role ARN : %s\n",
		clusterRoleARN,
	)

	fmt.Printf(
		"Node Role ARN    : %s\n",
		nodeRoleARN,
	)

	fmt.Printf(
		"ECR URI          : %s\n",
		repoURI,
	)

	fmt.Printf(
		"EKS Cluster      : %s\n",
		aws.ToString(cluster.Name),
	)

	fmt.Printf(
		"EKS Endpoint     : %s\n",
		aws.ToString(cluster.Endpoint),
	)

	fmt.Printf(
		"OIDC Issuer      : %s\n",
		oidcIssuer,
	)

	fmt.Printf(
		"Node Group       : %s (%s)\n",
		aws.ToString(
			nodeGroup.NodegroupName,
		),
		nodeGroup.Status,
	)

	fmt.Println(
		"============================================================",
	)
}
