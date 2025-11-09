#!/usr/bin/env bash
set -euo pipefail

show_help() {
  cat <<EOF
Usage: $0 <command> [options]

Available commands:
  init               初始化 KubeApp Operator 项目
  deploy             构建并部署 Operator 镜像
  test [opts]        创建 KubeApp 实例（支持传参）

Global options:
  -h, --help         Show this help message

For command-specific help:
  $0 <command> --help
EOF
}


init(){
  export OS=$(uname | awk '{print tolower($0)}')
  export ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
  export OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.41.1
  curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}
  chmod +x operator-sdk_${OS}_${ARCH} && sudo mv operator-sdk_${OS}_${ARCH} /usr/local/bin/operator-sdk
  operator-sdk init --repo=github.com/k8s/kube-app-operator --domain=kube.com --skip-go-version-check
  operator-sdk create api --group=apps --version=v1alpha1 --kind=KubeApp --resource --controller
}

deploy(){
  timestamp=$(date +"%Y%m%d_%H%M")
  make generate
  make manifests
  make docker-build docker-push IMG=yourrepo:$timestamp
  make deploy IMG=yourrepo:$timestamp
}

test(){
  cat >> kubeApp-crd.yaml << 'EOF'
apiVersion: apps.kube.com/v1alpha1
kind: KubeApp
metadata:
  name: KubeApp-sample
spec:
  replicas: 2
  image: nginx:latest
EOF
  kubectl apply -f kubeApp-crd.yaml
}

# 无默认输出提示，只根据第一个参数 dispatch 函数
case "${1:-}" in
  init)   init ;;
  deploy) deploy ;;
  test)   test ;;
  *) show_help
    ;;
esac

