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

package deployment

import (
	"github.com/golang/mock/gomock"
	"github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	crdreader "github.com/nephio-project/nf-deploy-controller/crd-reader"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	types2 "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	sampleUPFName    = "sample-upf"
	sampleSMFName    = "sample-smf"
	sampleAMFName    = "sample-amf"
	sampleNFTypeName = "sample-nf-type"
)

func createSampleDeployment() *Deployment {
	var crdReader crdreader.CRDReader
	d := Deployment{
		name:      "sample",
		crdReader: crdReader,
		upfNodes: map[string]UPFNode{
			sampleUPFName: UPFNode{
				Node: Node{
					Id: sampleUPFName, NFType: UPF,
					Connections: map[string]void{sampleSMFName: present},
				},
				Spec: UPFSpec{},
			},
		},
		smfNodes: map[string]SMFNode{
			sampleSMFName: SMFNode{
				Node: Node{
					Id: sampleSMFName, NFType: SMF, Connections: map[string]void{
						sampleUPFName: present, sampleAMFName: present,
					},
				},
				Spec: SMFSpec{},
			},
		},
		amfNodes: map[string]AMFNode{
			sampleAMFName: AMFNode{
				Node{
					Id: sampleAMFName, NFType: AMF,
					Connections: map[string]void{sampleSMFName: present},
				},
			},
		},
		edges: []Edge{
			Edge{FirstNode: sampleUPFName, SecondNode: sampleSMFName},
			Edge{FirstNode: sampleSMFName, SecondNode: sampleAMFName},
		},
		logger: zap.New(
			func(options *zap.Options) {
				options.Development = true
				options.DestWriter = GinkgoWriter
			},
		),
	}
	return &d
}

var _ = Describe(
	"Init", func() {
		var deployment Deployment

		Context(
			"When deployment is provided", func() {
				It(
					"Should initialize all required components", func() {
						var crdReader crdreader.CRDReader
						deployment.Init(
							crdReader, &crdreader.UPFIntent{}, &crdreader.SMFIntent{}, nil,
							nil, types2.NamespacedName{}, zap.New(
								func(options *zap.Options) {
									options.Development = true
									options.DestWriter = GinkgoWriter
								},
							),
						)
						Expect(deployment.upfNodes).NotTo(Equal(nil))
						Expect(deployment.smfNodes).NotTo(Equal(nil))
						Expect(deployment.amfNodes).NotTo(Equal(nil))
					},
				)
			},
		)
	},
)

var _ = Describe(
	"getNFType", func() {
		var deployment Deployment
		BeforeEach(
			func() {
				deployment = *createSampleDeployment()
			},
		)
		Context(
			"When Upf is provided", func() {
				It(
					"Should return NFType as UPF", func() {
						nfKind := deployment.getNFType(sampleUPFName)
						Expect(nfKind).To(Equal(UPF))
					},
				)
			},
		)
		Context(
			"When Amf is provided", func() {
				It(
					"Should return NFType as AMF", func() {
						nfKind := deployment.getNFType(sampleAMFName)
						Expect(nfKind).To(Equal(AMF))
					},
				)
			},
		)
		Context(
			"When Smf is provided", func() {
				It(
					"Should return NFType as SMF", func() {
						nfKind := deployment.getNFType(sampleSMFName)
						Expect(nfKind).To(Equal(SMF))
					},
				)
			},
		)
		Context(
			"When Unspecified NF id provided", func() {
				It(
					"Should return UnspecifiedNF", func() {
						NFKind := deployment.getNFType("random-name")
						Expect(NFKind).To(Equal(UnspecifiedNFType))
					},
				)
			},
		)
	},
)

var _ = Describe(
	"addOrUpdateUPFNode", func() {
		var deployment = createSampleDeployment()
		var expectedThroughput = "2000"
		var siteName = "upf-test"
		site := v1alpha1.Site{
			NFType:     "upf",
			NFTypeName: sampleNFTypeName,
			Id:         siteName,
		}
		BeforeEach(
			func() {
				ctrl := gomock.NewController(GinkgoT())
				mockUPFIntentProcessor := crdreader.NewMockUPFIntentProcessor(ctrl)
				mockUPFIntentProcessor.EXPECT().GetUPFIntent(
					sampleNFTypeName, deployment.crdReader,
				).
					Return(crdreader.UPFIntent{Throughput: expectedThroughput}, nil)
				deployment.upfIntentProcessor = mockUPFIntentProcessor
			},
		)
		Context(
			"When a new upf site is provided", func() {
				It(
					"Should create a upf node and update the node's spec", func() {
						initialUPFCount := len(deployment.upfNodes)
						deployment.addOrUpdateUPFNode(site)
						Expect(len(deployment.upfNodes)).To(Equal(initialUPFCount + 1))
						Expect(deployment.upfNodes[siteName].Spec.throughput).To(Equal(expectedThroughput))
					},
				)
			},
		)
		Context(
			"When an existing upf site is provided", func() {
				It(
					"Should update the node's spec", func() {
						deployment.upfNodes[site.Id] = UPFNode{
							Node: Node{
								Id: site.Id, NFType: UPF,
							},
							Spec: UPFSpec{},
						}
						deployment.addOrUpdateUPFNode(site)
						Expect(deployment.upfNodes[siteName].Spec.throughput).To(Equal(expectedThroughput))
					},
				)
			},
		)
	},
)

var _ = Describe(
	"addOrUpdateSMFNode", func() {
		var deployment = createSampleDeployment()
		var expectedSessions = "2000"
		var siteName = "smf-test"
		site := v1alpha1.Site{
			NFType:     "smf",
			NFTypeName: sampleNFTypeName,
			Id:         siteName,
		}
		BeforeEach(
			func() {
				ctrl := gomock.NewController(GinkgoT())
				mockSMFIntentProcessor := crdreader.NewMockSMFIntentProcessor(ctrl)
				mockSMFIntentProcessor.EXPECT().GetSMFIntent(
					sampleNFTypeName, deployment.crdReader,
				).
					Return(crdreader.SMFIntent{MaxSessions: expectedSessions}, nil)
				deployment.smfIntentProcessor = mockSMFIntentProcessor
			},
		)
		Context(
			"When a new smf site is provided", func() {
				It(
					"Should create an smf node and update the node's spec", func() {
						initialSMFCount := len(deployment.smfNodes)
						deployment.addOrUpdateSMFNode(site)
						Expect(len(deployment.smfNodes)).To(Equal(initialSMFCount + 1))
						Expect(deployment.smfNodes[siteName].Spec.maxSessions).To(Equal(expectedSessions))
					},
				)
			},
		)
		Context(
			"When an existing smf site is provided", func() {
				It(
					"Should update the node's spec", func() {
						deployment.smfNodes[site.Id] = SMFNode{
							Node: Node{
								Id: site.Id, NFType: SMF,
							},
							Spec: SMFSpec{},
						}
						deployment.addOrUpdateSMFNode(site)
						Expect(deployment.smfNodes[siteName].Spec.maxSessions).To(Equal(expectedSessions))
					},
				)
			},
		)
	},
)

var _ = Describe(
	"addOrUpdateAMFNode", func() {
		var deployment = createSampleDeployment()
		var siteName = "amf-test"
		site := v1alpha1.Site{

			NFType:     "amf",
			NFTypeName: sampleNFTypeName,
			Id:         siteName,
		}
		Context(
			"When a new amf site is provided", func() {
				It(
					"Should create an amf node", func() {
						initialAMFCount := len(deployment.amfNodes)
						deployment.addOrUpdateAMFNode(site)
						Expect(len(deployment.amfNodes)).To(Equal(initialAMFCount + 1))
					},
				)
			},
		)
	},
)

var _ = Describe(
	"addConnection", func() {
		var deployment Deployment
		BeforeEach(
			func() {
				deployment = *createSampleDeployment()
			},
		)
		Context(
			"When UPF Node Connection is requested", func() {
				It(
					"Should add the connection", func() {
						initialUPFConnections := len(deployment.upfNodes[sampleUPFName].Connections)
						initialSMFConnections := len(deployment.smfNodes[sampleSMFName].Connections)
						initialAMFConnections := len(deployment.amfNodes[sampleAMFName].Connections)
						deployment.addConnection(sampleUPFName, UPF, "smf-test")
						Expect(len(deployment.upfNodes[sampleUPFName].Connections)).To(Equal(initialUPFConnections + 1))
						Expect(len(deployment.smfNodes[sampleSMFName].Connections)).To(Equal(initialSMFConnections))
						Expect(len(deployment.amfNodes[sampleAMFName].Connections)).To(Equal(initialAMFConnections))
					},
				)
			},
		)
		Context(
			"When SMF Node Connection is requested", func() {
				It(
					"Should add the connection", func() {
						initialUPFConnections := len(deployment.upfNodes[sampleUPFName].Connections)
						initialSMFConnections := len(deployment.smfNodes[sampleSMFName].Connections)
						initialAMFConnections := len(deployment.amfNodes[sampleAMFName].Connections)
						deployment.addConnection(sampleSMFName, SMF, "upf-test")
						Expect(len(deployment.upfNodes[sampleUPFName].Connections)).To(Equal(initialUPFConnections))
						Expect(len(deployment.smfNodes[sampleSMFName].Connections)).To(Equal(initialSMFConnections + 1))
						Expect(len(deployment.amfNodes[sampleAMFName].Connections)).To(Equal(initialAMFConnections))
					},
				)
			},
		)
		Context(
			"When AMF Node Connection is requested", func() {
				It(
					"Should add the connection", func() {
						initialUPFConnections := len(deployment.upfNodes[sampleUPFName].Connections)
						initialSMFConnections := len(deployment.smfNodes[sampleSMFName].Connections)
						initialAMFConnections := len(deployment.amfNodes[sampleAMFName].Connections)
						deployment.addConnection(sampleAMFName, AMF, "smf-test")
						Expect(len(deployment.upfNodes[sampleUPFName].Connections)).To(Equal(initialUPFConnections))
						Expect(len(deployment.smfNodes[sampleSMFName].Connections)).To(Equal(initialSMFConnections))
						Expect(len(deployment.amfNodes[sampleAMFName].Connections)).To(Equal(initialAMFConnections + 1))
					},
				)
			},
		)
	},
)

var _ = Describe(
	"createEdge", func() {
		var deployment Deployment
		BeforeEach(
			func() {
				deployment = *createSampleDeployment()
			},
		)
		Context(
			"When Edge Connection is present", func() {
				It(
					"Should not add the edge", func() {
						initialEdgeSize := len(deployment.edges)
						deployment.createEdge(sampleAMFName, sampleSMFName)
						Expect(len(deployment.edges)).To(Equal(initialEdgeSize))
					},
				)
			},
		)
		Context(
			"When Edge Connection is not present", func() {
				It(
					"Should add the edge", func() {
						initialGraphSize := len(deployment.edges)
						deployment.createEdge(sampleAMFName, "test-smf")
						Expect(len(deployment.edges)).To(Equal(initialGraphSize + 1))
					},
				)
			},
		)
	},
)

var _ = Describe(
	"removeEdge", func() {
		Context(
			"If given edge is present", func() {
				It(
					"should remove the edge", func() {
						deployment := *createSampleDeployment()
						initialEdgeCount := len(deployment.edges)
						deployment.removeEdge(sampleUPFName, sampleSMFName)
						Expect(len(deployment.edges)).To(Equal(initialEdgeCount - 1))
					},
				)
			},
		)
	},
)

var _ = Describe(
	"removeNFs", func() {
		Context(
			"If any site is not present", func() {
				It(
					"should remove the site and its connections", func() {
						deployment := createSampleDeployment()
						testFile, _ := yaml.ReadFile("testfiles/nfdeploy.yaml")
						nfdeploy := &v1alpha1.NfDeploy{}
						testFile.YNode().Decode(nfdeploy)
						deployment.removeNFs(*nfdeploy)
						Expect(len(deployment.edges)).To(Equal(0))
						Expect(len(deployment.upfNodes)).To(Equal(0))
						Expect(len(deployment.smfNodes)).To(Equal(0))
						Expect(len(deployment.amfNodes)).To(Equal(0))

					},
				)
			},
		)
	},
)

var _ = Describe(
	"ReportNFDeployEvent", func() {
		var deployment = Deployment{}
		var crdReader crdreader.CRDReader
		logger := zap.New(
			func(options *zap.Options) {
				options.Development = true
				options.DestWriter = GinkgoWriter
			},
		)
		deployment.Init(
			crdReader, &crdreader.UPFIntent{}, &crdreader.SMFIntent{}, nil,
			nil, types2.NamespacedName{}, logger,
		)
		nfDeploy := &v1alpha1.NfDeploy{}
		nfDeploy2 := &v1alpha1.NfDeploy{}
		BeforeEach(
			func() {
				ctrl := gomock.NewController(GinkgoT())
				mockUPF := crdreader.NewMockUPFIntentProcessor(ctrl)
				testFile, _ := yaml.ReadFile("testfiles/nfdeploy.yaml")
				testFile2, _ := yaml.ReadFile("testfiles/nfdeploy-2.yaml")
				testFile2.YNode().Decode(nfDeploy2)
				testFile.YNode().Decode(nfDeploy)
				mockUPF.EXPECT().GetUPFIntent(
					nfDeploy.Spec.Sites[0].NFTypeName,
					deployment.crdReader,
				).
					Return(crdreader.UPFIntent{}, nil).AnyTimes()
				deployment.upfIntentProcessor = mockUPF
				mockSMF := crdreader.NewMockSMFIntentProcessor(ctrl)
				mockSMF.EXPECT().GetSMFIntent(
					nfDeploy.Spec.Sites[1].NFTypeName,
					deployment.crdReader,
				).
					Return(crdreader.SMFIntent{}, nil).AnyTimes()
				deployment.smfIntentProcessor = mockSMF
			},
		)
		Context(
			"When correct NfDeploy is provided", func() {
				It(
					"Should create correct deployment", func() {
						var deploymentProcessor DeploymentProcessor = &deployment
						deploymentProcessor.ReportNFDeployEvent(*nfDeploy)
						Expect(len(deployment.upfNodes)).To(Equal(1))
						Expect(len(deployment.smfNodes)).To(Equal(1))
					},
				)
			},
		)
		Context(
			"When correct NfDeploy is provided and deployment is already created",
			func() {
				It(
					"Should update the deployment", func() {
						var deploymentProcessor DeploymentProcessor = &deployment
						deploymentProcessor.ReportNFDeployEvent(*nfDeploy)
						Expect(len(deployment.upfNodes)).To(Equal(1))
						Expect(len(deployment.smfNodes)).To(Equal(1))
						deploymentProcessor.ReportNFDeployEvent(*nfDeploy2)
						_, isPresent := deployment.upfNodes["upf-dummy-2"]
						Expect(isPresent).To(BeTrue())
						_, isPresent = deployment.smfNodes["smf-dummy-2"]
						Expect(isPresent).To(BeTrue())
						Expect(len(deployment.upfNodes)).To(Equal(1))
						Expect(len(deployment.smfNodes)).To(Equal(1))
					},
				)
			},
		)

	},
)
