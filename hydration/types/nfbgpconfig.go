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

type NFBGPConfig struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec              NFBGPSpec `json:"spec" yaml:"spec"`
}

type NFBGPSpec struct {
	Name     string          `json:"name" yaml:"name"`
	NFName   string          `json:"nfName" yaml:"nfName"`
	VRouters []VirtualRouter `json:"virtualRouters" yaml:"virtualRouters"`
}

type VirtualRouter struct {
	Name               string           `json:"name" yaml:"name"`
	Number             int              `json:"number" yaml:"number"`
	RouteID            string           `json:"routeId" yaml:"routeId"`
	ASNumber           int              `json:"asNumber" yaml:"asNumber"`
	RouteDistinguisher string           `json:"routeDistinguisher" yaml:"routeDistinguisher"`
	GroupName          string           `json:"groupName" yaml:"groupName"`
	PeerASNumber       int              `json:"peerAsNumber" yaml:"peerAsNumber"`
	Interfaces         []NFBGPInterface `json:"interfaces" yaml:"interfaces"`
}

type NFBGPInterface struct {
	Name       string `json:"name" yaml:"name"`
	NeighborIP string `json:"neighborIp,omitempty" yaml:"neighborIp,omitempty"`
}
