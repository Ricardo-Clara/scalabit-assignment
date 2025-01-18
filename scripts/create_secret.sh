set -o allexport
source config.env
set +o allexport

kubectl create secret generic github-secrets \
  --from-literal=GITHUB_TOKEN="$TOKEN" \
  --from-literal=GITHUB_OWNER="$OWNER"