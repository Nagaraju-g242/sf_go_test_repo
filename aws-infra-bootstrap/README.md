# AWS Infrastructure Bootstrap for Snowflake Operator

Creates or reuses an ECR repository, EKS cluster, and EKS managed node group with AWS SDK for Go v2.

## Prerequisites
- AWS credentials already configured.
- At least two suitable subnet IDs.
- Existing EKS cluster IAM role ARN.
- Existing EKS worker-node IAM role ARN.
- Caller must be allowed to call ECR/EKS APIs and `iam:PassRole` for those two roles.

## Setup
```powershell
.\setup.ps1
Copy-Item .\env.example.ps1 .\env.local.ps1
notepad .\env.local.ps1
. .\env.local.ps1
aws sts get-caller-identity
go run .\cmd\infrastructure
```

## Idempotent behavior
- Existing ECR repo: reused.
- Existing ACTIVE EKS cluster: reused.
- Existing ACTIVE node group: reused.
- In-progress cluster/node group: waits until ACTIVE.
- Nothing is deleted.

## After provisioning
```powershell
aws eks update-kubeconfig `
  --name snowflake-operator-cluster `
  --region us-east-2

kubectl get nodes -o wide
```

Use the printed OIDC issuer for Snowflake WIF, then continue with CRDs, RBAC, Kustomize deployment, Prometheus, and Grafana.

Do not run this bootstrap inside the Snowflake Operator pod; run it from the workstation or CI/CD before deploying the operator.
