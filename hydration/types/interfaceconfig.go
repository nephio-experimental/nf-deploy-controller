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

type InterfaceConfig struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec              []InterfaceCfgSpec `json:"spec" yaml:"spec"`
}

type InterfaceCfgSpec struct {
	Name           string   `json:"name" yaml:"name"`
	NFName         string   `json:"nfName" yaml:"nfName"`
	InterfaceGroup string   `json:"interfaceGroup" yaml:"interfaceGroup"`
	ID             int      `json:"id" yaml:"id"`
	IPAddr         []string `json:"ipAddr" yaml:"ipAddr"`
	Vlan           []string `json:"vlan" yaml:"vlan"`
}
