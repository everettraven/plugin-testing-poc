package kubernetes

import (
	"encoding/json"
	"strings"
)

type VersionInfo interface {
	Major() string
	Minor() string
	GitVersion() string
}

type KubernetesVersion interface {
	ClientVersion() VersionInfo
	ServerVersion() VersionInfo
}

// TODO: Add a default here
type kubeVersionInfoJson struct {
	Major      string `json:"major"`
	Minor      string `json:"minor"`
	GitVersion string `json:"gitVersion"`
}

type KubeVersionInfo struct {
	kubeVersionInfoJson
}

func NewKubeVersionInfo(out string) (*KubeVersionInfo, error) {
	kvi := &KubeVersionInfo{}
	dec := json.NewDecoder(strings.NewReader(out))
	if err := dec.Decode(&kvi.kubeVersionInfoJson); err != nil {
		return nil, err
	}

	return kvi, nil
}

func (kvi *KubeVersionInfo) Major() string {
	return kvi.kubeVersionInfoJson.Major
}

func (kvi *KubeVersionInfo) Minor() string {
	return kvi.kubeVersionInfoJson.Minor
}

func (kvi *KubeVersionInfo) GitVersion() string {
	return kvi.kubeVersionInfoJson.GitVersion
}

type KubeVersion struct {
	clientVersion KubeVersionInfo `json:"clientVersion,omitempty"`
	serverVersion KubeVersionInfo `json:"serverVersion,omitempty"`
}

type KubeVersionOption func(kv *KubeVersion)

func WithClientVersion(clientVersion VersionInfo) KubeVersionOption {
	return func(kv *KubeVersion) {
		kv.clientVersion = KubeVersionInfo{
			kubeVersionInfoJson: kubeVersionInfoJson{
				Major:      clientVersion.Major(),
				Minor:      clientVersion.Minor(),
				GitVersion: clientVersion.GitVersion(),
			},
		}
	}
}

func WithServerVersion(serverVersion VersionInfo) KubeVersionOption {
	return func(kv *KubeVersion) {
		kv.serverVersion = KubeVersionInfo{
			kubeVersionInfoJson: kubeVersionInfoJson{
				Major:      serverVersion.Major(),
				Minor:      serverVersion.Minor(),
				GitVersion: serverVersion.GitVersion(),
			},
		}
	}
}

func NewKubeVersion(opts ...KubeVersionOption) *KubeVersion {
	kv := &KubeVersion{}

	for _, opt := range opts {
		opt(kv)
	}

	return kv
}

func (kv *KubeVersion) ClientVersion() VersionInfo {
	return &kv.clientVersion
}

func (kv *KubeVersion) ServerVersion() VersionInfo {
	return &kv.serverVersion
}
