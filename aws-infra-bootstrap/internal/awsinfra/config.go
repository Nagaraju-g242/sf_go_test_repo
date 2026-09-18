package awsinfra

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Region            string
	ClusterName       string
	KubernetesVersion string
	ECRRepository     string

	VpcCIDR           string
	PublicSubnet1CIDR string
	PublicSubnet2CIDR string

	ClusterRoleName string
	NodeRoleName    string

	NodeGroupName string
	InstanceTypes []string

	DesiredNodes int32
	MinNodes     int32
	MaxNodes     int32
}

func LoadConfigFromEnv() (Config, error) {

	cfg := Config{
		Region: requiredEnv("AWS_REGION"),

		ClusterName: requiredEnv(
			"EKS_CLUSTER_NAME",
		),

		KubernetesVersion: requiredEnv(
			"EKS_KUBERNETES_VERSION",
		),

		ECRRepository: requiredEnv(
			"ECR_REPOSITORY",
		),

		VpcCIDR: requiredEnv(
			"VPC_CIDR",
		),

		PublicSubnet1CIDR: requiredEnv(
			"PUBLIC_SUBNET_1_CIDR",
		),

		PublicSubnet2CIDR: requiredEnv(
			"PUBLIC_SUBNET_2_CIDR",
		),

		ClusterRoleName: requiredEnv(
			"EKS_CLUSTER_ROLE_NAME",
		),

		NodeRoleName: requiredEnv(
			"EKS_NODE_ROLE_NAME",
		),

		NodeGroupName: requiredEnv(
			"EKS_NODEGROUP_NAME",
		),

		InstanceTypes: splitCSV(
			requiredEnv("EKS_INSTANCE_TYPES"),
		),
	}

	var err error

	cfg.DesiredNodes, err =
		requiredInt32("EKS_DESIRED_NODES")
	if err != nil {
		return Config{}, err
	}

	cfg.MinNodes, err =
		requiredInt32("EKS_MIN_NODES")
	if err != nil {
		return Config{}, err
	}

	cfg.MaxNodes, err =
		requiredInt32("EKS_MAX_NODES")
	if err != nil {
		return Config{}, err
	}

	if cfg.MinNodes > cfg.DesiredNodes ||
		cfg.DesiredNodes > cfg.MaxNodes {

		return Config{}, fmt.Errorf(
			"invalid scaling configuration: require min <= desired <= max",
		)
	}

	if len(cfg.InstanceTypes) == 0 {
		return Config{}, fmt.Errorf(
			"EKS_INSTANCE_TYPES must contain at least one instance type",
		)
	}

	return cfg, nil
}

func requiredEnv(key string) string {

	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		panic(
			fmt.Sprintf(
				"required environment variable %s is not set",
				key,
			),
		)
	}

	return value
}

func requiredInt32(key string) (int32, error) {

	raw := strings.TrimSpace(os.Getenv(key))

	if raw == "" {
		return 0, fmt.Errorf(
			"required environment variable %s is not set",
			key,
		)
	}

	value, err :=
		strconv.ParseInt(raw, 10, 32)

	if err != nil {
		return 0, fmt.Errorf(
			"%s must be an integer: %w",
			key,
			err,
		)
	}

	return int32(value), nil
}

func splitCSV(value string) []string {

	parts := strings.Split(value, ",")

	result :=
		make([]string, 0, len(parts))

	for _, part := range parts {

		part = strings.TrimSpace(part)

		if part != "" {
			result = append(
				result,
				part,
			)
		}
	}

	return result
}
