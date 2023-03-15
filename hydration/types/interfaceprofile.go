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

type InterfaceProfile struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec              InterfaceProfileSpec `json:"spec" yaml:"spec"`
}

type InterfaceProfileSpec struct {
	ProfileName string `json:"profileName,omitempty" yaml:"profileName,omitempty"`
	Latency     string `json:"latency,omitempty" yaml:"latency,omitempty"`
	Bandwidth   string `json:"bandwidth,omitempty" yaml:"bandwidth,omitempty"`
}
