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

import (
	resource "k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type UpfDeploy struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec              UpfDeploySpec `json:"spec" yaml:"spec"`
}

type UpfDeploySpec struct {
	Capacity     UpfCapacity        `json:"capacity,omitempty" yaml:"capacity,omitempty"`
	BGPConfigs   []BGPConfig        `json:"BgpConfig,omitempty" yaml:"BgpConfig,omitempty"`
	N3Interfaces []NetworkInterface `json:"N3Interfaces,omitempty" yaml:"N3Interfaces,omitempty"`
	N4Interfaces []NetworkInterface `json:"N4Interfaces,omitempty" yaml:"N4Interfaces,omitempty"`
	N6Interfaces []NetworkInterface `json:"N6Interfaces,omitempty" yaml:"N6Interfaces,omitempty"`
	N9Interfaces []NetworkInterface `json:"N9Interfaces,omitempty" yaml:"N9Interfaces,omitempty"`
	VendorRef    *ObjectReference   `json:"vendorRef,omitempty" yaml:"vendorRef,omitempty"`
}

type UpfCapacity struct {
	UplinkThroughput   resource.Quantity `json:"uplinkThroughput" yaml:"uplinkThroughput"`
	DownlinkThroughput resource.Quantity `json:"downlinkThroughput" yaml:"downlinkThroughput"`
	MaximumConnections int               `json:"maximumConnections" yaml:"maximumConnections"`
}

type BGPConfig struct {
	VRName             string         `json:"virtualRouterName,omitempty" yaml:"virtualRouterName,omitempty"`
	VRNumber           int            `json:"virtualRouterNumber,omitempty" yaml:"virtualRouterNumber,omitempty"`
	RouteID            string         `json:"routeId,omitempty" yaml:"routeId,omitempty"`
	ASNumber           int            `json:"asNumber,omitempty" yaml:"asNumber,omitempty"`
	PeerASNumber       int            `json:"peerAsNumber,omitempty" yaml:"peerAsNumber,omitempty"`
	RouteDistinguisher string         `json:"routeDistinguisher,omitempty" yaml:"routeDistinguisher,omitempty"`
	Interfaces         []BGPInterface `json:"interfaces,omitempty" yaml:"interfaces,omitempty"`
}

type BGPInterface struct {
	Name       string `json:"interfaceName,omitempty" yaml:"interfaceName,omitempty"`
	NeighborIP string `json:"neighborIp,omitempty" yaml:"neighborIp,omitempty"`
}

type NetworkInterface struct {
	InterfaceName string   `json:"interfaceName" yaml:"interfaceName"`
	Latency       string   `json:"latency,omitempty" yaml:"latency,omitempty"`
	Bandwidth     string   `json:"bandwidth,omitempty" yaml:"bandwidth,omitempty"`
	IPAddr        []string `json:"ipAddr" yaml:"ipAddr"`
	Vlan          []string `json:"vlan" yaml:"vlan"`
}
