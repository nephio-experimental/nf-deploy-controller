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
	"sigs.k8s.io/kustomize/kyaml/yaml"

	ausftypes "github.com/nephio-project/common-lib/ausf"
	udmtypes "github.com/nephio-project/common-lib/udm"
)

type UpfCapacityProfile struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	UpfCPSpec         `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type UpfCPSpec struct {
	UpfCapacity `json:",inline" yaml:",inline"`
}

type SmfCapacityProfile struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	SmfCPSpec         `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type SmfCPSpec struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	MaxSessions int    `json:"maxSessions,omitempty" yaml:"maxSessions,omitempty"`
}

type AusfCapacityProfile struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec              ausftypes.CapacityProfile `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type UdmCapacityProfile struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec              udmtypes.CapacityProfile `json:"spec,omitempty" yaml:"spec,omitempty"`
}
