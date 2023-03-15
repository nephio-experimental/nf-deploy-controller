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
	types "github.com/nephio-project/common-lib/nfdeploy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	sampleUPFTypeName            = "sample-upf-type"
	sampleSMFTypeName            = "sample-smf-type"
	sampleUPFCapacityProfileName = "sample-upf-capacity-profile"
	sampleSMFCapacityProfileName = "sample-smf-capacity-profile"
)

func createFakeCRDSet() CRDSet {
	var crdSet = CRDSet{
		upfTypes:            map[string]types.UPFType{sampleUPFTypeName: {}},
		smfTypes:            map[string]types.SMFType{sampleSMFTypeName: {}},
		upfCapacityProfiles: map[string]types.UPFCapacityProfile{sampleUPFCapacityProfileName: {}},
		smfCapacityProfiles: map[string]types.SMFCapacityProfile{sampleSMFCapacityProfileName: {}},
	}
	return crdSet
}

var _ = Describe(
	"ReadCRDFiles", func() {
		var crdSet CRDSet
		BeforeEach(
			func() {
				crdSet.init()
			},
		)
		Context(
			"When correct directory is provided", func() {
				It(
					"Should return no error", func() {
						err := crdSet.ReadCRDFiles("testfiles/kyaml-readable-files")
						Expect(err).To(Not(HaveOccurred()))
						Expect(crdSet.upfTypes).To(Not(BeEmpty()))
						Expect(crdSet.smfTypes).To(Not(BeEmpty()))
						Expect(crdSet.smfCapacityProfiles).To(Not(BeEmpty()))
						Expect(crdSet.upfCapacityProfiles).To(Not(BeEmpty()))
					},
				)
			},
		)

		Context(
			"When incorrect directory is provided", func() {
				It(
					"Should return error", func() {
						err := crdSet.ReadCRDFiles("wrongdirectory")
						Expect(err).To(HaveOccurred())
						Expect(crdSet.upfTypes).To(BeEmpty())
						Expect(crdSet.smfTypes).To(BeEmpty())
						Expect(crdSet.smfCapacityProfiles).To(BeEmpty())
						Expect(crdSet.upfCapacityProfiles).To(BeEmpty())
					},
				)
			},
		)

		Context(
			"When directory having non KRM resources is provided", func() {
				It(
					"Should return error", func() {
						err := crdSet.ReadCRDFiles("testfiles/kyaml-unreadable-files")
						Expect(err).To(HaveOccurred())
						Expect(crdSet.upfTypes).To(BeEmpty())
						Expect(crdSet.smfTypes).To(BeEmpty())
						Expect(crdSet.smfCapacityProfiles).To(BeEmpty())
						Expect(crdSet.upfCapacityProfiles).To(BeEmpty())
					},
				)
			},
		)
	},
)

var _ = Describe(
	"GetUPFTypeObject", func() {
		var crdSet = createFakeCRDSet()

		Context(
			"When read UPFType Object is requested", func() {
				It(
					"Should return the requested struct", func() {
						obj, err := crdSet.GetUPFTypeObject(sampleUPFTypeName)
						Expect(err).To(Not(HaveOccurred()))
						Expect(obj).To(Equal(types.UPFType{}))
					},
				)
			},
		)
		Context(
			"When unread UPFType Object is requested", func() {
				It(
					"Should return error", func() {
						_, err := crdSet.GetUPFTypeObject("random-name")
						Expect(err).To(HaveOccurred())
					},
				)
			},
		)
	},
)

var _ = Describe(
	"GetSMFTypeObject", func() {
		var crdSet = createFakeCRDSet()

		Context(
			"When read SMFType Object is requested", func() {
				It(
					"Should return the requested struct", func() {
						obj, err := crdSet.GetSMFTypeObject(sampleSMFTypeName)
						Expect(err).To(Not(HaveOccurred()))
						Expect(obj).To(Equal(types.SMFType{}))
					},
				)
			},
		)
		Context(
			"When unread SMFType Object is requested", func() {
				It(
					"Should return error", func() {
						_, err := crdSet.GetSMFTypeObject("random-name")
						Expect(err).To(HaveOccurred())
					},
				)
			},
		)
	},
)

var _ = Describe(
	"GetUPFCapacityProfileObject", func() {
		var crdSet = createFakeCRDSet()

		Context(
			"When read UPFCapacityProfile Object is requested", func() {
				It(
					"Should return the requested struct", func() {
						obj, err := crdSet.GetUPFCapacityProfileObject(sampleUPFCapacityProfileName)
						Expect(err).To(Not(HaveOccurred()))
						Expect(obj).To(Equal(types.UPFCapacityProfile{}))
					},
				)
			},
		)
		Context(
			"When unread UPFCapacityProfile Object is requested", func() {
				It(
					"Should return error", func() {
						_, err := crdSet.GetUPFCapacityProfileObject("random-name")
						Expect(err).To(HaveOccurred())
					},
				)
			},
		)
	},
)

var _ = Describe(
	"GetSMFCapacityProfileObject", func() {
		var crdSet = createFakeCRDSet()

		Context(
			"When read SMFCapacityProfile Object is requested", func() {
				It(
					"Should return the requested struct", func() {
						obj, err := crdSet.GetSMFCapacityProfileObject(sampleSMFCapacityProfileName)
						Expect(err).To(Not(HaveOccurred()))
						Expect(obj).To(Equal(types.SMFCapacityProfile{}))
					},
				)
			},
		)
		Context(
			"When unread SMFCapacityProfile Object is requested", func() {
				It(
					"Should return error", func() {
						_, err := crdSet.GetSMFCapacityProfileObject("random-name")
						Expect(err).To(HaveOccurred())
					},
				)
			},
		)
	},
)
