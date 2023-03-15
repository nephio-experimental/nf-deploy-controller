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

package hydration_test

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	deployv1alpha1 "github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	"github.com/nephio-project/nf-deploy-controller/hydration"
	hydrationutil "github.com/nephio-project/nf-deploy-controller/hydration/utils"
	ps "github.com/nephio-project/nf-deploy-controller/packageservice"
	mps "github.com/nephio-project/nf-deploy-controller/packageservice/mock"
	nfdeployutil "github.com/nephio-project/nf-deploy-controller/util"
)

const (
	nfDeployName       = "nfDeploy1"
	clusterName        = "cluster1"
	expectedFileFormat = "%s-%s.yaml"
)

var (
	upfTypeSmall, upfcp, upfDeploy1    []byte
	smfTypeSmall, smfcp, smfDeploy1    []byte
	ausfcp, ausfDeploy1                []byte
	udmcp, udmDeploy1                  []byte
	interfaceConfig1, interfaceConfig2 []byte
	nfbgpconfig                        []byte
	ip41, ip71, ip101, ip111           []byte
	nc                                 nfdeployutil.NamingContext
)

func getSite(id, nfType, nfTypeName string) deployv1alpha1.Site {
	return deployv1alpha1.Site{
		Id:          id,
		ClusterName: clusterName,
		NFType:      nfType,
		NFTypeName:  nfTypeName,
		NFVendor:    "casa",
		NFVersion:   "1.0",
	}
}

func getNfDeployForSites(sites []deployv1alpha1.Site) deployv1alpha1.NfDeploy {
	return deployv1alpha1.NfDeploy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: deployv1alpha1.GroupVersion.String(),
			Kind:       "NfDeploy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nfDeployName,
			Namespace: "default",
		},
		Spec: deployv1alpha1.NfDeploySpec{
			Sites: sites,
		},
	}
}

func expectUpfType(mpsi *mps.MockPackageServiceInterface) {
	mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
		{
			ID:         1,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "UpfType",
			Name:       "upfsmall",
		},
	}), gomock.Eq(nc)).Return(map[int][]string{
		1: {string(upfTypeSmall)},
	}, nil).Times(1)
}

func expectReferencedProfiles(mpsi *mps.MockPackageServiceInterface) {
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
	}), gomock.Eq(nc)).Return(map[int][]string{
		1: {string(nfbgpconfig)},
		2: {string(interfaceConfig1), string(interfaceConfig2)},
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
	}), gomock.Eq(nc)).Return(map[int][]string{
		1: {string(upfcp)},
	}, nil).Times(1)
}

func expectSmfType(mpsi *mps.MockPackageServiceInterface) {
	mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
		{
			ID:         1,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "SmfType",
			Name:       "smfsmall",
		},
	}), gomock.Eq(nc)).Return(map[int][]string{
		1: {string(smfTypeSmall)},
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
	}), gomock.Eq(nc)).Return(map[int][]string{
		1: {string(smfcp)},
	}, nil).Times(1)
}

func expectInterfaceProfile(mpsi *mps.MockPackageServiceInterface) {
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
	}), gomock.Eq(nc)).Return(map[int][]string{
		0: {string(ip41)},
		1: {string(ip71)},
		2: {string(ip101)},
		3: {string(ip111)},
	}, nil).Times(1)
}

func expectAusfCapacityProfile(mpsi *mps.MockPackageServiceInterface) {
	mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
		{
			ID:         1,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "AusfCapacityProfile",
		},
	}), gomock.Eq(nc)).Return(map[int][]string{
		1: {string(ausfcp)},
	}, nil).Times(1)
}

func expectUdmCapacityProfile(mpsi *mps.MockPackageServiceInterface) {
	mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
		{
			ID:         1,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "UdmCapacityProfile",
		},
	}), gomock.Eq(nc)).Return(map[int][]string{
		1: {string(udmcp)},
	}, nil).Times(1)
}

func expectGetVendorExtnPkg(mpsi *mps.MockPackageServiceInterface,
	site deployv1alpha1.Site,
	extnPkg []string) {
	mpsi.EXPECT().GetVendorExtensionPackage(gomock.Any(), gomock.Any(), gomock.Eq(ps.VendorNFKey{
		Vendor:  site.NFVendor,
		Version: site.NFVersion,
		NFType:  site.NFType,
	})).Times(1).Return(extnPkg, nil)
}

var _ = Describe("Hydration", func() {
	var (
		mockCtrl *gomock.Controller
		mpsi     *mps.MockPackageServiceInterface
		h        *hydration.Hydration
	)
	ctx := context.Background()

	upfTypeSmall, _ = os.ReadFile("testhelper/upftype_small.yaml")
	smfTypeSmall, _ = os.ReadFile("testhelper/smftype_small.yaml")
	interfaceConfig1, _ = os.ReadFile("testhelper/interfaceconfig1.yaml")
	interfaceConfig2, _ = os.ReadFile("testhelper/interfaceconfig2.yaml")
	nfbgpconfig, _ = os.ReadFile("testhelper/nfbgpconfig.yaml")
	upfcp, _ = os.ReadFile("testhelper/upfcapacityprofile.yaml")
	smfcp, _ = os.ReadFile("testhelper/smfcapacityprofile.yaml")
	ausfcp, _ = os.ReadFile("testhelper/ausfcapacityprofile.yaml")
	udmcp, _ = os.ReadFile("testhelper/udmcapacityprofile.yaml")
	upfDeploy1, _ = os.ReadFile("testhelper/upfdeploy1.yaml")
	smfDeploy1, _ = os.ReadFile("testhelper/smfdeploy1.yaml")
	ausfDeploy1, _ = os.ReadFile("testhelper/ausfdeploy1.yaml")
	udmDeploy1, _ = os.ReadFile("testhelper/udmdeploy1.yaml")
	ip41, _ = os.ReadFile("testhelper/interfaceprofile41.yaml")
	ip71, _ = os.ReadFile("testhelper/interfaceprofile71.yaml")
	ip101, _ = os.ReadFile("testhelper/interfaceprofile101.yaml")
	ip111, _ = os.ReadFile("testhelper/interfaceprofile111.yaml")
	nc, _ = nfdeployutil.NewNamingContext(clusterName, nfDeployName)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mpsi = mps.NewMockPackageServiceInterface(mockCtrl)
		h = &hydration.Hydration{
			PS:  mpsi,
			Log: ctrl.Log.WithName("Hydration"),
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Testing NfDeploy Hydration for single upf site", func() {
		nfDeploy := getNfDeployForSites([]deployv1alpha1.Site{
			getSite("upf1", "upf", "upfsmall"),
		})
		Context("testing upfdeploy with small upftype", func() {
			BeforeEach(func() {
				expectUpfType(mpsi)
				expectReferencedProfiles(mpsi)
				expectUpfCapacityProfile(mpsi)
				expectGetVendorExtnPkg(mpsi, nfDeploy.Spec.Sites[0], []string{})
			})
			It("should process a single upf and return upfdeploy", func() {
				mpsi.EXPECT().CreateDeployPackage(gomock.Any(), gomock.Eq(map[string]string{
					fmt.Sprintf(expectedFileFormat, nfDeployName, "upf1"): string(upfDeploy1),
				}), gomock.Eq(nc)).Return("resourceName", nil).Times(1)
				n, err := h.Hydrate(ctx, nfDeploy)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(n)).To(Equal(1))
				Expect(n[0]).To(Equal("resourceName"))
			})
			It("should process a single upf and return an error while creating Deploy Package", func() {
				expectedErr := errors.New("error from porch")
				mpsi.EXPECT().CreateDeployPackage(gomock.Any(), gomock.Eq(map[string]string{
					fmt.Sprintf(expectedFileFormat, nfDeployName, "upf1"): string(upfDeploy1),
				}), gomock.Eq(nc)).Return("", expectedErr).Times(1)
				n, err := h.Hydrate(ctx, nfDeploy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
				Expect(n).To(BeNil())
			})
		})
		Context("expecting error from porch for GetNFProfiles", func() {
			It("should return an error", func() {
				mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
					{
						ID:         1,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "UpfType",
						Name:       "upfsmall",
					},
				}), gomock.Eq(nc)).Return(nil, errors.New("error from porch")).Times(1)
				n, err := h.Hydrate(ctx, nfDeploy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error hydrating sites: [upf1]"))
				Expect(n).To(BeNil())
			})
		})
	})

	Describe("Testing NfDeploy Hydration for single smf site", func() {
		nfDeploy := getNfDeployForSites([]deployv1alpha1.Site{
			getSite("smf1", "smf", "smfsmall"),
		})
		Context("testing smfdeploy with small smftype", func() {
			BeforeEach(func() {
				expectSmfType(mpsi)
				expectReferencedProfiles(mpsi)
				expectSmfCapacityProfile(mpsi)
				expectInterfaceProfile(mpsi)
			})
			It("should process a single smf and return smfdeploy", func() {
				mpsi.EXPECT().CreateDeployPackage(gomock.Any(), gomock.Eq(map[string]string{
					fmt.Sprintf(expectedFileFormat, nfDeployName, "smf1"): string(smfDeploy1),
				}), gomock.Eq(nc)).Return("resourceName", nil).Times(1)
				n, err := h.Hydrate(ctx, nfDeploy)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(n)).To(Equal(1))
				Expect(n[0]).To(Equal("resourceName"))
			})
			It("should process a single smf and return an error while creating Deploy Package", func() {
				expectedErr := errors.New("error from porch")
				mpsi.EXPECT().CreateDeployPackage(gomock.Any(), gomock.Eq(map[string]string{
					fmt.Sprintf(expectedFileFormat, nfDeployName, "smf1"): string(smfDeploy1),
				}), gomock.Eq(nc)).Return("", expectedErr).Times(1)
				n, err := h.Hydrate(ctx, nfDeploy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
				Expect(n).To(BeNil())
			})
		})
		Context("expecting error from porch for GetNFProfiles", func() {
			It("should return an error", func() {
				mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
					{
						ID:         1,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "SmfType",
						Name:       "smfsmall",
					},
				}), gomock.Eq(nc)).Return(nil, errors.New("error from porch")).Times(1)
				n, err := h.Hydrate(ctx, nfDeploy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error hydrating sites: [smf1]"))
				Expect(n).To(BeNil())
			})
		})
	})

	Describe("Testing NfDeploy Hydration for invalide NfType", func() {
		Context("expecting error for invalid NfType", func() {
			nfDeploy := getNfDeployForSites([]deployv1alpha1.Site{
				getSite("invalid1", "invalid", "invalidTypeName"),
			})
			It("should return an error for invalid NfType", func() {
				expectedErr := errors.New("error hydrating sites: [invalid1]")
				n, err := h.Hydrate(ctx, nfDeploy)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(expectedErr))
				Expect(n).To(BeNil())
			})
		})
	})

	Describe("Testing NfDeploy Hydration for upf and smf sites together", func() {
		nfDeploy := getNfDeployForSites([]deployv1alpha1.Site{
			getSite("upf1", "upf", "upfsmall"),
			getSite("smf1", "smf", "smfsmall"),
		})
		Context("testing hydration with small upftype and small smftype", func() {
			BeforeEach(func() {
				expectUpfType(mpsi)
				expectReferencedProfiles(mpsi)
				expectUpfCapacityProfile(mpsi)
				expectGetVendorExtnPkg(mpsi, nfDeploy.Spec.Sites[0], []string{})

				expectSmfType(mpsi)
				expectReferencedProfiles(mpsi)
				expectSmfCapacityProfile(mpsi)
				expectInterfaceProfile(mpsi)
			})
			It("should process upf and smf and return both upfdeploy and smfdeploy", func() {
				mpsi.EXPECT().CreateDeployPackage(gomock.Any(), gomock.Eq(map[string]string{
					fmt.Sprintf(expectedFileFormat, nfDeployName, "upf1"): string(upfDeploy1),
					fmt.Sprintf(expectedFileFormat, nfDeployName, "smf1"): string(smfDeploy1),
				}), gomock.Eq(nc)).Return("resourceName", nil).Times(1)
				n, err := h.Hydrate(ctx, nfDeploy)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(n)).To(Equal(1))
				Expect(n[0]).To(Equal("resourceName"))
			})
		})
	})

	Describe("Testing NfDeploy Hydration for single ausf site", func() {
		nfDeploy := getNfDeployForSites([]deployv1alpha1.Site{
			getSite("ausf1", "ausf", "ausfsmall"),
		})
		Context("testing ausfdeploy", func() {
			BeforeEach(func() {
				expectAusfCapacityProfile(mpsi)
			})
			It("should process a single ausf and return ausfdeploy", func() {
				mpsi.EXPECT().CreateDeployPackage(gomock.Any(), gomock.Eq(map[string]string{
					fmt.Sprintf(expectedFileFormat, nfDeployName, "ausf1"): string(ausfDeploy1),
				}), gomock.Eq(nc)).Return("resourceName", nil).Times(1)
				n, err := h.Hydrate(ctx, nfDeploy)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(n)).To(Equal(1))
				Expect(n[0]).To(Equal("resourceName"))
			})
			It("should process a single ausf and return an error while creating Deploy Package", func() {
				expectedErr := errors.New("error from porch")
				mpsi.EXPECT().CreateDeployPackage(gomock.Any(), gomock.Eq(map[string]string{
					fmt.Sprintf(expectedFileFormat, nfDeployName, "ausf1"): string(ausfDeploy1),
				}), gomock.Eq(nc)).Return("", expectedErr).Times(1)
				n, err := h.Hydrate(ctx, nfDeploy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
				Expect(n).To(BeNil())
			})
		})
	})

	Describe("Testing NfDeploy Hydration for single udm site", func() {
		nfDeploy := getNfDeployForSites([]deployv1alpha1.Site{
			getSite("udm1", "udm", "udmsmall"),
		})
		Context("testing udmdeploy", func() {
			BeforeEach(func() {
				expectUdmCapacityProfile(mpsi)
			})
			It("should process a single udm and return udmdeploy", func() {
				mpsi.EXPECT().CreateDeployPackage(gomock.Any(), gomock.Eq(map[string]string{
					fmt.Sprintf(expectedFileFormat, nfDeployName, "udm1"): string(udmDeploy1),
				}), gomock.Eq(nc)).Return("resourceName", nil).Times(1)
				n, err := h.Hydrate(ctx, nfDeploy)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(n)).To(Equal(1))
				Expect(n[0]).To(Equal("resourceName"))
			})
			It("should process a single udm and return an error while creating Deploy Package", func() {
				expectedErr := errors.New("error from porch")
				mpsi.EXPECT().CreateDeployPackage(gomock.Any(), gomock.Eq(map[string]string{
					fmt.Sprintf(expectedFileFormat, nfDeployName, "udm1"): string(udmDeploy1),
				}), gomock.Eq(nc)).Return("", expectedErr).Times(1)
				n, err := h.Hydrate(ctx, nfDeploy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
				Expect(n).To(BeNil())
			})
		})
	})

	Describe("Testing CreateNFDeployActuators to validate the actuator package creation", func() {
		nfDeploy := getNfDeployForSites([]deployv1alpha1.Site{
			{ClusterName: "cluster1", NFVendor: "ABC", NFVersion: "1.0", NFType: "upf"},
			{ClusterName: "cluster1", NFVendor: "ABC", NFVersion: "1.0", NFType: "upf"},
			{ClusterName: "cluster1", NFVendor: "ABC", NFVersion: "1.0", NFType: "smf"},
			{ClusterName: "cluster1", NFVendor: "ABC", NFVersion: "2.0", NFType: "upf"},
			{ClusterName: "cluster1", NFVendor: "XYZ", NFVersion: "1.0", NFType: "upf"},
			{ClusterName: "cluster2", NFVendor: "ABC", NFVersion: "1.0", NFType: "upf"},
		})
		Context("Valid inputs with different combination of uniqueness and duplicates in sites", func() {
			It("Should call packageService to create actuators package for unique keys only", func() {
				nc1, _ := nfdeployutil.NewNamingContext("cluster1", nfDeploy.Name)
				nc2, _ := nfdeployutil.NewNamingContext("cluster2", nfDeploy.Name)
				expectedCluster1VendorNFs := []ps.VendorNFKey{
					{Vendor: "ABC", Version: "1.0", NFType: "upf"},
					{Vendor: "ABC", Version: "1.0", NFType: "smf"},
					{Vendor: "ABC", Version: "2.0", NFType: "upf"},
					{Vendor: "XYZ", Version: "1.0", NFType: "upf"},
				}
				expectedBoolReturn := []bool{true, true, false, false}
				for index, expA := range expectedCluster1VendorNFs {
					mpsi.EXPECT().
						CreateNFDeployActuators(ctx, nc1, expA).
						Return(fmt.Sprintf("package%d", index+1), expectedBoolReturn[index], nil).
						Times(1)
				}
				mpsi.EXPECT().
					CreateNFDeployActuators(ctx, nc2, ps.VendorNFKey{Vendor: "ABC", Version: "1.0", NFType: "upf"}).
					Return("package5", true, nil).
					Times(1)
				expectedPkgNames := []string{"package1", "package2", "package5"}
				pkgNames, err := h.CreateNFDeployActuators(ctx, nfDeploy)
				Expect(err).NotTo(HaveOccurred())
				Expect(pkgNames).To(ConsistOf(expectedPkgNames))
			})

			It("Should return empty list and nil error when sites is empty", func() {
				nfdeploy := getNfDeployForSites([]deployv1alpha1.Site{})
				mpsi.EXPECT().
					CreateNFDeployActuators(ctx, gomock.Any(), gomock.Any()).
					Times(0)
				pkgNames, err := h.CreateNFDeployActuators(ctx, nfdeploy)
				Expect(err).NotTo(HaveOccurred())
				Expect(pkgNames).NotTo(BeNil())
				Expect(len(pkgNames)).To(Equal(0))
			})
		})

		Context("Invalid inputs to test error scenarios", func() {
			It("Should return error when error calling package service", func() {
				cause := errors.New("Error fetching pkg revision")
				mpsi.EXPECT().
					CreateNFDeployActuators(ctx, gomock.Any(), gomock.Any()).
					Return("", false, cause).
					Times(1)
				pkgNames, err := h.CreateNFDeployActuators(ctx, nfDeploy)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(cause))
				Expect(pkgNames).To(BeNil())
			})

			It("Should return error when error calling package service", func() {
				nfdeploy := getNfDeployForSites([]deployv1alpha1.Site{
					{ClusterName: "", NFVendor: "ABC", NFVersion: "1.0", NFType: "upf"},
				})
				mpsi.EXPECT().
					CreateNFDeployActuators(ctx, gomock.Any(), gomock.Any()).
					Times(0)
				pkgNames, err := h.CreateNFDeployActuators(ctx, nfdeploy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error creating naming context"))
				Expect(pkgNames).To(BeNil())
			})
		})
	})
})
