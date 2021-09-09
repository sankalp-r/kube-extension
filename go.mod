module github.com/sankalp-r/extension-operator

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/gofrs/flock v0.8.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/pkg/errors v0.9.1
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.6.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.9.6
)
