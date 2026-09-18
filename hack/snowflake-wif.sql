-- Replace placeholders before execution.
-- The SUBJECT must match the Kubernetes service account used by the Deployment.
CREATE USER SNOWFLAKE_OPERATOR_SERVICE
  WORKLOAD_IDENTITY = (
    TYPE = OIDC
    ISSUER = 'https://oidc.eks.<region>.amazonaws.com/id/<issuer-id>'
    SUBJECT = 'system:serviceaccount:snowflake-system:snowflake-operator'
  )
  TYPE = SERVICE;

-- Grant only what the controller needs. Start with ROLE management for the first test.
-- Example role hierarchy depends on your account security model.
GRANT ROLE <OPERATOR_ADMIN_ROLE> TO USER SNOWFLAKE_OPERATOR_SERVICE;
