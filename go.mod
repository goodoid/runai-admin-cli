module github.com/kubeflow/arena

go 1.13

require (
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/frankban/quicktest v1.9.0 // indirect
	github.com/go-openapi/spec v0.19.3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/snappy v0.0.1 // indirect
	github.com/kubeflow/mpi-operator v0.2.2 // indirect
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.2.2
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/ulikunitz/xz v0.5.7 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/cli-runtime v0.17.4
	k8s.io/client-go v0.17.4
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	k8s.io/kubectl v0.17.4
)

replace vbom.ml/util => github.com/fvbommel/util v0.0.0-20160121211510-db5cfe13f5cc
