$env:AWS_REGION = "us-east-2"

$env:ECR_REPOSITORY = "snowflake-operator"

$env:EKS_CLUSTER_NAME = "snowflake-operator-cluster"

$env:EKS_KUBERNETES_VERSION = "1.31"


# Network

$env:VPC_CIDR = "10.30.0.0/16"

$env:PUBLIC_SUBNET_1_CIDR = "10.30.0.0/20"

$env:PUBLIC_SUBNET_2_CIDR = "10.30.16.0/20"


# IAM

$env:EKS_CLUSTER_ROLE_NAME = "snowflake-operator-eks-cluster-role"

$env:EKS_NODE_ROLE_NAME = "snowflake-operator-eks-node-role"


# Node group

$env:EKS_NODEGROUP_NAME = "operator-nodes-large"

$env:EKS_INSTANCE_TYPES = "m7i-flex.large"

$env:EKS_DESIRED_NODES = "2"

$env:EKS_MIN_NODES = "2"

$env:EKS_MAX_NODES = "4"


# Safety

$env:ALLOW_INFRA_CREATE = "false"
$env:ALLOW_INFRA_DESTROY = "true"
