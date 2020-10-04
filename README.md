# Purpose
This program can be used to fetch secrets from vault utilizing the IAM role attached to the AWS services.

# Recommendation from Hashicorp Vault.
The iam auth method is the recommended approach as it is more flexible and aligns with best practices to perform access control and authentication.

# Work flow
- AWS STS API includes a method, `sts:GetCallerIdentity`, which allows you to validate the identity of a client.

- The client(this program) signs a `GetCallerIdentity` query using AWS Signature V4 Algorithm and sends it to the vault server

- `GetCallerIdentity` contains 4 pieces of information:
- 1. The request URL
- 2. The request body
- 3. The request headers
- 4. The request method
- AWS signature is computed over these 4 fields.

- The vault server reconstructs the query using this information and forwards it on the AWS STS Service.

- Depending on the response from the STS service, vault authenticates the client and returns a token to perform operations that the role allows.

- IAM Auth method allows you to specify bound IAM principal ARNs.

- Client authenticating to the vault mush have an ARN that matches one of the ARNs bound to the role they are attempting to login to.

- Note: In this example I have used a specify IAM role ARN but it can be configured for IAM account arn as well with wildcards
- eg: `arn:aws:iam::123456789012:*` - would allow any principal in AWS account 123456789012 to login to it.
- eg: `arn:aws:iam::123456789012:role/*` - would allow any IAM role in the AWS account to login to it.

# Minimum IAM Permission
## On the vault server
- 1. The vault server should be able to reach the AWS STS service to validate the signature.
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "iam:GetInstanceProfile"
            ],
            "Resource": "*"
        }
    ]
}
```

- 2. The AWS service should be able to prove its identity for the role attached to it.
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "iam:GetUser",
                "iam:GetRole"
            ],
            "Resource": "*"
        }
    ]
}
```
- If a role is being assumed then the following block would also need to be added.
```
    {
      "Effect": "Allow",
      "Action": ["sts:AssumeRole"],
      "Resource": ["arn:aws:iam::<AccountId>:role/<VaultRole>"]
    }
```

# Mitigations for some common attacks
## Replay attacks
- The request header includes a timestamp and the signature validation expires after 15 minutes - which gives attackers a window of only 15 mins.

## Theft blast radius reduction
- The request header can include a value named - `X-Vault-AWS-IAM-Server-ID` - which can be used to specify which vault server is to be used and we can get segregation based on say dev and prod env and reduce the blast radius.

# Running this program
## Executable
- The executable is a standalone executable that can be run and it can fetch the following values from the environment variables
1. "VAULT_ADDR" - Address of the vault server to fetch the secrets from,
2. "VAULT_AUTH_PROVIDER" - In our case this value is aws. Can be hardcoded in the program if the use case doesn't change.
3. "VAULT_AUTH_ROLE" - The role within vault allowing access to the secret under a path.
4. "VAULT_AUTH_HEADER" (optional but recommended) - To reduce theft attack blast radius - this value can ideally match the vault address.

`./vaultSecret`

## Running the go program (Not recommended)
- For debugging purposes the go program from which the executable is created can be used.
- This would mean pulling in all the dependencies related to vault and aws into the machine.

## Generating an executable
- `sh buildExecutable.sh` - would build executable for linux environment

## Output
- Currently the program outputs the secret to the stdout - but can be configured to write to a location of choice and configured throught the environment variables

# Troubleshooting
## 404 not found data (nil within the program) was being returned when trying to read the data with proper permissions
- If data is stored under path secret/data/foo use secret/data/data/foo within the program vs secret/data/foo which is through the cli.

- Cli appends the path /data automatically.

## Running into IAM Principal "arn:aws:sts::1234567890:assumed-role/vault-ec2-user-verify/i-0f0e3646af1a76f01" does not belong to the role "example-role-name"
- Used Instance Profile ARN instead of Role ARN

# Ref
- https://www.vaultproject.io/docs/auth/aws (Recommended Read for deeper understanding)

- https://learn.hashicorp.com/tutorials/vault/getting-started-policies?in=vault/getting-started

- https://github.com/hashicorp/vault/blob/master/builtin/credential/aws/cli.go

- https://github.com/daveadams/onthelambda/blob/master/onthelambda.go

- https://gist.github.com/jun06t/c5a628abae1cb1562d16f369ca31b22a

- https://blog.gruntwork.io/a-guide-to-automating-hashicorp-vault-3-authenticating-with-an-iam-user-or-role-a3203a3ee088