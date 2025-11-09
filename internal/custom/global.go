package define

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	GlobalClient client.Client
	GlobalScheme *runtime.Scheme
)

// Init 用于初始化全局 client 和 scheme

func Init(k8sClient client.Client, scheme *runtime.Scheme) {
	GlobalClient = k8sClient
	GlobalScheme = scheme
}