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

type UpfType struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec              UpfTypeSpec `json:"spec" yaml:"spec"`
}

type UpfTypeSpec struct {
	Name               string                   `json:"name" yaml:"name"`
	UpfCapacityProfile NFTypeCapacityProfile    `json:"UpfCapacityProfile,omitempty" yaml:"UpfCapacityProfile,omitempty"`
	N3InterfaceProfile []NFTypeInterfaceProfile `json:"N3InterfaceProfile,omitempty" yaml:"N3InterfaceProfile,omitempty"`
	N4InterfaceProfile []NFTypeInterfaceProfile `json:"N4InterfaceProfile,omitempty" yaml:"N4InterfaceProfile,omitempty"`
	N6InterfaceProfile []NFTypeInterfaceProfile `json:"N6InterfaceProfile,omitempty" yaml:"N6InterfaceProfile,omitempty"`
	N9InterfaceProfile []NFTypeInterfaceProfile `json:"N9InterfaceProfile,omitempty" yaml:"N9InterfaceProfile,omitempty"`
}

type NFTypeCapacityProfile struct {
	ProfileName string `json:"profileName,omitempty" yaml:"profileName,omitempty"`
}

type NFTypeInterfaceProfile struct {
	InterfaceProfileName string `json:"intfProfileName,omitempty" yaml:"intfProfileName,omitempty"`
	ID                   int    `json:"id,omitempty" yaml:"id,omitempty"`
}
