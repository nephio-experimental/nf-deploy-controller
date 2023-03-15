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
	ctrl "sigs.k8s.io/controller-runtime"

	deployv1alpha1 "github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	"github.com/nephio-project/nf-deploy-controller/hydration/nftypehydration"
	hydrationutil "github.com/nephio-project/nf-deploy-controller/hydration/utils"
	ps "github.com/nephio-project/nf-deploy-controller/packageservice"
	mps "github.com/nephio-project/nf-deploy-controller/packageservice/mock"
	nfdeployutil "github.com/nephio-project/nf-deploy-controller/util"
)

const (
	upfNfDeployName = "nfDeploy1"
	upfClusterName  = "cluster1"
)

var (
	upfTypeSmall, upfcp, upfDeploy1, upfDeploy1WithExtn      []byte
	upfInterfaceConfig1, upfInterfaceConfig2, upfNfbgpconfig []byte
	upfExtension                                             []byte
	upfNC                                                    nfdeployutil.NamingContext
)

func expectUpfType(mpsi *mps.MockPackageServiceInterface) {
	mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
		{
			ID:         1,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "UpfType",
			Name:       "upfsmall",
		},
	}), gomock.Eq(upfNC)).Return(map[int][]string{
		1: {string(upfTypeSmall)},
	}, nil).Times(1)
}

func expectUpfReferencedProfiles(mpsi *mps.MockPackageServiceInterface) {
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
	}), gomock.Eq(upfNC)).Return(map[int][]string{
		1: {string(upfNfbgpconfig)},
		2: {string(upfInterfaceConfig1), string(upfInterfaceConfig2)},
	}, nil).Times(1)
}

func expectUpfCapacityProfile(mpsi *mps.MockPackageServiceInterface) {
	mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
		{
			ID:         1,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "UpfCapacityProfile",
			Name:       "upfCapacityProfile1",
		},
	}), gomock.Eq(upfNC)).Return(map[int][]string{
		1: {string(upfcp)},
	}, nil).Times(1)
}

var _ = Describe("Upfdeploy", func() {
	var (
		mockCtrl *gomock.Controller
		mpsi     *mps.MockPackageServiceInterface
		udi      *nftypehydration.UpfDeployImpl
	)
	ctx := context.Background()

	upfTypeSmall, _ = os.ReadFile("../testhelper/upftype_small.yaml")
	upfNfbgpconfig, _ = os.ReadFile("../testhelper/nfbgpconfig.yaml")
	upfInterfaceConfig1, _ = os.ReadFile("../testhelper/interfaceconfig1.yaml")
	upfInterfaceConfig2, _ = os.ReadFile("../testhelper/interfaceconfig2.yaml")
	upfcp, _ = os.ReadFile("../testhelper/upfcapacityprofile.yaml")
	upfDeploy1, _ = os.ReadFile("../testhelper/upfdeploy1.yaml")
	upfDeploy1WithExtn, _ = os.ReadFile("../testhelper/upfdeploy1withextn.yaml")
	upfExtension, _ = os.ReadFile("../testhelper/upfextension.yaml")
	upfNC, _ = nfdeployutil.NewNamingContext(upfClusterName, upfNfDeployName)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mpsi = mps.NewMockPackageServiceInterface(mockCtrl)
		udi = &nftypehydration.UpfDeployImpl{
			PS:  mpsi,
			Log: ctrl.Log.WithName("upfdeploy"),
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Testing GenerateNfTypeDeploy for single upf site", func() {
		site := deployv1alpha1.Site{
			Id:          "upf1",
			ClusterName: upfClusterName,
			NFType:      "upf",
			NFTypeName:  "upfsmall",
			NFVendor:    "casa",
			NFVersion:   "1.0",
		}
		Context("testing upfdeploy with small upftype", func() {
			BeforeEach(func() {
				expectUpfType(mpsi)
				expectUpfReferencedProfiles(mpsi)
				expectUpfCapacityProfile(mpsi)
			})
			It("should process a single upf and return upfdeploy without extension", func() {
				mpsi.EXPECT().GetVendorExtensionPackage(ctx, upfNC, gomock.Eq(ps.VendorNFKey{
					Vendor: site.NFVendor, Version: site.NFVersion, NFType: site.NFType,
				})).Times(1).Return([]string{}, nil)
				resp, err := udi.GenerateNfTypeDeploy(ctx, site, upfNfDeployName)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).To(Equal(upfDeploy1))
			})
			It("should process a single upf and return upfdeploy with extension", func() {
				mpsi.EXPECT().GetVendorExtensionPackage(ctx, upfNC, gomock.Eq(ps.VendorNFKey{
					Vendor: site.NFVendor, Version: site.NFVersion, NFType: site.NFType,
				})).Times(1).Return([]string{string(upfExtension)}, nil)
				resp, err := udi.GenerateNfTypeDeploy(ctx, site, upfNfDeployName)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).To(Equal(upfDeploy1WithExtn))
			})
		})

		Context("error scenarios for vendor extension package", func() {
			BeforeEach(func() {
				expectUpfType(mpsi)
				expectUpfReferencedProfiles(mpsi)
				expectUpfCapacityProfile(mpsi)
			})
			It("should return error when error getting from packageService", func() {
				cause := errors.New("error getting extension package")
				mpsi.EXPECT().GetVendorExtensionPackage(ctx, upfNC, gomock.Eq(ps.VendorNFKey{
					Vendor: site.NFVendor, Version: site.NFVersion, NFType: site.NFType,
				})).Times(1).Return(nil, cause)
				resp, err := udi.GenerateNfTypeDeploy(ctx, site, upfNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(cause))
				Expect(resp).To(BeNil())
			})
			It("should return error when package service returned multiple k8s objects", func() {
				mpsi.EXPECT().GetVendorExtensionPackage(ctx, upfNC, gomock.Eq(ps.VendorNFKey{
					Vendor: site.NFVendor, Version: site.NFVersion, NFType: site.NFType,
				})).Times(1).Return([]string{string(upfExtension), string(upfExtension)}, nil)
				resp, err := udi.GenerateNfTypeDeploy(ctx, site, upfNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("More than one extension object found"))
				Expect(resp).To(BeNil())
			})
		})

		Context("expecting error from packageservice for GetNFProfiles", func() {
			expectedErr := errors.New("error from packageservice")
			It("should return an error while getting upfCapacityProfile from packageservice", func() {
				expectUpfType(mpsi)
				expectUpfReferencedProfiles(mpsi)

				mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
					{
						ID:         1,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "UpfCapacityProfile",
						Name:       "upfCapacityProfile1",
					},
				}), gomock.Eq(upfNC)).Return(nil, expectedErr).Times(1)

				resp, err := udi.GenerateNfTypeDeploy(ctx, site, upfNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
			})

			It("should return an error while getting referencedProfiles from packageservice", func() {
				expectUpfType(mpsi)

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
				}), gomock.Eq(upfNC)).Return(nil, expectedErr).Times(1)

				resp, err := udi.GenerateNfTypeDeploy(ctx, site, upfNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
			})

			It("should return an error while getting upfType from packageservice", func() {
				mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
					{
						ID:         1,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "UpfType",
						Name:       "upfsmall",
					},
				}), gomock.Eq(upfNC)).Return(nil, expectedErr).Times(1)

				resp, err := udi.GenerateNfTypeDeploy(ctx, site, upfNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
			})
		})
	})
})
