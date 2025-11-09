# kube-app-operator
// TODO(user): Add simple overview of use/purpose

## Description
// TODO(user): An in-depth paragraph about your project and overview of use

## Getting Started

### Prerequisites
- go version v1.24.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/kube-app-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/kube-app-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/kube-app-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/kube-app-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
operator-sdk edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
   can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.


# user used  api

./auto-client-jwt

#api post create
curl -X POST http://127.0.0.1:8088/api/v1/apps/create -H "Content-Type: application/json" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTYzNjAwNzUsInVzZXJfaWQiOjEyM30.X7FTzKNHr311vU2eUtLhZzd05C1v4lqoLT4QXuEotUY" -d '{ "name": "nginx-app-auto", "namespace": "default", "templateType": "backend", "image":"nginx:latest","replicas":2}'


#delete app
curl -X POST http://127.0.0.1:8088/api/v1/apps/delete -H "Content-Type: application/json" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTYzNjAwNzUsInVzZXJfaWQiOjEyM30.X7FTzKNHr311vU2eUtLhZzd05C1v4lqoLT4QXuEotUY" -d '{ "name": "nginx-app-auto", "namespace": "default",  "deleteKubeApp": true}'

#delete deployment
curl -X POST http://127.0.0.1:8088/api/v1/apps/delete -H "Content-Type: application/json" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTYzNjAwNzUsInVzZXJfaWQiOjEyM30.X7FTzKNHr311vU2eUtLhZzd05C1v4lqoLT4QXuEotUY" -d '{ "name": "nginx-app-auto", "namespace": "default",  "deleteDeployment": true}'


#delete service
curl -X POST http://127.0.0.1:8088/api/v1/apps/delete -H "Content-Type: application/json" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTYzNjAwNzUsInVzZXJfaWQiOjEyM30.X7FTzKNHr311vU2eUtLhZzd05C1v4lqoLT4QXuEotUY" -d '{ "name": "nginx-app-auto", "namespace": "default",  "deleteService": true}'


#delete ingress
curl -X POST http://127.0.0.1:8088/api/v1/apps/delete -H "Content-Type: application/json" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTYzNjAwNzUsInVzZXJfaWQiOjEyM30.X7FTzKNHr311vU2eUtLhZzd05C1v4lqoLT4QXuEotUY" -d '{ "name": "nginx-app-auto", "namespace": "default",  "deleteIngress": true}


#delete pvc
curl -X POST http://127.0.0.1:8088/api/v1/apps/delete -H "Content-Type: application/json" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTYzNjAwNzUsInVzZXJfaWQiOjEyM30.X7FTzKNHr311vU2eUtLhZzd05C1v4lqoLT4QXuEotUY" -d '{ "name": "nginx-app-auto", "namespace": "default",  "deletePvc": true}


#user manager
curl -X POST http://127.0.0.1:8088/users  -H "Content-Type: application/json"   -d '{"name":"lisi","email":"lisi@example.com","password":"123456","groups":"dev"}'
curl -X PUT http://127.0.0.1:8088/users/5c459d9d-4cdd-4a7d-a62e-556902ee91fe  -H "Content-Type: application/json"   -d '{"name":"zhangsan01","email":"zhangsan01@example.com","password":"1234567","groups":"dev"}'
curl  http://127.0.0.1:8088/users/9138caed-b4f5-43cb-8819-7515a5853032
curl -X DELETE  http://127.0.0.1:8088/users/9af4c7d4-0670-4814-affd-df999b5320e32

#查询用户角色
curl -X GET http://localhost:8088/users/742d0212-b549-4702-8a86-193ba12297b0 \
-H "Content-Type: application/json"

# 创建用户不带角色
curl -X POST http://localhost:8080/users \
-H "Content-Type: application/json" \
-d '{
"name": "alice",
"email": "alice@example.com",
"password": "123456"
}'

# 创建用户（带角色）
curl -X POST http://localhost:8080/users \
-H "Content-Type: application/json" \
-d '{
"name": "bob",
"email": "bob@example.com",
"password": "123456",
"roles": ["admin","sre"]
}'

# 更新用户（比如改名 + 改角色）
curl -X PUT http://localhost:8088/users/USER_ID_HERE \
-H "Content-Type: application/json" \
-d '{
"name": "bob-renamed",
"roles": ["ops","k8s"]
}'

# 给已有用户绑定角色
curl -X POST http://localhost:8088/users/USER_ID_HERE/roles \
-H "Content-Type: application/json" \
-d '{
"roles": ["ops","sre"]
}'

# 给已有用户解绑角色

curl -X DELETE http://localhost:8088/users/USER_ID_HERE/roles \
-H "Content-Type: application/json" \
-d '{
"roles": ["sre"]
}'



#login
curl -X POST http://127.0.0.1:8088/login \
-H "Content-Type: application/json"
-d '{"email": "lisi@example.com","password": "123456"}'


#Approve

curl -X POST http://127.0.0.1:8088/approvals  \
-H "Content-Type: application/json" \
-d '{
"created_by": "lisi",
"business_line": "支付系统",
"service_name": "payment-service",
"image": "registry.example.com/payment:v1.2.3",
"replicas": 1,
"template_name": "k8s-deployment-template",
"purpose": "上线支付服务新版本"
}'

#accept

curl -X POST http://127.0.0.1:8088/approvals/d6ac2bfd-6fbc-4b5a-ba54-04a64bc84e3d/approve \
-H "Content-Type: application/json" \
-d '{
"approver_role": "OPS",
"approver_name": "bob",
"decision": "APPROVE",
"comment": "运维审核通过"
}'


// OPS 审批通过 → 状态变成 OPS_APPROVED 下一步必须是 SRE 审批。

curl -X POST http://127.0.0.1:8088/approvals/<request_id>/approve \
-H "Content-Type: application/json" \
-d '{
"approver_role": "SRE",
"approver_name": "alice",
"decision": "APPROVE",
"comment": "SRE 审核通过"
}'


// SRE 审批通过 → 状态变成 SRE_APPROVED 下一步必须是 K8S 审批。

curl -X POST http://127.0.0.1:8088/approvals/<request_id>/approve \
-H "Content-Type: application/json" \
-d '{
"approver_role": "K8S",
"approver_name": "tom",
"decision": "APPROVE",
"comment": "K8S 审核通过，允许上线"
}'

curl -X POST http://127.0.0.1:8088/approvals/<request_id>/approve \


#reject
curl -X POST "http://127.0.0.1:8088/approvals/d6ac2bfd-6fbc-4b5a-ba54-04a64bc84e3d/reject" \
-H "Content-Type: application/json" \
-d '{
"approver_role": "OPS",
"approver_name": "alice",
"decision": "REJECT",
"comment": "测试发现配置有问题，拒绝上线"
}'


// 当前状态必须是 OPS_APPROVED，才能让 SRE 进行审批。

curl -X POST "http://127.0.0.1:8088/approvals/d6ac2bfd-6fbc-4b5a-ba54-04a64bc84e3d/reject" \
-H "Content-Type: application/json" \
-d '{
"approver_role": "SRE",
"approver_name": "charlie",
"decision": "REJECT",
"comment": "安全检查未通过，拒绝上线"
}'



curl -X POST http://127.0.0.1:8088/approvals/<request_id>/approve \


#list
curl -X POST http://127.0.0.1:8088/approvals/list?page=1&page_size=10


#batch
curl -X POST http://127.0.0.1:8088/approvals/batch   -H "Content-Type: application/json" -d '{"request_ids": ["req-123", "req-456", "req-789"]}'


#temple create
curl --location 'http://localhost:8088/templates/create' \
--header 'Content-Type: application/json' \
--data '{
"name": "nginx-template",
"type": "deployment",
"description": "基础 Nginx 模版",
"content": {
"replicas": 2,
"image": "nginx:1.25",
"port": 80
}
}'

#temple update
curl --location --request PUT 'http://localhost:8088/templates/update/1' \
--header 'Content-Type: application/json' \
--data '{
"name": "nginx-template",
"type": "deployment",
"description": "更新后的 Nginx 模版（带环境变量）",
"content": {
"port": 80,
"image": "nginx:1.27",
"replicas": 3,
"env": {
"LOG_LEVEL": "debug",
"TZ": "Asia/Shanghai"
}
}
}'
# id 查询模版
curl --location 'http://localhost:8088/templates/1' \
--header 'Content-Type: application/json' \
--data ''

#删除模版

curl --location --request DELETE 'http://localhost:8088/templates/delete/1' \
--header 'Content-Type: application/json' \
--data ''


#创建应用
curl --location 'http://localhost:8088/apps/create' \
--header 'Content-Type: application/json' \
--data '{
"name": "demo-app",
"namespace": "uat",
"image": "nginx:latest",
"replicas": 2,
"template_id": 2
}'

#查询应用
curl --location 'http://localhost:8088/apps/list'

#更新应用
curl --location --request PUT 'http://localhost:8088/apps/update/3' \
--header 'Content-Type: application/json' \
--data '{
"name": "demo-app-updated",
"image": "nginx:1.27.0",
"replicas": 3,
"template_id": 2
}'

#删除应用
curl --location --request DELETE 'http://localhost:8088/apps/delete/2'




