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

type SmfType struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec              SmfTypeSpec `json:"spec" yaml:"spec"`
}

type SmfTypeSpec struct {
	Name                string                   `json:"name" yaml:"name"`
	CapacityProfile     NFTypeCapacityProfile    `json:"capacityProfile,omitempty" yaml:"capacityProfile,omitempty"`
	N4InterfaceProfile  []NFTypeInterfaceProfile `json:"N4InterfaceProfile,omitempty" yaml:"N4InterfaceProfile,omitempty"`
	N7InterfaceProfile  []NFTypeInterfaceProfile `json:"N7InterfaceProfile,omitempty" yaml:"N7InterfaceProfile,omitempty"`
	N10InterfaceProfile []NFTypeInterfaceProfile `json:"N10InterfaceProfile,omitempty" yaml:"N10InterfaceProfile,omitempty"`
	N11InterfaceProfile []NFTypeInterfaceProfile `json:"N11InterfaceProfile,omitempty" yaml:"N11InterfaceProfile,omitempty"`
}
