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
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/nephio-project/nf-deploy-controller/util"

	types3 "github.com/nephio-project/common-lib/ausf"
	types2 "github.com/nephio-project/common-lib/udm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/go-logr/logr"
	types "github.com/nephio-project/common-lib/nfdeploy"
	"github.com/nephio-project/edge-watcher/preprocessor"
	"github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	crdreader "github.com/nephio-project/nf-deploy-controller/crd-reader"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Deployment : This fundamental module represents a single NFDeploy
// intent. It is responsible for constructing in-memory graph of NFDeploy and storing
// spec and status of individual NFs.
type Deployment struct {
	ctx       context.Context
	cancelCtx func()
	name      string
	// protects NF Nodes and Edges
	deploymentMu sync.RWMutex
	upfNodes     map[string]UPFNode
	smfNodes     map[string]SMFNode
	amfNodes     map[string]AMFNode
	ausfNodes    map[string]AUSFNode
	udmNodes     map[string]UDMNode
	edges        []Edge

	// TODO: add intent processing for other NFs once we start
	// using this intent in status flow
	crdReader          crdreader.CRDReader
	upfIntentProcessor crdreader.UPFIntentProcessor
	smfIntentProcessor crdreader.SMFIntentProcessor
	edgeErrorChan      chan error
	edgeEventsChan     chan preprocessor.Event

	statusReader   client.Reader
	statusWriter   client.StatusWriter
	namespacedName NamespacedName
	logger         logr.Logger
}

var _ DeploymentProcessor = &Deployment{}

// connected : Represents if two nodes are connected
var connected void

// present : Represents if a node is present in list
var present void

// Init : This method initialises the deployment
func (deployment *Deployment) Init(
	CRDReader crdreader.CRDReader,
	upfIntentProcessor crdreader.UPFIntentProcessor,
	smfIntentProcessor crdreader.SMFIntentProcessor,
	statusReader client.Reader,
	statusWriter client.StatusWriter,
	namespacedName NamespacedName,
	logger logr.Logger,
) {
	deployment.ctx = context.Background()
	deployment.ctx, deployment.cancelCtx = context.WithCancel(deployment.ctx)
	deployment.upfNodes = make(map[string]UPFNode)
	deployment.smfNodes = make(map[string]SMFNode)
	deployment.amfNodes = make(map[string]AMFNode)
	deployment.ausfNodes = make(map[string]AUSFNode)
	deployment.udmNodes = make(map[string]UDMNode)
	deployment.upfIntentProcessor = upfIntentProcessor
	deployment.smfIntentProcessor = smfIntentProcessor
	deployment.crdReader = CRDReader
	deployment.edgeErrorChan = make(chan error)
	deployment.edgeEventsChan = make(chan preprocessor.Event)
	deployment.statusWriter = statusWriter
	deployment.statusReader = statusReader
	deployment.namespacedName = namespacedName
	deployment.logger = logger
}

// GetNFType := Returns the Type of NF
func (deployment *Deployment) getNFType(nfId string) NFType {
	if _, isPresent := deployment.upfNodes[nfId]; isPresent {
		return UPF
	}
	if _, isPresent := deployment.smfNodes[nfId]; isPresent {
		return SMF
	}
	if _, isPresent := deployment.amfNodes[nfId]; isPresent {
		return AMF
	}
	if _, isPresent := deployment.ausfNodes[nfId]; isPresent {
		return AUSF
	}
	if _, isPresent := deployment.udmNodes[nfId]; isPresent {
		return UDM
	}
	return UnspecifiedNFType
}

func (deployment *Deployment) addOrUpdateUPFNode(site v1alpha1.Site) {
	nfTypeName := site.NFTypeName
	nfId := site.Id
	clusterName := site.ClusterName
	if _, isPresent := deployment.upfNodes[nfId]; !isPresent {
		node := Node{Id: nfId, NFType: UPF, Connections: make(map[string]void)}
		upfNode := UPFNode{Node: node, Spec: UPFSpec{clusterName: clusterName}}
		deployment.upfNodes[upfNode.Id] = upfNode
	}
	upfIntent, err := deployment.upfIntentProcessor.GetUPFIntent(
		nfTypeName, deployment.crdReader,
	)
	if err != nil {
		deployment.logger.Error(
			err, "CRD Reader failed to fetch upf intent", "UPFDeploy", nfId,
		)
		return
	}
	upfNode := deployment.upfNodes[nfId]
	upfNode.Spec.throughput = upfIntent.Throughput
	deployment.upfNodes[nfId] = upfNode
}

func (deployment *Deployment) addOrUpdateSMFNode(site v1alpha1.Site) {
	nfTypeName := site.NFTypeName
	nfId := site.Id
	clusterName := site.ClusterName
	if _, isPresent := deployment.smfNodes[nfId]; !isPresent {
		node := Node{Id: nfId, NFType: SMF, Connections: make(map[string]void)}
		smfNode := SMFNode{Node: node, Spec: SMFSpec{clusterName: clusterName}}
		deployment.smfNodes[smfNode.Id] = smfNode
	}
	smfNode := deployment.smfNodes[nfId]

	smfIntent, err := deployment.smfIntentProcessor.GetSMFIntent(
		nfTypeName, deployment.crdReader,
	)
	if err != nil {
		deployment.logger.Error(
			err, "CRD Reader failed to fetch smf intent", "SMFDeploy", nfId,
		)
		return
	}
	smfNode.Spec.maxSessions = smfIntent.MaxSessions
	deployment.smfNodes[nfId] = smfNode
}

func (deployment *Deployment) addOrUpdateAMFNode(site v1alpha1.Site) {
	nfId := site.Id
	if _, isPresent := deployment.amfNodes[nfId]; !isPresent {
		amfNode := AMFNode{
			Node{
				Id: nfId, NFType: AMF, Connections: make(map[string]void),
			},
		}
		// TODO : Update AMFNode once AMFDeploy is finalised
		deployment.amfNodes[amfNode.Id] = amfNode
	}
}

func (deployment *Deployment) addOrUpdateAUSFNode(site v1alpha1.Site) {
	nfId := site.Id
	if _, isPresent := deployment.ausfNodes[nfId]; !isPresent {
		node := Node{Id: nfId, NFType: AUSF, Connections: make(map[string]void)}
		ausfNode := AUSFNode{Node: node}
		deployment.ausfNodes[ausfNode.Id] = ausfNode
	}
}

func (deployment *Deployment) addOrUpdateUDMNode(site v1alpha1.Site) {
	nfId := site.Id
	if _, isPresent := deployment.udmNodes[nfId]; !isPresent {
		node := Node{Id: nfId, NFType: UDM, Connections: make(map[string]void)}
		udmNode := UDMNode{Node: node}
		deployment.udmNodes[udmNode.Id] = udmNode
	}
}

// AddConnection := Add nfId2 to nfId1's connection list
func (deployment *Deployment) addConnection(
	nfId1 string, nfType1 NFType, nfId2 string,
) {
	switch nfType1 {
	case UPF:
		deployment.upfNodes[nfId1].Connections[nfId2] = connected
	case SMF:
		deployment.smfNodes[nfId1].Connections[nfId2] = connected
	case AMF:
		deployment.amfNodes[nfId1].Connections[nfId2] = connected
	case UDM:
		deployment.udmNodes[nfId1].Connections[nfId2] = connected
	case AUSF:
		deployment.ausfNodes[nfId1].Connections[nfId2] = connected
	}
}

// CreateEdge := Creates edge connection between two nfs if the connection
// is not already present
func (deployment *Deployment) createEdge(nfId1 string, nfId2 string) {
	for _, edge := range deployment.edges {
		if edge.IsEqual(nfId2, nfId1) {
			return
		}
	}
	edge := Edge{FirstNode: nfId1, SecondNode: nfId2}
	nfType1 := deployment.getNFType(nfId1)
	nfType2 := deployment.getNFType(nfId2)
	deployment.addConnection(nfId1, nfType1, nfId2)
	deployment.addConnection(nfId2, nfType2, nfId1)
	deployment.edges = append(deployment.edges, edge)
}

// removeEdge: removes an edge connecting nfId1 and nfId2 from deployment
// no-ops if edge is not present in deployment
func (deployment *Deployment) removeEdge(nfId1 string, nfId2 string) {
	for index, edge := range deployment.edges {
		if edge.IsEqual(nfId1, nfId2) {
			deployment.edges = append(
				deployment.edges[:index], deployment.edges[index+1:]...,
			)
			return
		}
	}
}

// removeNFs: removes NFs that are not present in nfDeploy
func (deployment *Deployment) removeNFs(nfDeploy v1alpha1.NfDeploy) {
	var nfList = make(map[string]NFType)
	for _, site := range nfDeploy.Spec.Sites {
		nfList[site.Id] = NFType(site.NFType)
	}
	for key, upfNode := range deployment.upfNodes {
		if _, isPresent := nfList[key]; !isPresent || nfList[key] != UPF {
			for node := range upfNode.Connections {
				deployment.removeEdge(key, node)
			}
			delete(deployment.upfNodes, key)
		}
	}
	for key, smfNode := range deployment.smfNodes {
		if _, isPresent := nfList[key]; !isPresent || nfList[key] != SMF {
			for node := range smfNode.Connections {
				deployment.removeEdge(key, node)
			}
			delete(deployment.smfNodes, key)
		}
	}
	for key, amfNode := range deployment.amfNodes {
		if _, isPresent := nfList[key]; !isPresent || nfList[key] != AMF {
			for node := range amfNode.Connections {
				deployment.removeEdge(key, node)
			}
			delete(deployment.amfNodes, key)
		}
	}
	for key, udmNode := range deployment.udmNodes {
		if _, isPresent := nfList[key]; !isPresent || nfList[key] != UDM {
			for node := range udmNode.Connections {
				deployment.removeEdge(key, node)
			}
			delete(deployment.udmNodes, key)
		}
	}
	for key, ausfNode := range deployment.ausfNodes {
		if _, isPresent := nfList[key]; !isPresent || nfList[key] != AUSF {
			for node := range ausfNode.Connections {
				deployment.removeEdge(key, node)
			}
			delete(deployment.ausfNodes, key)
		}
	}
}

// ReportNFDeployEvent := Takes nfDeploy and creates & updates deployment graph structure.
// It also updates the spec of individual NFs
func (deployment *Deployment) ReportNFDeployEvent(nfDeploy v1alpha1.NfDeploy) {

	deployment.deploymentMu.Lock()
	defer deployment.deploymentMu.Unlock()
	deployment.name = nfDeploy.Name
	for _, site := range nfDeploy.Spec.Sites {
		switch NFType(site.NFType) {
		case UPF:
			deployment.addOrUpdateUPFNode(site)

		case SMF:
			deployment.addOrUpdateSMFNode(site)

		// TODO: Uncomment when AMFDeploy is finalised
		//case AMF:
		//	deployment.addOrUpdateAMFNode(site)
		case AUSF:
			deployment.addOrUpdateAUSFNode(site)
		case UDM:
			deployment.addOrUpdateUDMNode(site)
		}
	}
	for _, site := range nfDeploy.Spec.Sites {
		for _, connection := range site.Connectivities {
			deployment.createEdge(site.Id, connection.NeighborName)
		}
	}
	deployment.removeNFs(nfDeploy)
	deployment.logger.Info(
		"Report NFDeploy succeeded for", "NFDeploy", nfDeploy.Name,
	)
}

// updateSubscriptionFailureCondition: updates all NFConditions in NFDeploy status
// to unknown with given reason and message
func (deployment *Deployment) updateSubscriptionFailureCondition(
	reason string, message string,
) error {
	var conditionSet []v1alpha1.NFDeployCondition
	currentTime := metav1.NewTime(time.Now())
	stalledCondition := v1alpha1.NFDeployCondition{
		Type: v1alpha1.DeploymentStalled, Status: corev1.ConditionUnknown,
		LastUpdateTime:     currentTime,
		LastTransitionTime: currentTime,
		Reason:             reason,
		Message:            message,
	}
	readyCondition := v1alpha1.NFDeployCondition{
		Type: v1alpha1.DeploymentReady, Status: corev1.ConditionUnknown,
		LastUpdateTime:     currentTime,
		LastTransitionTime: currentTime,
		Reason:             reason,
		Message:            message,
	}
	peeringCondition := v1alpha1.NFDeployCondition{
		Type: v1alpha1.DeploymentPeering, Status: corev1.ConditionUnknown,
		LastUpdateTime:     currentTime,
		LastTransitionTime: currentTime,
		Reason:             reason,
		Message:            message,
	}
	reconcilingCondition := v1alpha1.NFDeployCondition{
		Type: v1alpha1.DeploymentReconciling, Status: corev1.ConditionUnknown,
		LastUpdateTime:     currentTime,
		LastTransitionTime: currentTime,
		Reason:             reason,
		Message:            message,
	}
	conditionSet = append(
		conditionSet, stalledCondition, readyCondition, peeringCondition,
		reconcilingCondition,
	)
	return deployment.updateNFDeployStatus(
		0, 0, 0, 0, &stalledCondition, &readyCondition, &peeringCondition,
		&reconcilingCondition,
	)
}

// ListenSubscriptionStatus listens for errors from edgewatcher during subscription
// creation. In case of no errors, it starts a thread to listen to edge events
func (deployment *Deployment) ListenSubscriptionStatus() {
	err, ok := <-deployment.edgeErrorChan
	if ok {
		if err == nil {
			deployment.listenEdgeEvents()
		} else {
			subscriptionErr := deployment.updateSubscriptionFailureCondition(
				"EdgeConnectionFailure",
				"Connection with edge failed due to "+err.Error(),
			)
			if subscriptionErr != nil {
				deployment.logger.Error(
					err, "FATAL ERROR: Unable to create connection with edge.",
					"NFDeploy",
					deployment.name,
				)
			}
		}
	} else {
		subscriptionErr := deployment.updateSubscriptionFailureCondition(
			"EdgeConnectionFailure", "Edge connection broke unexpectedly",
		)
		if subscriptionErr != nil {
			deployment.logger.Error(
				err, "FATAL ERROR: Unable to create connection with edge.", "NFDeploy",
				deployment.name,
			)
		}
	}
}

// processEdgeEvent processes a single edge event and updates nfDeploy status
func (deployment *Deployment) processEdgeEvent(object *preprocessor.Event) {
	deployment.deploymentMu.Lock()
	defer deployment.deploymentMu.Unlock()

	obj, ok := object.Object.(*unstructured.Unstructured)
	if !ok {
		deployment.logger.Info(
			"Received object is not of type *unstructured.Unstructured", "type",
			reflect.TypeOf(object.Object),
		)
		return
	}

	switch object.Key.Kind {
	case "UPFDeploy":
		upfDeploy := &types.UpfDeploy{}
		if err := runtime.DefaultUnstructuredConverter.
			FromUnstructured(obj.Object, upfDeploy); err != nil {
			deployment.logger.Info(
				"Unable to convert received UPFDeploy object to UPFDeploy type from *unstructured.Unstructured",
				"err", err.Error(),
			)
			return
		}
		upfName := upfDeploy.ObjectMeta.Labels[util.NFSiteIDLabel]
		deployment.logger.Info(
			"Edge event received for", "UPFDeploy", upfName,
		)
		if _, isPresent := deployment.upfNodes[upfName]; !isPresent {
			deployment.logger.Info(
				"The NF is not present in current deployment", "UPFDeploy",
				upfName,
			)
			return
		}
		// TODO: add testcase to verify stale events are discarded
		if object.Timestamp.Before(deployment.upfNodes[upfName].Status.lastEventTimestamp) {
			deployment.logger.Info(
				"The NF event received is of previous timestamp", "UPFDeploy",
				upfName,
			)
			return
		}
		upfNode := deployment.upfNodes[upfName]
		upfNode.Status.lastEventTimestamp = object.Timestamp
		deployment.upfNodes[upfName] = upfNode
		deployment.processNFEdgeEvent(
			&upfDeploy.Status.Conditions, upfName,
		)
	case "SMFDeploy":
		smfDeploy := &types.SmfDeploy{}
		if err := runtime.DefaultUnstructuredConverter.
			FromUnstructured(obj.Object, smfDeploy); err != nil {
			deployment.logger.Info(
				"Unable to convert received SMFDeploy object to SMFDeploy type from *unstructured.Unstructured",
				"err", err.Error(),
			)
			return
		}
		smfName := smfDeploy.ObjectMeta.Labels[util.NFSiteIDLabel]
		deployment.logger.Info(
			"Edge event received for", "SMFDeploy", smfName,
		)
		if _, isPresent := deployment.smfNodes[smfName]; !isPresent {
			deployment.logger.Info(
				"The NF is not present in current deployment", "SMFDeploy",
				smfName,
			)
			return
		}
		if object.Timestamp.Before(deployment.smfNodes[smfName].Status.lastEventTimestamp) {
			deployment.logger.Info(
				"The NF event received is of previous timestamp", "SMFDeploy",
				smfName,
			)
			return
		}
		smfNode := deployment.smfNodes[smfName]
		smfNode.Status.lastEventTimestamp = object.Timestamp
		deployment.smfNodes[smfName] = smfNode
		deployment.processNFEdgeEvent(
			&smfDeploy.Status.Conditions, smfName,
		)
	case "UDMDeploy":
		udmDeploy := &types2.UdmDeploy{}
		if err := runtime.DefaultUnstructuredConverter.
			FromUnstructured(obj.Object, udmDeploy); err != nil {
			deployment.logger.Info(
				"Unable to convert received UDMDeploy object to UDMDeploy type from *unstructured.Unstructured",
				"err", err.Error(),
			)
			return
		}
		udmName := udmDeploy.ObjectMeta.Labels[util.NFSiteIDLabel]
		deployment.logger.Info(
			"Edge event received for", "UDMDeploy", udmName,
		)
		if _, isPresent := deployment.udmNodes[udmName]; !isPresent {
			deployment.logger.Info(
				"The NF is not present in current deployment", "UDMDeploy",
				udmName,
			)
			return
		}
		if object.Timestamp.Before(deployment.udmNodes[udmName].Status.lastEventTimestamp) {
			deployment.logger.Info(
				"The NF event received is of previous timestamp", "UDMDeploy",
				udmName,
			)
			return
		}
		udmNode := deployment.udmNodes[udmName]
		udmNode.Status.lastEventTimestamp = object.Timestamp
		deployment.udmNodes[udmName] = udmNode
		deployment.processNFEdgeEvent(
			&udmDeploy.Status.Conditions, udmName,
		)
	case "AUSFDeploy":
		ausfDeploy := &types3.AusfDeploy{}
		if err := runtime.DefaultUnstructuredConverter.
			FromUnstructured(obj.Object, ausfDeploy); err != nil {
			// TODO: add test to execute error path
			deployment.logger.Info(
				"Unable to convert received AUSFDeploy object to AUSFDeploy type from *unstructured.Unstructured",
				"err", err.Error(),
			)
			return
		}
		ausfName := ausfDeploy.ObjectMeta.Labels[util.NFSiteIDLabel]
		deployment.logger.Info(
			"Edge event received for", "AUSFDeploy", ausfName,
		)
		if _, isPresent := deployment.ausfNodes[ausfName]; !isPresent {
			deployment.logger.Info(
				"The NF is not present in current deployment", "AUSFDeploy",
				ausfName,
			)
			return
		}
		if object.Timestamp.Before(deployment.ausfNodes[ausfName].Status.lastEventTimestamp) {
			deployment.logger.Info(
				"The NF event received is of previous timestamp", "AUSFDeploy",
				ausfName,
			)
			return
		}
		ausfNode := deployment.ausfNodes[ausfName]
		ausfNode.Status.lastEventTimestamp = object.Timestamp
		deployment.ausfNodes[ausfName] = ausfNode
		deployment.processNFEdgeEvent(
			&ausfDeploy.Status.Conditions, ausfName,
		)
	}
	// TODO: implement AMF Status once AMFDeploy is finalised
}

// listenEdgeEvents listens for events from edgewatcher through deploymentChan
func (deployment *Deployment) listenEdgeEvents() {
	for {
		select {
		case <-deployment.ctx.Done():
			return
		case object, ok := <-deployment.edgeEventsChan:
			if ok {
				deployment.processEdgeEvent(&object)
			} else {
				err := deployment.updateSubscriptionFailureCondition(
					"EdgeConnectionBroken", "Connection to edge broke unexpectedly.",
				)
				if err != nil {
					deployment.logger.Error(
						err, "FATAL ERROR: Connection to edge broke unexpectedly.",
						"NFDeploy",
						deployment.name,
					)
				}
			}
		}

	}
}

// processNFEdgeEvent : This method computes and updates aggregated status of
// NFDeploy resource based on the change in status of a single NF
func (deployment *Deployment) processNFEdgeEvent(
	nfConditions *[]types.NFCondition, nfId string,
) {
	conditions, conditionMessage := deployment.calculateNFConditionSet(nfConditions)

	if deployment.isAmbiguousConditionSet(conditions) {
		deployment.logger.Info(
			"Ambiguous NFConditions received. Edge event dropped for", "NF", nfId,
		)
		return
	}
	deployment.updateCurrentNFStatus(
		nfId, conditions, conditionMessage,
	)
	availableNFs, readyNFs, stalledNFs, targetedNFs := deployment.calculateNFCount()

	stalledCondition := deployment.computeStalledCondition(
		stalledNFs, targetedNFs,
	)
	readyCondition := deployment.computeReadyCondition(readyNFs, targetedNFs)

	peeringCondition := deployment.computePeeringCondition(readyNFs, targetedNFs)

	reconcilingCondition := deployment.computeReconcilingCondition(
		readyNFs, targetedNFs,
	)

	err := deployment.updateNFDeployStatus(
		int32(availableNFs), int32(readyNFs), int32(stalledNFs), int32(targetedNFs),
		&stalledCondition,
		&readyCondition, &peeringCondition, &reconcilingCondition,
	)
	if err != nil {
		deployment.logger.Error(
			err, "Failed to update NFDeployStatus for ", "NF", nfId,
		)
	}
}
