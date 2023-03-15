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

package crdreader

import (
	"errors"
	"os"
	"path/filepath"

	types "github.com/nephio-project/common-lib/nfdeploy"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type CRDSet struct {
	upfTypes            map[string]types.UPFType
	smfTypes            map[string]types.SMFType
	upfCapacityProfiles map[string]types.UPFCapacityProfile
	smfCapacityProfiles map[string]types.SMFCapacityProfile
	// TODO: Add amf crds once they are finalised
}

type ObjectType string

var _ CRDReader = &CRDSet{}

const (
	UPFTypeObject            ObjectType = "UpfType"
	SMFTypeObject            ObjectType = "SmfType"
	UPFCapacityProfileObject ObjectType = "UpfCapacityProfile"
	SMFCapacityProfileObject ObjectType = "SmfCapacityProfile"
)

// filterYamlsByKind: This method filters out yaml RNodes containing a specific
// kind from the list of yaml RNodes provided
func filterYamlsByKind(kind ObjectType, yamlNodes []*yaml.RNode) []*yaml.RNode {
	s := &framework.Selector{
		Kinds: []string{string(kind)},
	}
	filteredYamlNodes, _ := s.Filter(yamlNodes)
	return filteredYamlNodes
}

// init: This method initialises the crdSet
func (crdSet *CRDSet) init() {
	crdSet.upfTypes = make(map[string]types.UPFType)
	crdSet.smfTypes = make(map[string]types.SMFType)
	crdSet.upfCapacityProfiles = make(map[string]types.UPFCapacityProfile)
	crdSet.smfCapacityProfiles = make(map[string]types.SMFCapacityProfile)
}

// ReadCRDFiles : This method reads all the yaml files from the given directory
// and stores them in an in-memory map
func (crdSet *CRDSet) ReadCRDFiles(directory string) error {
	crdSet.init()
	var yamlNodes []*yaml.RNode
	err := filepath.Walk(
		directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fileInfo, err := os.Stat(path)
			if err != nil {
				return err
			}
			if !info.IsDir() && !fileInfo.IsDir() {
				yamlRNode, err := yaml.ReadFile(path)
				if err != nil {
					return err
				}
				yamlNodes = append(yamlNodes, yamlRNode)
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	filteredNodes := filterYamlsByKind(UPFTypeObject, yamlNodes)
	for _, node := range filteredNodes {
		upfType := &types.UPFType{}
		node.YNode().Decode(upfType)
		crdSet.upfTypes[node.GetName()] = *upfType
	}
	filteredNodes = filterYamlsByKind(SMFTypeObject, yamlNodes)
	for _, node := range filteredNodes {
		smfType := &types.SMFType{}
		node.YNode().Decode(smfType)
		crdSet.smfTypes[node.GetName()] = *smfType
	}
	filteredNodes = filterYamlsByKind(UPFCapacityProfileObject, yamlNodes)
	for _, node := range filteredNodes {
		upfCapacityProfile := &types.UPFCapacityProfile{}
		node.YNode().Decode(upfCapacityProfile)
		crdSet.upfCapacityProfiles[node.GetName()] = *upfCapacityProfile
	}
	filteredNodes = filterYamlsByKind(SMFCapacityProfileObject, yamlNodes)
	for _, node := range filteredNodes {
		smfCapacityProfile := &types.SMFCapacityProfile{}
		node.YNode().Decode(smfCapacityProfile)
		crdSet.smfCapacityProfiles[node.GetName()] = *smfCapacityProfile
	}
	return nil
}

// GetUPFTypeObject : This method returns a UpfType object based on its metadata name
func (crdSet *CRDSet) GetUPFTypeObject(crdName string) (types.UPFType, error) {
	if value, ok := crdSet.upfTypes[crdName]; ok {
		return value, nil
	}
	return types.UPFType{}, errors.New(
		"UPFType CRD with name " + crdName +
			" not found",
	)
}

// GetSMFTypeObject : This method returns an SmfType object based on its metadata name
func (crdSet *CRDSet) GetSMFTypeObject(crdName string) (types.SMFType, error) {
	if value, ok := crdSet.smfTypes[crdName]; ok {
		return value, nil
	}
	return types.SMFType{}, errors.New(
		"SMFType CRD with name " + crdName +
			" not found",
	)
}

// GetUPFCapacityProfileObject : This method returns a UPFCapacityProfile
// object based on its metadata name
func (crdSet *CRDSet) GetUPFCapacityProfileObject(crdName string) (
	types.UPFCapacityProfile, error,
) {
	if value, ok := crdSet.upfCapacityProfiles[crdName]; ok {
		return value, nil
	}
	return types.UPFCapacityProfile{}, errors.New(
		"UPFCapacityProfile CRD with" +
			" name " + crdName + " not found",
	)
}

// GetSMFCapacityProfileObject : This method returns an SMFCapacityProfile object
// based on its metadata name
func (crdSet *CRDSet) GetSMFCapacityProfileObject(crdName string) (
	types.SMFCapacityProfile, error,
) {
	if value, ok := crdSet.smfCapacityProfiles[crdName]; ok {
		return value, nil
	}
	return types.SMFCapacityProfile{}, errors.New(
		"SMFCapacityProfile CRD with name " +
			crdName + " not found",
	)
}
