echo "Setting Vault env variables"
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN=

echo "Creating read only policy for data path"
vault policy write "example-policy" -<<EOF
path "secret/data/*" {
  capabilities = ["read"]
}
EOF

echo "Storing a sample secret at secret/data/foo"
vault kv put secret/data/foo pass=pass

echo "Enabling AWS Auth"
vault auth enable -path aws -description "IAM Auth for AWS" aws

echo "Enabling theft protection"
vault write auth/aws/config/client iam_server_id_header_value=vault.example.com

echo "Assigning the vault IAM role to the policy"
vault write \
  auth/aws/role/example-role-name \
  auth_type=iam \
  policies=example-policy \
  max_ttl=500h \
  bound_iam_principal_arn=arn:aws:iam::1234567890:role/vault-ec2-user-verify