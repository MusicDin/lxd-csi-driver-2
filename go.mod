module github.com/canonical/lxd-csi-driver

go 1.24.5

replace github.com/canonical/lxd => ./lxd

require (
	github.com/canonical/lxd v0.0.0-20250818084043-b690439a8cfa
	github.com/container-storage-interface/spec v1.11.0
	google.golang.org/grpc v1.74.2
	google.golang.org/protobuf v1.36.7
	k8s.io/klog/v2 v2.130.1
	k8s.io/mount-utils v0.33.4
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397
)

require (
	github.com/flosch/pongo2 v0.0.0-20200913210552-0d938eb266f3 // indirect
	github.com/go-jose/go-jose/v4 v4.1.2 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/muhlemmer/gu v0.3.1 // indirect
	github.com/pkg/sftp v1.13.9 // indirect
	github.com/pkg/xattr v0.4.12 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/zitadel/logging v0.6.2 // indirect
	github.com/zitadel/oidc/v3 v3.44.0 // indirect
	github.com/zitadel/schema v1.3.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250811230008-5f3141c8851a // indirect
)
