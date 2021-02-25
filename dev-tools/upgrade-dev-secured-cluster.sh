#!/usr/bin/env bash
set -eo pipefail

# Upgrade the secured cluster services deployed by the ./deployed scripts.
# Usage: ./upgrade-dev-secured-cluster <helm-flags...>

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

custom_helm_args=("$@")
secured_cluster_chart_path="$DIR/../deploy/k8s/sensor-deploy/chart"

roxctl helm output secured-cluster-services --output-dir "$secured_cluster_chart_path" --remove

if [[ ! -d "$secured_cluster_chart_path" ]]; then
  echo "Could not find chart: $secured_cluster_chart_path"
  exit 1
fi

helm_args=(
  --reuse-values
)

default_init_bundle_path="$secured_cluster_chart_path/../init-bundle.yaml"
if [[ -f  "$default_init_bundle_path" ]]; then
  echo "Using init bundle: $default_init_bundle_path"
  helm_args+=(
    -f "$default_init_bundle_path"
  )
fi

if [[ "$MAIN_IMAGE_TAG" == "latest-local-build" ]]; then
  MAIN_IMAGE_TAG="$(docker images --filter="reference=stackrox/main" --format "{{.Tag}}" | head -1)"
else
  MAIN_IMAGE_TAG="${MAIN_IMAGE_TAG:-$(make --quiet -C "$DIR/.." tag)}"
fi

helm_args+=(
  --set "image.main.tag=$MAIN_IMAGE_TAG"
  "${custom_helm_args[@]}"
)

helm -n stackrox upgrade --install stackrox-secured-cluster-services "$secured_cluster_chart_path" "${helm_args[@]}"
