module github.com/Luet-lab/extensions/extensions/repo-devkit

go 1.15

replace github.com/docker/docker => github.com/Luet-lab/moby v17.12.0-ce-rc1.0.20200605210607-749178b8f80d+incompatible

replace github.com/containerd/containerd => github.com/containerd/containerd v1.3.1-0.20200227195959-4d242818bf55

replace github.com/hashicorp/go-immutable-radix => github.com/tonistiigi/go-immutable-radix v0.0.0-20170803185627-826af9ccf0fe

replace github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305

replace github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc9.0.20200221051241-688cf6d43cc4

require (
	github.com/Luet-lab/luet-portage-converter v0.4.2-0.20210811064616-ed4133e4bdd6
	github.com/MottainaiCI/mottainai-server v0.0.2-0.20210531211337-27f12a56ea5f
	github.com/geaaru/time-master v0.3.1
	github.com/jaypipes/ghw v0.6.1 // indirect
	github.com/minio/minio-go/v7 v7.0.10
	github.com/mitchellh/hashstructure/v2 v2.0.1 // indirect
	github.com/mudler/luet v0.0.0-20210604142351-a7b4ae67c9b8
	github.com/rickb777/date v1.13.0 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/stevenle/topsort v0.0.0-20130922064739-8130c1d7596b // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	gopkg.in/yaml.v2 v2.4.0
	mvdan.cc/sh/v3 v3.0.0-beta1 // indirect
)
