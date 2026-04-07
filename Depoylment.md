# 🏴‍☠️ Strawhats — Deployment Guide
### Zoro (Go, port 3000) & Sanji (Python/FastAPI, port 4000) on AWS ECS Fargate, Nami (React) on Vercel

---

## Architecture

```
Internet ──► Cloudflare Tunnel (HTTPS) ──► Zoro (ECS) ──► Sanji (ECS, internal)
                  api.visora.me                    sanji-genai.strawhats.local

Internet ──► Vercel (HTTPS)
                  visora.me
```

- **Nami** (React frontend) → Vercel, served at `visora.me`
- **Zoro** (Go backend) → ECS Fargate, exposed via Cloudflare Tunnel at `https://api.visora.me`
- **Sanji** (FastAPI GenAI) → ECS Fargate, internal only — reachable by Zoro via `sanji-genai.strawhats.local` (AWS Cloud Map)
- **Cloudflare Tunnel** runs as a sidecar container in Zoro's task — provides free HTTPS without an ALB
- **AWS Cloud Map** provides service discovery — Sanji gets a stable DNS name that auto-updates when its IP changes

---

## Before You Start

- [ ] AWS CLI installed and configured (`aws configure`)
- [ ] Docker running locally
- [ ] Monorepo with `/backend`, `/genAI`, and `/frontend` folders
- [ ] `.env` file with all secrets
- [ ] A domain on Cloudflare (we use `visora.me`)
- [ ] Cloudflare Tunnel token (see Step 11)

Set your region. Run this at the start of every terminal session:
```bash
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export AWS_REGION=eu-central-1
export ECR_BASE=$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com
echo $AWS_ACCOUNT_ID
```

---

## Step 1 — Create ECR Repositories

```bash
aws ecr create-repository --repository-name zoro-backend --region $AWS_REGION
aws ecr create-repository --repository-name sanji-genai --region $AWS_REGION
```

---

## Step 2 — Build & Push Images to ECR

> ⚠️ **Apple Silicon users (M1/M2/M3):** You MUST use `--platform linux/amd64` or Fargate will fail with `image Manifest does not contain descriptor matching platform 'linux/amd64'`.

```bash
aws ecr get-login-password --region $AWS_REGION | \
  docker login --username AWS --password-stdin $ECR_BASE

# Build for linux/amd64 (required for Fargate)
docker build --platform linux/amd64 -f Dockerfile-Zoro -t zoro-backend ./backend
docker build --platform linux/amd64 -f Dockerfile-Sanji -t sanji-genai ./genAI

docker tag zoro-backend:latest $ECR_BASE/zoro-backend:latest
docker tag sanji-genai:latest $ECR_BASE/sanji-genai:latest

docker push $ECR_BASE/zoro-backend:latest
docker push $ECR_BASE/sanji-genai:latest
```

---

## Step 3 — Create IAM Task Execution Role

```bash
aws iam create-role \
  --role-name ecsTaskExecutionRole \
  --assume-role-policy-document '{
    "Version": "2012-10-17",
    "Statement": [{
      "Effect": "Allow",
      "Principal": { "Service": "ecs-tasks.amazonaws.com" },
      "Action": "sts:AssumeRole"
    }]
  }'

aws iam attach-role-policy \
  --role-name ecsTaskExecutionRole \
  --policy-arn arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy

aws iam attach-role-policy \
  --role-name ecsTaskExecutionRole \
  --policy-arn arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess
```

> ⚠️ **KMS Decrypt permission:** SSM SecureString parameters are encrypted with KMS. You MUST add decrypt permission or tasks will fail with `unable to pull secrets from ssm`:

```bash
aws iam put-role-policy --role-name ecsTaskExecutionRole --policy-name SSMDecryptPolicy --policy-document '{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["ssm:GetParameters", "kms:Decrypt"],
      "Resource": "*"
    }
  ]
}'
```

---

## Step 4 — Create ECS Cluster

```bash
aws ecs create-cluster --cluster-name strawhats --region $AWS_REGION
```

---

## Step 5 — Create CloudWatch Log Groups

```bash
aws logs create-log-group --log-group-name /ecs/zoro-backend --region $AWS_REGION
aws logs create-log-group --log-group-name /ecs/sanji-genai --region $AWS_REGION
```

---

## Step 6 — Store Secrets in SSM Parameter Store

> ⚠️ **Region matters!** Store secrets in the SAME region as your ECS cluster. Mismatched regions cause `invalid ssm parameters` errors.

### Zoro secrets:
```bash
aws ssm put-parameter --name "/strawhats/zoro/DATABASE_CONNECTION_STRING" \
  --value "your-db-url" --type SecureString --region $AWS_REGION

aws ssm put-parameter --name "/strawhats/zoro/ENCRYPTION_SECRET_KEY" \
  --value "your-secret-key" --type SecureString --region $AWS_REGION

aws ssm put-parameter --name "/strawhats/zoro/BUCKET_ENDPOINT_STRING" \
  --value "your-bucket-endpoint" --type SecureString --region $AWS_REGION
```

### Sanji secrets:
```bash
aws ssm put-parameter --name "/strawhats/sanji/GEMINI_API_KEY" \
  --value "your-key" --type SecureString --region $AWS_REGION

aws ssm put-parameter --name "/strawhats/sanji/GEMINI_MODEL" \
  --value "gemini-2.5-flash" --type SecureString --region $AWS_REGION

aws ssm put-parameter --name "/strawhats/sanji/GROQ_API_KEY" \
  --value "your-key" --type SecureString --region $AWS_REGION

aws ssm put-parameter --name "/strawhats/sanji/GROQ_MODEL" \
  --value "llama-3.3-70b-versatile" --type SecureString --region $AWS_REGION

aws ssm put-parameter --name "/strawhats/sanji/OCR_API_KEY" \
  --value "your-key" --type SecureString --region $AWS_REGION

aws ssm put-parameter --name "/strawhats/sanji/OCR_MODEL_ID" \
  --value "your-model-id" --type SecureString --region $AWS_REGION
```

### Verify they exist:
```bash
aws ssm get-parameters \
  --names "/strawhats/zoro/DATABASE_CONNECTION_STRING" "/strawhats/zoro/ENCRYPTION_SECRET_KEY" "/strawhats/zoro/BUCKET_ENDPOINT_STRING" \
  --region $AWS_REGION --query '{found:Parameters[*].Name,invalid:InvalidParameters}'

aws ssm get-parameters \
  --names "/strawhats/sanji/GEMINI_API_KEY" "/strawhats/sanji/GEMINI_MODEL" "/strawhats/sanji/GROQ_API_KEY" "/strawhats/sanji/GROQ_MODEL" "/strawhats/sanji/OCR_API_KEY" "/strawhats/sanji/OCR_MODEL_ID" \
  --region $AWS_REGION --query '{found:Parameters[*].Name,invalid:InvalidParameters}'
```

---

## Step 7 — Create Task Definitions

> ⚠️ **Critical:** All ARNs (ECR image, SSM parameters, CloudWatch region) must use the SAME region as your ECS cluster. This was the #1 cause of deployment failures.

### 7a. `task-def-zoro.json`

Zoro's task includes a `cloudflared` sidecar container for the Cloudflare Tunnel. This gives HTTPS access without an ALB.

The Go backend reads `GENAI_HOST` and `GENAI_PORT` separately and builds the Sanji URL as:
```go
fmt.Sprintf("http://%s%s%s", cfg.GenAIHost, cfg.GenAIPort, cfg.GenAIUploadEndpoint)
```

So we pass them as separate env vars, NOT a single URL.

```json
{
  "family": "zoro-backend",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::<YOUR_ACCOUNT_ID>:role/ecsTaskExecutionRole",
  "containerDefinitions": [
    {
      "name": "zoro-backend",
      "image": "<YOUR_ACCOUNT_ID>.dkr.ecr.<YOUR_REGION>.amazonaws.com/zoro-backend:latest",
      "portMappings": [
        { "containerPort": 3000, "protocol": "tcp" }
      ],
      "environment": [
        { "name": "ENV", "value": "production" },
        { "name": "BACKEND_PORT", "value": ":3000" },
        { "name": "BACKEND_HOST", "value": "zoro" },
        { "name": "BACKEND_LOGIN_API", "value": "/auth/login" },
        { "name": "BACKEND_SIGNUP_API", "value": "/auth/signup" },
        { "name": "BACKEND_UPLOAD_API", "value": "/uploadreceipt" },
        { "name": "BACKEND_MANUAL_EXPENSE_API", "value": "/manualexpense" },
        { "name": "BACKEND_ANALYTICS_API", "value": "/useranalytics" },
        { "name": "BACKEND_INSIGHTS_API", "value": "/userinsights" },
        { "name": "BACKEND_DAYRECEIPTS_API", "value": "/todayreceipts" },
        { "name": "GENAI_HOST", "value": "sanji-genai.strawhats.local" },
        { "name": "GENAI_PORT", "value": ":4000" },
        { "name": "GENAI_UPLOAD_API", "value": "/uploadreceipt" },
        { "name": "GENAI_GET_ANALYTICS_API", "value": "/getanalytics" },
        { "name": "GENAI_GENERATE_SUMMARY_API", "value": "/generatesummary" }
      ],
      "secrets": [
        {
          "name": "ENCRYPTION_SECRET_KEY",
          "valueFrom": "arn:aws:ssm:<YOUR_REGION>:<YOUR_ACCOUNT_ID>:parameter/strawhats/zoro/ENCRYPTION_SECRET_KEY"
        },
        {
          "name": "DATABASE_CONNECTION_STRING",
          "valueFrom": "arn:aws:ssm:<YOUR_REGION>:<YOUR_ACCOUNT_ID>:parameter/strawhats/zoro/DATABASE_CONNECTION_STRING"
        },
        {
          "name": "BUCKET_ENDPOINT_STRING",
          "valueFrom": "arn:aws:ssm:<YOUR_REGION>:<YOUR_ACCOUNT_ID>:parameter/strawhats/zoro/BUCKET_ENDPOINT_STRING"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/zoro-backend",
          "awslogs-region": "<YOUR_REGION>",
          "awslogs-stream-prefix": "zoro"
        }
      }
    },
    {
      "name": "cloudflared",
      "image": "cloudflare/cloudflared:latest",
      "command": ["tunnel", "--no-autoupdate", "run", "--token", "<YOUR_TUNNEL_TOKEN>"],
      "essential": true,
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/zoro-backend",
          "awslogs-region": "<YOUR_REGION>",
          "awslogs-stream-prefix": "cloudflared"
        }
      }
    }
  ]
}
```

> CPU is 512 and memory is 1024 because Zoro runs two containers (backend + cloudflared).

### 7b. `task-def-sanji.json`

```json
{
  "family": "sanji-genai",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::<YOUR_ACCOUNT_ID>:role/ecsTaskExecutionRole",
  "containerDefinitions": [
    {
      "name": "sanji-genai",
      "image": "<YOUR_ACCOUNT_ID>.dkr.ecr.<YOUR_REGION>.amazonaws.com/sanji-genai:latest",
      "portMappings": [
        { "containerPort": 4000, "protocol": "tcp" }
      ],
      "environment": [
        { "name": "GENAI_HOST", "value": "0.0.0.0" },
        { "name": "GENAI_PORT", "value": ":4000" },
        { "name": "GENAI_UPLOAD_API", "value": "/uploadreceipt" },
        { "name": "GENAI_GENERATE_SUMMARY_API", "value": "/generatesummary" },
        { "name": "GENAI_GET_ANALYTICS_API", "value": "/getanalytics" }
      ],
      "secrets": [
        {
          "name": "GEMINI_API_KEY",
          "valueFrom": "arn:aws:ssm:<YOUR_REGION>:<YOUR_ACCOUNT_ID>:parameter/strawhats/sanji/GEMINI_API_KEY"
        },
        {
          "name": "GEMINI_MODEL",
          "valueFrom": "arn:aws:ssm:<YOUR_REGION>:<YOUR_ACCOUNT_ID>:parameter/strawhats/sanji/GEMINI_MODEL"
        },
        {
          "name": "GROQ_API_KEY",
          "valueFrom": "arn:aws:ssm:<YOUR_REGION>:<YOUR_ACCOUNT_ID>:parameter/strawhats/sanji/GROQ_API_KEY"
        },
        {
          "name": "GROQ_MODEL",
          "valueFrom": "arn:aws:ssm:<YOUR_REGION>:<YOUR_ACCOUNT_ID>:parameter/strawhats/sanji/GROQ_MODEL"
        },
        {
          "name": "OCR_API_KEY",
          "valueFrom": "arn:aws:ssm:<YOUR_REGION>:<YOUR_ACCOUNT_ID>:parameter/strawhats/sanji/OCR_API_KEY"
        },
        {
          "name": "OCR_MODEL_ID",
          "valueFrom": "arn:aws:ssm:<YOUR_REGION>:<YOUR_ACCOUNT_ID>:parameter/strawhats/sanji/OCR_MODEL_ID"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/sanji-genai",
          "awslogs-region": "<YOUR_REGION>",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

### 7c. Replace placeholders and register

```bash
sed -i '' "s/<YOUR_ACCOUNT_ID>/$AWS_ACCOUNT_ID/g" task-def-zoro.json task-def-sanji.json
sed -i '' "s/<YOUR_REGION>/$AWS_REGION/g" task-def-zoro.json task-def-sanji.json

aws ecs register-task-definition --cli-input-json file://task-def-zoro.json --region $AWS_REGION
aws ecs register-task-definition --cli-input-json file://task-def-sanji.json --region $AWS_REGION
```

> On macOS, `sed -i ''` is required (empty string after `-i`). On Linux, use `sed -i` without the `''`.

---

## Step 8 — Networking Setup

Two security groups — Zoro is public-facing, Sanji is internal only.

```bash
export VPC_ID=$(aws ec2 describe-vpcs \
  --filters Name=isDefault,Values=true \
  --query 'Vpcs[0].VpcId' --output text --region $AWS_REGION)

export SUBNETS=$(aws ec2 describe-subnets \
  --filters Name=defaultForAz,Values=true \
  --query 'Subnets[*].SubnetId' --output text --region $AWS_REGION | tr '\t' ',')

# Zoro SG — public on port 3000
export ZORO_SG=$(aws ec2 create-security-group \
  --group-name zoro-sg \
  --description "Zoro — public port 3000" \
  --vpc-id $VPC_ID \
  --query 'GroupId' --output text --region $AWS_REGION)

aws ec2 authorize-security-group-ingress \
  --group-id $ZORO_SG --protocol tcp --port 3000 --cidr 0.0.0.0/0 --region $AWS_REGION

# Sanji SG — only Zoro can reach port 4000
export SANJI_SG=$(aws ec2 create-security-group \
  --group-name sanji-sg \
  --description "Sanji — only reachable by Zoro" \
  --vpc-id $VPC_ID \
  --query 'GroupId' --output text --region $AWS_REGION)

aws ec2 authorize-security-group-ingress \
  --group-id $SANJI_SG --protocol tcp --port 4000 \
  --source-group $ZORO_SG --region $AWS_REGION

echo "Zoro SG: $ZORO_SG"
echo "Sanji SG: $SANJI_SG"
```

---

## Step 9 — Service Discovery (AWS Cloud Map)

Cloud Map gives Sanji a stable DNS name (`sanji-genai.strawhats.local`) that auto-updates when its IP changes. This eliminates the need to manually update Zoro's config on every Sanji redeploy.

```bash
# Create a private DNS namespace
aws servicediscovery create-private-dns-namespace \
  --name strawhats.local \
  --vpc $VPC_ID \
  --region $AWS_REGION
```

Wait ~30 seconds for it to create, then get the namespace ID:

```bash
export NAMESPACE_ID=$(aws servicediscovery list-namespaces --region $AWS_REGION \
  --query 'Namespaces[?Name==`strawhats.local`].Id' --output text)
echo $NAMESPACE_ID
```

Create a service discovery service for Sanji:

```bash
aws servicediscovery create-service \
  --name sanji-genai \
  --dns-config "NamespaceId=$NAMESPACE_ID,DnsRecords=[{Type=A,TTL=10}]" \
  --health-check-custom-config FailureThreshold=1 \
  --region $AWS_REGION
```

Save the `Id` from the output — you'll need it in Step 11.

> After this, Sanji will be reachable at `sanji-genai.strawhats.local` from within the VPC. This is what `GENAI_HOST` in Zoro's task definition should be set to.

---

## Step 10 — Identify the Working Subnet

> ⚠️ **Subnet connectivity issue:** Not all default subnets may have proper internet access. Some subnets can fail to reach ECR or SSM, causing `ResourceInitializationError`. To avoid this, identify a subnet that works and pin both services to it.

```bash
# Pick the first subnet and test with Sanji first
export WORKING_SUBNET=$(echo $SUBNETS | cut -d',' -f1)
echo "Testing subnet: $WORKING_SUBNET"
```

If a service fails with connectivity errors, try the next subnet in the list. Once you find one that works, use it for both services.

---

## Step 11 — Create ECS Services

Deploy Sanji first with service discovery attached, then Zoro.

```bash
export DISCOVERY_SERVICE_ARN=arn:aws:servicediscovery:$AWS_REGION:$AWS_ACCOUNT_ID:service/<DISCOVERY_SERVICE_ID>

# Sanji — with service discovery, pinned to working subnet
aws ecs create-service \
  --cluster strawhats \
  --service-name sanji-genai \
  --task-definition sanji-genai \
  --desired-count 1 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[$WORKING_SUBNET],securityGroups=[$SANJI_SG],assignPublicIp=ENABLED}" \
  --service-registries "registryArn=$DISCOVERY_SERVICE_ARN" \
  --region $AWS_REGION
```

> Replace `<DISCOVERY_SERVICE_ID>` with the `Id` from Step 9.
> `assignPublicIp=ENABLED` is needed so Sanji can reach ECR and SSM. The security group ensures only Zoro can reach port 4000.

Wait ~90 seconds, then verify Sanji is running:
```bash
export SANJI_TASK=$(aws ecs list-tasks --cluster strawhats --service-name sanji-genai \
  --region $AWS_REGION --query 'taskArns[0]' --output text)
aws ecs describe-tasks --cluster strawhats --tasks $SANJI_TASK \
  --region $AWS_REGION --query 'tasks[0].{status:lastStatus,reason:stoppedReason}'
```

Deploy Zoro (no need to get Sanji's IP — Cloud Map handles it via `sanji-genai.strawhats.local`):
```bash
aws ecs create-service \
  --cluster strawhats \
  --service-name zoro-backend \
  --task-definition zoro-backend \
  --desired-count 1 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[$WORKING_SUBNET],securityGroups=[$ZORO_SG],assignPublicIp=ENABLED}" \
  --region $AWS_REGION
```

Verify both are running:
```bash
export ZORO_TASK=$(aws ecs list-tasks --cluster strawhats --service-name zoro-backend \
  --region $AWS_REGION --query 'taskArns[0]' --output text)
aws ecs describe-tasks --cluster strawhats --tasks $SANJI_TASK $ZORO_TASK \
  --region $AWS_REGION --query 'tasks[*].{service:containers[0].name,status:lastStatus,reason:stoppedReason}'
```

Health check Zoro:
```bash
curl https://api.visora.me/health
```

---

## Step 12 — Cloudflare Tunnel Setup (HTTPS without ALB)

> An ALB costs ~$16-22/month. A Cloudflare Tunnel is free and provides HTTPS.

### 11a. Add domain to Cloudflare
1. Go to [dash.cloudflare.com](https://dash.cloudflare.com) → Add site → `visora.me` → Free plan
2. Cloudflare gives you two nameservers
3. Go to your domain registrar → replace nameservers with Cloudflare's
4. Wait for Cloudflare to show domain as active

### 11b. Create the tunnel
1. Cloudflare dashboard → Networking → Tunnels → Create Tunnel
2. Name: `zoro-tunnel`
3. Select Docker → copy the tunnel token
4. Add the token to `task-def-zoro.json` in the `cloudflared` container's `command` array (see Step 7a)

### 11c. Add public hostname route
1. In the tunnel config → Add route → Published application
2. Subdomain: `api`, Domain: `visora.me`
3. Service URL: `http://localhost:3000`

### 11d. DNS records in Cloudflare
The tunnel auto-creates the `api` DNS record. For the frontend, add:
- Type: `A`, Name: `@`, Value: `216.198.79.1`, Proxy: OFF
- Type: `CNAME`, Name: `www`, Value: (get from Vercel domain settings), Proxy: OFF

---

## Step 13 — Deploy Nami on Vercel

1. Go to [vercel.com/new](https://vercel.com/new) → import your GitHub repo
2. Configure:
   - Framework: Vite
   - Root Directory: `frontend`
   - Build Command: `npm run build`
   - Output Directory: `dist`
3. Environment variable:
   - `VITE_API_BASE` = `https://api.visora.me`
4. Deploy
5. Add custom domain: Settings → Domains → add `visora.me`

---

## How to Redeploy After Code Changes

```bash
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export AWS_REGION=eu-central-1
export ECR_BASE=$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

aws ecr get-login-password --region $AWS_REGION | \
  docker login --username AWS --password-stdin $ECR_BASE

# Sanji
docker build --platform linux/amd64 -f Dockerfile-Sanji -t sanji-genai ./genAI
docker tag sanji-genai:latest $ECR_BASE/sanji-genai:latest
docker push $ECR_BASE/sanji-genai:latest
aws ecs update-service --cluster strawhats --service sanji-genai \
  --force-new-deployment --region $AWS_REGION

# Zoro
docker build --platform linux/amd64 -f Dockerfile-Zoro -t zoro-backend ./backend
docker tag zoro-backend:latest $ECR_BASE/zoro-backend:latest
docker push $ECR_BASE/zoro-backend:latest
aws ecs update-service --cluster strawhats --service zoro-backend \
  --force-new-deployment --region $AWS_REGION
```

> **No manual IP updates needed.** Cloud Map automatically resolves `sanji-genai.strawhats.local` to Sanji's current IP.
> **Zoro's public IP also changes**, but since we use Cloudflare Tunnel, the `api.visora.me` URL stays the same — no Vercel update needed.
> **In short:** just rebuild, push, and force redeploy. Everything else is automatic.

---

## Viewing Logs

```bash
aws logs tail /ecs/zoro-backend --since 30m --region $AWS_REGION
aws logs tail /ecs/sanji-genai --since 30m --region $AWS_REGION

# Real-time
aws logs tail /ecs/zoro-backend --follow --region $AWS_REGION
```

---

## Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| `image Manifest does not contain descriptor matching platform 'linux/amd64'` | Built on Apple Silicon without `--platform` flag | Rebuild with `--platform linux/amd64` |
| `invalid ssm parameters` | SSM parameters in wrong region or don't exist | Verify with `aws ssm get-parameters` in the correct region |
| `unable to pull secrets from ssm: connection issue` | Task placed in subnet without internet access | Pin service to a working subnet |
| `unable to pull secrets: kms decrypt` | Missing KMS decrypt permission on execution role | Add `kms:Decrypt` inline policy to `ecsTaskExecutionRole` |
| `ResourceInitializationError: unable to pull registry auth` | Can't reach ECR — subnet/networking issue | Same as above — use a subnet with internet gateway route |
| Mixed content errors in browser | Frontend (HTTPS) calling backend (HTTP) | Use Cloudflare Tunnel for HTTPS on backend |
| `GENAI_HOST` and `GENAI_PORT` swapped | Copy-paste error in task definition | Double-check: HOST = `sanji-genai.strawhats.local`, PORT = `:4000` |

---

## Estimated Monthly Cost

| Resource | Cost |
|---|---|
| Zoro — Fargate (512 CPU / 1024MB) | ~$6-8 |
| Sanji — Fargate (256 CPU / 512MB) | ~$3-4 |
| ECR storage | ~$0.10 |
| CloudWatch Logs | ~$0.50 |
| SSM Parameter Store | Free |
| Cloud Map | Free tier |
| Cloudflare Tunnel | Free |
| Vercel (Hobby) | Free |
| Domain (visora.me) | ~$9/year |
| **Total** | **~$10-13/month** |

---

## Quick Reference

| # | Step | One-time? |
|---|------|-----------|
| 1 | Create ECR repos | ✅ |
| 2 | Build & push images (`--platform linux/amd64`) | 🔁 Every deploy |
| 3 | Create IAM role + KMS decrypt policy | ✅ |
| 4 | Create ECS cluster | ✅ |
| 5 | Create CloudWatch log groups | ✅ |
| 6 | Store secrets in SSM (same region!) | ✅ |
| 7 | Create task definitions | ✅ (update if config changes) |
| 8 | Create security groups | ✅ |
| 9 | Service discovery (Cloud Map) | ✅ |
| 10 | Identify working subnet | ✅ |
| 11 | Deploy ECS services | ✅ (redeploy = just force new deployment) |
| 12 | Cloudflare Tunnel + DNS | ✅ |
| 13 | Deploy Nami on Vercel | ✅ |
