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

package nftypehydration_test

import (
	"context"
	"errors"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	ctrl "sigs.k8s.io/controller-runtime"

	deployv1alpha1 "github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	"github.com/nephio-project/nf-deploy-controller/hydration/nftypehydration"
	hydrationutil "github.com/nephio-project/nf-deploy-controller/hydration/utils"
	ps "github.com/nephio-project/nf-deploy-controller/packageservice"
	mps "github.com/nephio-project/nf-deploy-controller/packageservice/mock"
	nfdeployutil "github.com/nephio-project/nf-deploy-controller/util"
)

const (
	smfNfDeployName = "nfDeploy1"
	smfClusterName  = "cluster1"
)

var (
	smfTypeSmall, smfcp, smfDeploy1                          []byte
	smfIP41, smfIP71, smfIP101, smfIP111                     []byte
	smfInterfaceConfig1, smfInterfaceConfig2, smfNfbgpconfig []byte
	smfNC                                                    nfdeployutil.NamingContext
)

func expectSmfType(mpsi *mps.MockPackageServiceInterface) {
	mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
		{
			ID:         1,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "SmfType",
			Name:       "smfsmall",
		},
	}), gomock.Eq(smfNC)).Return(map[int][]string{
		1: {string(smfTypeSmall)},
	}, nil).Times(1)
}

func expectSmfReferencedProfiles(mpsi *mps.MockPackageServiceInterface) {
	mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
		{
			ID:         1,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "NfBgpConfig",
		},
		{
			ID:         2,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "InterfaceConfig",
		},
	}), gomock.Eq(smfNC)).Return(map[int][]string{
		1: {string(smfNfbgpconfig)},
		2: {string(smfInterfaceConfig1), string(smfInterfaceConfig2)},
	}, nil).Times(1)
}

func expectSmfCapacityProfile(mpsi *mps.MockPackageServiceInterface) {
	mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
		{
			ID:         1,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "SmfCapacityProfile",
			Name:       "smfCapacityProfile1",
		},
	}), gomock.Eq(smfNC)).Return(map[int][]string{
		1: {string(smfcp)},
	}, nil).Times(1)
}

func expectSmfInterfaceProfile(mpsi *mps.MockPackageServiceInterface) {
	mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
		{
			ID:         0,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "InterfaceProfile",
			Name:       "profile41",
		},
		{
			ID:         1,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "InterfaceProfile",
			Name:       "profile71",
		},
		{
			ID:         2,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "InterfaceProfile",
			Name:       "profile101",
		},
		{
			ID:         3,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "InterfaceProfile",
			Name:       "profile111",
		},
	}), gomock.Eq(smfNC)).Return(map[int][]string{
		0: {string(smfIP41)},
		1: {string(smfIP71)},
		2: {string(smfIP101)},
		3: {string(smfIP111)},
	}, nil).Times(1)
}

var _ = Describe("Smfdeploy", func() {
	var (
		mockCtrl *gomock.Controller
		mpsi     *mps.MockPackageServiceInterface
		sdi      *nftypehydration.SmfDeployImpl
	)
	ctx := context.Background()

	smfTypeSmall, _ = os.ReadFile("../testhelper/smftype_small.yaml")
	smfNfbgpconfig, _ = os.ReadFile("../testhelper/nfbgpconfig.yaml")
	smfInterfaceConfig1, _ = os.ReadFile("../testhelper/interfaceconfig1.yaml")
	smfInterfaceConfig2, _ = os.ReadFile("../testhelper/interfaceconfig2.yaml")
	smfcp, _ = os.ReadFile("../testhelper/smfcapacityprofile.yaml")
	smfDeploy1, _ = os.ReadFile("../testhelper/smfdeploy1.yaml")
	smfIP41, _ = os.ReadFile("../testhelper/interfaceprofile41.yaml")
	smfIP71, _ = os.ReadFile("../testhelper/interfaceprofile71.yaml")
	smfIP101, _ = os.ReadFile("../testhelper/interfaceprofile101.yaml")
	smfIP111, _ = os.ReadFile("../testhelper/interfaceprofile111.yaml")
	smfNC, _ = nfdeployutil.NewNamingContext(smfClusterName, smfNfDeployName)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mpsi = mps.NewMockPackageServiceInterface(mockCtrl)
		sdi = &nftypehydration.SmfDeployImpl{
			PS:  mpsi,
			Log: ctrl.Log.WithName("smfdeploy"),
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Testing GenerateNfTypeDeploy for single smf site", func() {
		site := deployv1alpha1.Site{
			Id:          "smf1",
			ClusterName: smfClusterName,
			NFType:      "smf",
			NFTypeName:  "smfsmall",
			NFVendor:    "casa",
			NFVersion:   "1.0",
		}
		Context("testing smfdeploy with small smftype", func() {
			BeforeEach(func() {
				expectSmfType(mpsi)
				expectSmfReferencedProfiles(mpsi)
				expectSmfCapacityProfile(mpsi)
				expectSmfInterfaceProfile(mpsi)
			})
			It("should process a single smf and return smfdeploy", func() {
				format.MaxLength = 0
				resp, err := sdi.GenerateNfTypeDeploy(ctx, site, smfNfDeployName)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).To(Equal(smfDeploy1))
			})
		})
		Context("expecting error from packageservice for GetNFProfiles", func() {
			expectedErr := errors.New("error from packageservice")
			It("should return an error while getting interfaceProfile from packageservice", func() {
				expectSmfType(mpsi)
				expectSmfReferencedProfiles(mpsi)
				expectSmfCapacityProfile(mpsi)

				mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
					{
						ID:         0,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "InterfaceProfile",
						Name:       "profile41",
					},
					{
						ID:         1,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "InterfaceProfile",
						Name:       "profile71",
					},
					{
						ID:         2,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "InterfaceProfile",
						Name:       "profile101",
					},
					{
						ID:         3,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "InterfaceProfile",
						Name:       "profile111",
					},
				}), gomock.Eq(smfNC)).Return(nil, expectedErr).Times(1)

				resp, err := sdi.GenerateNfTypeDeploy(ctx, site, smfNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
			})

			It("should return an error while getting smfCapacityProfile from packageservice", func() {
				expectSmfType(mpsi)
				expectSmfReferencedProfiles(mpsi)

				mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
					{
						ID:         1,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "SmfCapacityProfile",
						Name:       "smfCapacityProfile1",
					},
				}), gomock.Eq(smfNC)).Return(nil, expectedErr).Times(1)

				resp, err := sdi.GenerateNfTypeDeploy(ctx, site, smfNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
			})

			It("should return an error while getting referencedProfiles from packageservice", func() {
				expectSmfType(mpsi)

				mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
					{
						ID:         1,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "NfBgpConfig",
					},
					{
						ID:         2,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "InterfaceConfig",
					},
				}), gomock.Eq(smfNC)).Return(nil, expectedErr).Times(1)

				resp, err := sdi.GenerateNfTypeDeploy(ctx, site, smfNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
			})

			It("should return an error while getting smfType from packageservice", func() {
				mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
					{
						ID:         1,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "SmfType",
						Name:       "smfsmall",
					},
				}), gomock.Eq(smfNC)).Return(nil, expectedErr).Times(1)

				resp, err := sdi.GenerateNfTypeDeploy(ctx, site, smfNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
			})
		})
	})
})
