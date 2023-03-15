/*
Copyright 2022-2023 The Nephio Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import "sigs.k8s.io/kustomize/kyaml/yaml"

type SmfDeploy struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec              SmfDeploySpec `json:"spec" yaml:"spec"`
}

type SmfDeploySpec struct {
	Capacity      SmfCapacity        `json:"capacity,omitempty" yaml:"capacity,omitempty"`
	BGPConfigs    []BGPConfig        `json:"BgpConfig,omitempty" yaml:"BgpConfig,omitempty"`
	N4Interfaces  []NetworkInterface `json:"N4Interfaces,omitempty" yaml:"N4Interfaces,omitempty"`
	N7Interfaces  []NetworkInterface `json:"N7Interfaces,omitempty" yaml:"N7Interfaces,omitempty"`
	N10Interfaces []NetworkInterface `json:"N10Interfaces,omitempty" yaml:"N10Interfaces,omitempty"`
	N11Interfaces []NetworkInterface `json:"N11Interfaces,omitempty" yaml:"N11Interfaces,omitempty"`
}

type SmfCapacity struct {
	MaxSubscriber int `json:"maxSubscriber,omitempty" yaml:"maxSubscriber,omitempty"`
	AvgSubscriber int `json:"avgSubscriber,omitempty" yaml:"avgSubscriber,omitempty"`
	MaxSession    int `json:"maxSession,omitempty" yaml:"maxSession,omitempty"`
}
