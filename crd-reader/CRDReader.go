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

import types "github.com/nephio-project/common-lib/nfdeploy"

// CRDReader : CRDReader interface exposes CRD config files and optionally
// converts them to corresponding Objects.
type CRDReader interface {

	// ReadCRDFiles : This method reads all the yaml files from the given directory
	// and stores them in an in-memory map
	ReadCRDFiles(directory string) error

	// GetUPFTypeObject : This method returns a UpfType object based on its metadata name
	GetUPFTypeObject(crdName string) (types.UPFType, error)

	// GetSMFTypeObject : This method returns an SmfType object based on its metadata name
	GetSMFTypeObject(crdName string) (types.SMFType, error)

	// GetUPFCapacityProfileObject : This method returns a UPFCapacityProfile
	// object based on its metadata name
	GetUPFCapacityProfileObject(crdName string) (types.UPFCapacityProfile, error)

	// GetSMFCapacityProfileObject : This method returns an SMFCapacityProfile object
	// based on its metadata name
	GetSMFCapacityProfileObject(crdName string) (types.SMFCapacityProfile, error)
}
