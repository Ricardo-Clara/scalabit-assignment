set -o allexport
source config.env
set +o allexport

kubectl create secret generic github-secrets \
  --from-literal=TOKEN="$TOKEN" \
  --from-literal=OWNER="$OWNER"