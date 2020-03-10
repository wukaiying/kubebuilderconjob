### kubebuilder 项目开发
controller开发主要还是参考Kubebuilder官方文档。

### webhook原理
webhook: 它本质上是一种 HTTP 回调，会注册到 apiserver 上。在 apiserver 特定事件发生时，会查询已注册的 webhook，并把相应的消息转发过去。

### 启动webhook流程：
1. make run启动
2.此时contoller service和webhhook service都启动了，webhook service在https://127.0.0.1:9443/mutate-batch-tutorial-kubebuilder-io-v1-cronjob
3.然后修改webhook中manifest.yaml
```
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  creationTimestamp: null
  name: mutating-webhook-configuration
webhooks:
  - clientConfig:
      caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRFNU1UQXhNakF4TURnME9Gb1hEVEk1TVRBd09UQXhNRGcwT0Zvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTWthCkRTY2t4aW51aS9lZEtrTTR4UjM3S1FwbFFxazVyNXV2ZnI4ZVA2dTIrbUw3N3ZoSnRkZElzYWZnNS9kMXJ5eFUKM01EcnN5M0hvSCticGMxR0M4ZkgvTk5VaFlnYmJ6SjkydUpuOUkweEVmNmhZSEpnZzh6NXFOekNmbFhpZGFZTwpEMTIyMzd5NWd4WkZYRzVEMVdaZmVjc2Fad3YrS0V2ZTRMYTRrKzYzUVg3dGVVOWF3bmpjL1d2QXhZdHk4U2tNClhHL0ZFM0lxbGg0cUFLSm9XdHJKSXdoWHRLOVpYeHVWeWdxQXpTbGV4T29sV1hlYXlJMDdKMWswY1JHYXFRajYKeEo3YXpua3dNUkhxNkZnUDlaVVlyWTBzSHdnblE1bllJT2hLcjRmdTExM21MZ0Y2ZDV1QU9SdEVCMzlKYmNTNAorYUlOWUtuYWNCQUs3YTVwWTc4Q0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFDRTJsbEhyK09tTGxWNitLdkl0UEM4VkRndGMKMFRmR2EzMU01V0JFY0svRzZsVWVVN0RGancvZEdGZ3FpeCtQZno2cDFYL052YXIraHRrc3pVYzBnUE9lOVZUVwpvcnpSU0pyZ2Yydld3Z2M3T3pQVzV1T0U1Q1MyZVhxUXdDcXZwdUM2MFZSb0diWGQ4UVQvdWpQZ2JxS05TWStMCmlyR2pBcjkzMmFWMmZEV01TakVxNll3MVJyb3Z3T1gwcmNTaytJaVJIdlV4WjFJZ0lYU0NpNUtkZjNObDhHSjkKR1BiSG1pK1ZBakw0U2pDVXZGTnc5bXZIaHA4anJZWUpqVC9zZGJSOVFRTEFiMEg0VWE0QVdlVVlHckVPU3E0ZgpvZmR0UXpkUTZOOGNORGZxdVNWWjQ5VjU1akJRdFVWbVRCWUNWQjRIblZGMm5wQ1ZjQWNhMkw4c3pKaz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
      url: https://127.0.0.1:9443/mutate-batch-tutorial-kubebuilder-io-v1-cronjob
    failurePolicy: Fail
    name: mcronjob.kb.io
    rules:
      - apiGroups:
          - batch.tutorial.kubebuilder.io
        apiVersions:
          - v1
        operations:
          - CREATE
          - UPDATE
        resources:
          - cronjobs

---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  creationTimestamp: null
  name: validating-webhook-configuration
webhooks:
  - clientConfig:
      caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRFNU1UQXhNakF4TURnME9Gb1hEVEk1TVRBd09UQXhNRGcwT0Zvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTWthCkRTY2t4aW51aS9lZEtrTTR4UjM3S1FwbFFxazVyNXV2ZnI4ZVA2dTIrbUw3N3ZoSnRkZElzYWZnNS9kMXJ5eFUKM01EcnN5M0hvSCticGMxR0M4ZkgvTk5VaFlnYmJ6SjkydUpuOUkweEVmNmhZSEpnZzh6NXFOekNmbFhpZGFZTwpEMTIyMzd5NWd4WkZYRzVEMVdaZmVjc2Fad3YrS0V2ZTRMYTRrKzYzUVg3dGVVOWF3bmpjL1d2QXhZdHk4U2tNClhHL0ZFM0lxbGg0cUFLSm9XdHJKSXdoWHRLOVpYeHVWeWdxQXpTbGV4T29sV1hlYXlJMDdKMWswY1JHYXFRajYKeEo3YXpua3dNUkhxNkZnUDlaVVlyWTBzSHdnblE1bllJT2hLcjRmdTExM21MZ0Y2ZDV1QU9SdEVCMzlKYmNTNAorYUlOWUtuYWNCQUs3YTVwWTc4Q0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFDRTJsbEhyK09tTGxWNitLdkl0UEM4VkRndGMKMFRmR2EzMU01V0JFY0svRzZsVWVVN0RGancvZEdGZ3FpeCtQZno2cDFYL052YXIraHRrc3pVYzBnUE9lOVZUVwpvcnpSU0pyZ2Yydld3Z2M3T3pQVzV1T0U1Q1MyZVhxUXdDcXZwdUM2MFZSb0diWGQ4UVQvdWpQZ2JxS05TWStMCmlyR2pBcjkzMmFWMmZEV01TakVxNll3MVJyb3Z3T1gwcmNTaytJaVJIdlV4WjFJZ0lYU0NpNUtkZjNObDhHSjkKR1BiSG1pK1ZBakw0U2pDVXZGTnc5bXZIaHA4anJZWUpqVC9zZGJSOVFRTEFiMEg0VWE0QVdlVVlHckVPU3E0ZgpvZmR0UXpkUTZOOGNORGZxdVNWWjQ5VjU1akJRdFVWbVRCWUNWQjRIblZGMm5wQ1ZjQWNhMkw4c3pKaz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
      url: https://127.0.0.1:9443/validate-batch-tutorial-kubebuilder-io-v1-cronjob
    failurePolicy: Fail
    name: vcronjob.kb.io
    rules:
      - apiGroups:
          - batch.tutorial.kubebuilder.io
        apiVersions:
          - v1
        operations:
          - CREATE
          - UPDATE
        resources:
          - cronjobs
```

4.此时在终端

```
curl -i -k https://127.0.0.1:9443/validate-batch-tutorial-kubebuilder-io-v1-cronjob
```
可以看到结果，webhook设置成功
HTTP/2 200
content-type: text/plain; charset=utf-8
content-length: 128
date: Sun, 08 Mar 2020 13:00:22 GMT

{"response":{"uid":"","allowed":false,"status":{"metadata":{},"message":"contentType=, expected application/json","code":400}}}
有个问题，妈的跑在mac上，配置webhookvalidation，无法访问到，还是的跑道容器里面。


### 我们自己写webhook配置文件的方式完成
参考.pae/eskort-webhook/chart 中的实现，使用make docker-build会生成controller:latest 镜像，
该镜像中包括controller和webhook，webhook使用端口9443，(向外暴露为443), controller使用端口8443，向外暴露可以自己定义。

我们完全可以抛弃项目自己生成的config来自己定义chart,crd,cr。
查看./pae/eskort-webhook/chart中的cr,crds。
再看templates中的这些文件，主要是deployment.yaml和validatingwebhook-ca-bundle.yaml

- deployment.yaml中主要是启动controller:latest 镜像，绑定serviceaccount。映射tls.key,tls.crt。
volumeMounts: 将secret.yaml中的tls.crt,tls.key映射到/tmp/k8s-webhook-server/serving-certs中，kubebuilder默认回去该目录下读取这些key.
```
containers:
        - name: {{ template "fullname" . }}
          image: {{ .Values.image.repository }}:{{ .Values.image.tag }}
          imagePullPolicy: {{ .Values.image.pullPolicy}}
          volumeMounts:
            - name: webhook-certs
              mountPath: /tmp/k8s-webhook-server/serving-certs
              readOnly: true
            volumes:
                    - name: webhook-certs
                      secret:
                        secretName: {{ template "fullname" . }}-certs
```


- secret.yaml主要是将tls.crt, tls.key部署到集群中，deployment.yaml可以读取到他

- validatingwebhook-ca-bundle.yaml，主要定义validation和mutation如何访问

caBundle：生成证书的三个文件之一，分别是tls.key, tls.crt, 以及caboundle。查看我的映像笔记中的文章（关键字搜索webhook）。唯一 需要要注意的是生成这个三个文件需要用到service的名字，而service的名字与helm chart中value.yaml的override name绑定，所以每次生成的key都是跟name绑定的。

service：name字段指定了webhook service 的名称，path就是程序中kubebuilder的path。

```
webhooks:
  - clientConfig:
      caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN2RENDQWFRQ0NRREFhQ1dNVGxmYmZUQU5CZ2txaGtpRzl3MEJBUXNGQURBZ01SNHdIQVlEVlFRRERCVmwKYzJ0dmNuUXRkMlZpYUc5dmF5MXpkbU1nUTBFd0hoY05NakF3TWpBM01EVXdNVEE1V2hjTk16QXdNakEwTURVdwpNVEE1V2pBZ01SNHdIQVlEVlFRRERCVmxjMnR2Y25RdGQyVmlhRzl2YXkxemRtTWdRMEV3Z2dFaU1BMEdDU3FHClNJYjNEUUVCQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUUMzLytnYXJMY2l0ekZmUjYvS3IwbGRwQ3gvUUVURkJXd2wKSFd2NFBvbUFHZG9TYy9VN2FRSzRybkxROFEySGUxK2JPRGp3Z2lvdDBzS0E0NHFnbks0d1BZb1NCWjJobXRRWQpkT2lNSGM3Z052MWprUFByQlNacE1iQWxRcFlHWm9xT21GWmJlTlVQSGR6R29IajhNVVIvTGpjalJLWHRxRFI3Cjl0SDBUTEtXVWpuTzJ0VGZOSi9KV255bnE5K3BPN2RPZWNWM0dxL05vdUZMU1FxV01JYkZYZHdoaHEzZnhCazUKOWlHUDROb2tJR0ZSUGJRMlNaakRIb0hKRUtjTmsxTjZQZ2Z4dGZPOC9pY2xEcm82TnhFT0tBVnNCRGtMR3Q3VQpIVERJU1d3Q0pyZ1o0Q2Y5M0Jza2s5c2VYM0FuWUpXUndtZU13TUM4RlArSnRjemQ3WUpKQWdNQkFBRXdEUVlKCktvWklodmNOQVFFTEJRQURnZ0VCQUprN3AvaXA2ZGROWm5FSXArQjVZbHFTL21qcXYvTy82VG5YTjlINGdKVDEKamtoeTllTmRQTis4LzJkTjNEVFhqaXF3OVVnZnhtdWxSOFAvbmpjTFo2WGhKNjJSTTdQM1c3Ri95TXNnS2dQcApXaEU3VUtjZlJYUDRvTjV0VGdOSG5rTXVKbUljN3RPdkpRaksrSkZGTm11QkhXc3FuTXpwTzVRNmlNalZORE5zCkFGb09oOGJHZFdwQlE2UEZyNkV5K2RzcDR3VVA0eXd6QTBUM0h2V2NlWE9GengzRnVxU2hNdnh5TlpoOVFDMzUKUUtncmxvaFF3SldBcmpONEh5MDZYcE93bGw2Z0RGZmNIK3ZScStFcnVuWXZMU0ZGR0NJWStVZTJjY0ZVVmZjSgo2VDRpb3NvYUx0cFJZNTh4VmpvSER0dmI2OHkzMU4zWFpDR000aEJPZ2R3PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
      service:
        name: {{ template "fullname" . }}-svc
        namespace: {{ .Release.Namespace }}
        path: /mutate-batch-tutorial-kubebuilder-io-v1-cronjob
    failurePolicy: Fail
    name: {{ template "fullname" . }}-svc.{{ .Release.Namespace }}.svc
    rules:
      - apiGroups:
          - batch.tutorial.kubebuilder.io
        apiVersions:
          - v1
        operations:
          - CREATE
          - UPDATE
        resources:
          - cronjobs
```
最后运行 build.sh开始使用helm chart安装controller和chart.
k get pods -n eskort  #查看controller是否安装成功
ka /Users/wukaiying/go/src/wukaiying/kubebuilderconjob/config/samples/batch_v1_cronjob.yaml
k get cronjobs.batch.tutorial.kubebuilder.io

### 其他：
webhook 路径/validate-batch-tutorial-kubebuilder-io-v1-cronjob 是有规则的
func generateMutatePath(gvk schema.GroupVersionKind) string {
	return "/mutate-" + strings.Replace(gvk.Group, ".", "-", -1) + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}

func generateValidatePath(gvk schema.GroupVersionKind) string {
	return "/validate-" + strings.Replace(gvk.Group, ".", "-", -1) + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}

make  run 会修改webhook文件件下的内容

### 参考
https://segmentfault.com/a/1190000020338350
https://book.kubebuilder.io/cronjob-tutorial/running-webhook.html