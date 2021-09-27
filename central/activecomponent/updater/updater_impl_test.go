package updater

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	acConverter "github.com/stackrox/rox/central/activecomponent/converter"
	acMocks "github.com/stackrox/rox/central/activecomponent/datastore/mocks"
	aggregatorPkg "github.com/stackrox/rox/central/activecomponent/updater/aggregator"
	"github.com/stackrox/rox/central/activecomponent/updater/aggregator/mocks"
	deploymentMocks "github.com/stackrox/rox/central/deployment/datastore/mocks"
	imageMocks "github.com/stackrox/rox/central/image/datastore/mocks"
	piMocks "github.com/stackrox/rox/central/processindicator/datastore/mocks"
	v1 "github.com/stackrox/rox/generated/api/v1"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/dackbox/edges"
	"github.com/stackrox/rox/pkg/features"
	"github.com/stackrox/rox/pkg/protoconv"
	"github.com/stackrox/rox/pkg/scancomponent"
	"github.com/stackrox/rox/pkg/search"
	"github.com/stackrox/rox/pkg/set"
	"github.com/stackrox/rox/pkg/simplecache"
	"github.com/stackrox/rox/pkg/testutils/envisolator"
	"github.com/stackrox/rox/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type indicatorModel struct {
	DeploymentID  string
	ContainerName string
	ImageID       string
	ExePaths      []string
}

var (
	mockDeployments = []*storage.Deployment{
		{
			Id: "depA",
			Containers: []*storage.Container{
				{
					Name:  "depA-C1-image1",
					Image: &storage.ContainerImage{Id: "image1"},
				},
				{
					Name:  "depA-C2-image1",
					Image: &storage.ContainerImage{Id: "image1"},
				},
			},
		},
		{
			Id: "depB",
			Containers: []*storage.Container{
				{
					Name:  "depB-C1-image2",
					Image: &storage.ContainerImage{Id: "image2"},
				},
				{
					Name:  "depB-C2-image1",
					Image: &storage.ContainerImage{Id: "image1"},
				},
			},
		},
		{
			Id: "depC",
			Containers: []*storage.Container{
				{
					Name:  "depC-C1-image2",
					Image: &storage.ContainerImage{Id: "image2"},
				},
			},
		},
	}
	mockImage = &storage.Image{
		Id: "image1",
		Scan: &storage.ImageScan{
			ScanTime: protoconv.ConvertTimeToTimestamp(time.Now()),
			Components: []*storage.EmbeddedImageScanComponent{
				{
					Name:    "image1_component1",
					Version: "1",
					Source:  storage.SourceType_OS,
					Executables: []*storage.EmbeddedImageScanComponent_Executable{
						{Path: "/root/bin/image1_component1_match_file1"},
						{Path: "/root/bin/image1_component1_nonmatch_file2"},
						{Path: "/root/bin/image1_component1_nonmatch_file3"},
					},
				},
				{
					Name:    "image1_component2",
					Version: "2",
					Source:  storage.SourceType_OS,
					Executables: []*storage.EmbeddedImageScanComponent_Executable{
						{Path: "/root/bin/image1_component2_nonmatch_file1"},
						{Path: "/root/bin/image1_component2_nonmatch_file2"},
						{Path: "/root/bin/image1_component2_match_file3"},
					},
				},
				{
					Name:    "image1_component3",
					Version: "2",
					Source:  storage.SourceType_JAVA,
				},
				{
					Name:    "image1_component4",
					Version: "2",
					Source:  storage.SourceType_OS,
					Executables: []*storage.EmbeddedImageScanComponent_Executable{
						{Path: "/root/bin/image1_component4_nonmatch_file1"},
						{Path: "/root/bin/image1_component4_nonmatch_file2"},
						{Path: "/root/bin/image1_component4_match_file3"},
					},
				},
			},
		},
	}
	mockIndicators = []indicatorModel{
		{
			DeploymentID:  "depA",
			ContainerName: "depA-C1-image1",
			ImageID:       mockImage.Id,
			ExePaths: []string{
				"/root/bin/image1_component1_match_file1",
				"/root/bin/image1_component2_match_file3",
				"/root/bin/image1_component3_match_file1",
				"/root/bin/image1_component3_match_file2",
			},
		},
		{
			DeploymentID:  "depB",
			ContainerName: "depB-C2-image1",
			ImageID:       mockImage.Id,
			ExePaths: []string{
				"/root/bin/image1_component1_match_file1",
				"/root/bin/image1_component3_match_file3",
				"/root/bin/image1_component4_match_file3",
			},
		},
	}
)

func TestActiveComponentUpdater(t *testing.T) {
	suite.Run(t, new(acUpdaterTestSuite))
}

type acUpdaterTestSuite struct {
	suite.Suite

	mockCtrl                      *gomock.Controller
	mockDeploymentDatastore       *deploymentMocks.MockDataStore
	mockActiveComponentDataStore  *acMocks.MockDataStore
	mockProcessIndicatorDataStore *piMocks.MockDataStore
	envIsolator                   *envisolator.EnvIsolator
	mockImageDataStore            *imageMocks.MockDataStore
	executableCache               simplecache.Cache
	mockAggregator                *mocks.MockProcessAggregator
}

func (s *acUpdaterTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockDeploymentDatastore = deploymentMocks.NewMockDataStore(s.mockCtrl)
	s.mockActiveComponentDataStore = acMocks.NewMockDataStore(s.mockCtrl)
	s.mockProcessIndicatorDataStore = piMocks.NewMockDataStore(s.mockCtrl)
	s.mockImageDataStore = imageMocks.NewMockDataStore(s.mockCtrl)
	s.executableCache = simplecache.New()
	s.envIsolator = envisolator.NewEnvIsolator(s.T())
	s.envIsolator.Setenv(features.ActiveVulnManagement.EnvVar(), "true")
	s.mockAggregator = mocks.NewMockProcessAggregator(s.mockCtrl)

	if !features.ActiveVulnManagement.Enabled() {
		s.T().Skip("Skip active component updater test")
		s.T().SkipNow()
	}
}

func (s *acUpdaterTestSuite) TearDownTest() {
	s.envIsolator.RestoreAll()
	s.mockCtrl.Finish()
}

func (s *acUpdaterTestSuite) TestUpdater() {
	imageID := "image1"
	updater := &updaterImpl{
		acStore:         s.mockActiveComponentDataStore,
		deploymentStore: s.mockDeploymentDatastore,
		piStore:         s.mockProcessIndicatorDataStore,
		imageStore:      s.mockImageDataStore,
		aggregator:      aggregatorPkg.NewAggregator(),
		executableCache: simplecache.New(),
	}
	var deploymentIDs []string
	for _, deployment := range mockDeployments {
		for _, container := range deployment.GetContainers() {
			if container.GetImage().GetId() == imageID {
				deploymentIDs = append(deploymentIDs, deployment.GetId())
			}
		}
	}
	s.mockDeploymentDatastore.EXPECT().GetDeploymentIDs().AnyTimes().Return(deploymentIDs, nil)
	s.mockActiveComponentDataStore.EXPECT().SearchRawActiveComponents(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
	s.mockProcessIndicatorDataStore.EXPECT().SearchRawProcessIndicators(gomock.Any(), gomock.Any()).Times(3).DoAndReturn(
		func(ctx context.Context, query *v1.Query) ([]*storage.ProcessIndicator, error) {
			queries := query.GetConjunction().GetQueries()
			s.Assert().Len(queries, 2)
			var containerName, deploymentID string
			for _, q := range queries {
				mf := q.GetBaseQuery().GetMatchFieldQuery()

				switch mf.GetField() {
				case search.DeploymentID.String():
					deploymentID = stripQuotes(mf.GetValue())
				case search.ContainerName.String():
					containerName = stripQuotes(mf.GetValue())
				default:
					s.Assert().Fail("unexpected query")
				}
			}
			for _, pi := range mockIndicators {
				if pi.ContainerName == containerName && deploymentID == pi.DeploymentID {
					var ret []*storage.ProcessIndicator
					for _, exec := range pi.ExePaths {
						ret = append(ret, &storage.ProcessIndicator{
							Id:      uuid.NewV4().String(),
							ImageId: pi.ImageID,
							Signal:  &storage.ProcessSignal{ExecFilePath: exec}},
						)
					}
					return ret, nil
				}
			}
			return nil, nil
		})
	s.mockImageDataStore.EXPECT().Search(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
		func(ctx context.Context, query *v1.Query) ([]search.Result, error) {
			return []search.Result{{ID: imageID}}, nil
		})
	s.mockActiveComponentDataStore.EXPECT().UpsertBatch(gomock.Any(), gomock.Any()).AnyTimes().Return(nil).Do(func(_ context.Context, acs []*acConverter.CompleteActiveComponent) {
		s.Assert().Equal(2, len(acs))
		for _, ac := range acs {
			edge, err := edges.FromString(ac.ActiveComponent.GetId())
			s.Assert().NoError(err)
			// Deployment C does not have image1.
			s.Assert().NotEqual(edge.ParentID, mockDeployments[2].GetId())
			imageComponent, err := edges.FromString(edge.ChildID)
			s.Assert().NoError(err)
			s.Assert().True(strings.HasPrefix(imageComponent.ParentID, mockImage.GetId()))
			s.Assert().NotEqual(imageComponent.ParentID, mockImage.GetScan().GetComponents()[2].GetName())
			s.Assert().Len(ac.ActiveComponent.ActiveContexts, 1)

			var expectedComponent *storage.EmbeddedImageScanComponent
			var expectedContainer string
			if edge.ParentID == mockDeployments[0].Id {
				expectedContainer = mockIndicators[0].ContainerName
				// Component 1 or 2
				expectedComponent = mockImage.GetScan().GetComponents()[0]
				if imageComponent.ParentID != mockImage.GetScan().GetComponents()[0].GetName() {
					expectedComponent = mockImage.GetScan().GetComponents()[1]
				}
			} else {
				s.Assert().Equal(edge.ParentID, mockDeployments[1].Id)
				expectedContainer = mockIndicators[1].ContainerName
				// Component 1 or 4
				expectedComponent = mockImage.GetScan().GetComponents()[0]
				if imageComponent.ParentID != mockImage.GetScan().GetComponents()[0].GetName() {
					expectedComponent = mockImage.GetScan().GetComponents()[3]
				}
			}
			s.Assert().Contains(ac.ActiveComponent.ActiveContexts, expectedContainer)
			s.Assert().True(strings.HasSuffix(imageComponent.ParentID, expectedComponent.GetName()))
			s.Assert().Equal(imageComponent, edges.EdgeID{ParentID: expectedComponent.GetName(), ChildID: expectedComponent.Version})
		}
	})

	s.Assert().NoError(updater.PopulateExecutableCache(updaterCtx, mockImage))
	for _, deployment := range mockDeployments {
		updater.aggregator.RefreshDeployment(deployment)
	}
	updater.Update()
}

func (s *acUpdaterTestSuite) TestUpdater_PopulateExecutableCache() {
	updater := &updaterImpl{
		acStore:         s.mockActiveComponentDataStore,
		deploymentStore: s.mockDeploymentDatastore,
		piStore:         s.mockProcessIndicatorDataStore,
		imageStore:      s.mockImageDataStore,
		aggregator:      s.mockAggregator,
		executableCache: simplecache.New(),
	}

	// Initial population
	image := mockImage.Clone()
	s.Assert().NoError(updater.PopulateExecutableCache(updaterCtx, image))
	s.verifyExecutableCache(updater, mockImage)

	// Verify the executables are not stored.
	for _, component := range image.GetScan().GetComponents() {
		s.Assert().Empty(component.Executables)
	}

	// Image won't be processed again.
	s.Assert().NoError(updater.PopulateExecutableCache(updaterCtx, image))
	s.verifyExecutableCache(updater, mockImage)

	// New update without the first component
	image = mockImage.Clone()
	image.GetScan().ScanTime = protoconv.ConvertTimeToTimestamp(time.Now().Add(1))
	image.GetScan().Components = image.GetScan().GetComponents()[1:]
	imageForVerify := image.Clone()
	s.Assert().NoError(updater.PopulateExecutableCache(updaterCtx, image))
	s.verifyExecutableCache(updater, imageForVerify)
}

func (s *acUpdaterTestSuite) verifyExecutableCache(updater *updaterImpl, image *storage.Image) {
	s.Assert().Len(updater.executableCache.Keys(), 1)
	result, ok := updater.executableCache.Get(image.GetId())
	s.Assert().True(ok)
	execToComponent := result.(*imageExecutable).execToComponent
	allExecutables := set.NewStringSet()
	for _, component := range image.GetScan().GetComponents() {
		if component.Source != storage.SourceType_OS {
			continue
		}
		componentID := scancomponent.ComponentID(component.GetName(), component.GetVersion())
		for _, exec := range component.Executables {
			s.Assert().Contains(execToComponent, exec.GetPath())
			s.Assert().Equal(componentID, execToComponent[exec.GetPath()])
			allExecutables.Add(exec.GetPath())
		}
	}
	s.Assert().Len(execToComponent, len(allExecutables))
}
func (s *acUpdaterTestSuite) TestUpdater_Update() {
	updater := &updaterImpl{
		acStore:         s.mockActiveComponentDataStore,
		deploymentStore: s.mockDeploymentDatastore,
		piStore:         s.mockProcessIndicatorDataStore,
		imageStore:      s.mockImageDataStore,
		aggregator:      s.mockAggregator,
		executableCache: simplecache.New(),
	}
	image := &storage.Image{
		Id: "image1",
		Scan: &storage.ImageScan{
			ScanTime: protoconv.ConvertTimeToTimestamp(time.Now()),
			Components: []*storage.EmbeddedImageScanComponent{
				{
					Name:    "component1",
					Version: "1",
					Source:  storage.SourceType_OS,
					Executables: []*storage.EmbeddedImageScanComponent_Executable{
						{Path: "/usr/bin/component1_file1"},
						{Path: "/usr/bin/component1_file2"},
						{Path: "/usr/bin/component1_file3"},
					},
				},
				{
					Name:    "component2",
					Version: "1",
					Source:  storage.SourceType_OS,
					Executables: []*storage.EmbeddedImageScanComponent_Executable{
						{Path: "/usr/bin/component2_file1"},
						{Path: "/usr/bin/component2_file2"},
						{Path: "/usr/bin/component2_file3"},
					},
				},
			},
		},
	}
	imageScan := image.GetScan()
	components := imageScan.GetComponents()
	deployment := mockDeployments[0]
	var componentsIDs []string
	for _, component := range components {
		componentsIDs = append(componentsIDs, scancomponent.ComponentID(component.GetName(), component.GetVersion()))
	}

	var containerNames []string
	for _, container := range deployment.GetContainers() {
		containerNames = append(containerNames, container.GetName())
	}

	s.mockImageDataStore.EXPECT().Search(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, query *v1.Query) ([]search.Result, error) {
			return []search.Result{{ID: image.GetId()}}, nil
		})
	s.Assert().NoError(updater.PopulateExecutableCache(updaterCtx, image.Clone()))
	s.mockDeploymentDatastore.EXPECT().GetDeploymentIDs().AnyTimes().Return([]string{deployment.GetId()}, nil)

	// Test active components with designated image and deployment
	var testCases = []struct {
		description string

		updates     []*aggregatorPkg.ProcessUpdate
		indicatiors map[string]indicatorModel
		existingAcs map[string]set.StringSet // componentID to container name map

		acsToUpdate map[string]set.StringSet // expected Acs to be updated, componentID to container name map
		acsToDelete []string
	}{
		{
			description: "First populate from database",
			updates: []*aggregatorPkg.ProcessUpdate{
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[0], set.NewStringSet(), aggregatorPkg.FromDatabase),
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[1], set.NewStringSet(), aggregatorPkg.FromDatabase),
			},
			indicatiors: map[string]indicatorModel{
				containerNames[0]: {
					DeploymentID:  deployment.GetId(),
					ContainerName: containerNames[0],
					ImageID:       image.GetId(),
					ExePaths: []string{
						components[0].Executables[0].Path,
						components[1].Executables[1].Path,
					},
				},
				containerNames[1]: {
					DeploymentID:  deployment.GetId(),
					ContainerName: containerNames[1],
					ImageID:       image.GetId(),
					ExePaths: []string{
						components[1].Executables[0].Path,
					},
				},
			},
			acsToUpdate: map[string]set.StringSet{
				componentsIDs[0]: set.NewStringSet(deployment.Containers[0].GetName()),
				componentsIDs[1]: set.NewStringSet(deployment.Containers[0].GetName(), deployment.Containers[1].GetName()),
			},
		},
		{
			description: "Image change populate from database no updates",
			updates: []*aggregatorPkg.ProcessUpdate{
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[0], set.NewStringSet(), aggregatorPkg.FromDatabase),
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[1], set.NewStringSet(), aggregatorPkg.FromDatabase),
			},
			indicatiors: map[string]indicatorModel{
				containerNames[0]: {
					DeploymentID:  deployment.GetId(),
					ContainerName: containerNames[0],
					ImageID:       image.GetId(),
					ExePaths: []string{
						components[0].Executables[0].Path,
						components[1].Executables[1].Path,
					},
				},
				containerNames[1]: {
					DeploymentID:  deployment.GetId(),
					ContainerName: containerNames[1],
					ImageID:       image.GetId(),
					ExePaths: []string{
						components[1].Executables[0].Path,
					},
				},
			},
			existingAcs: map[string]set.StringSet{
				componentsIDs[0]: set.NewStringSet(deployment.Containers[0].GetName()),
				componentsIDs[1]: set.NewStringSet(deployment.Containers[0].GetName(), deployment.Containers[1].GetName()),
			},
		},
		{
			description: "Image change populate from database with updates",
			updates: []*aggregatorPkg.ProcessUpdate{
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[0], set.NewStringSet(), aggregatorPkg.FromDatabase),
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[1], set.NewStringSet(), aggregatorPkg.FromDatabase),
			},
			indicatiors: map[string]indicatorModel{
				containerNames[0]: {
					DeploymentID:  deployment.GetId(),
					ContainerName: containerNames[0],
					ImageID:       image.GetId(),
					ExePaths: []string{
						components[0].Executables[0].Path,
						components[1].Executables[1].Path,
					},
				},
				containerNames[1]: {
					DeploymentID:  deployment.GetId(),
					ContainerName: containerNames[1],
					ImageID:       image.GetId(),
					ExePaths: []string{
						components[1].Executables[0].Path,
					},
				},
			},
			existingAcs: map[string]set.StringSet{
				componentsIDs[0]: set.NewStringSet(deployment.Containers[1].GetName()),
				componentsIDs[1]: set.NewStringSet(deployment.Containers[0].GetName()),
			},
			acsToUpdate: map[string]set.StringSet{
				componentsIDs[0]: set.NewStringSet(deployment.Containers[0].GetName()),
				componentsIDs[1]: set.NewStringSet(deployment.Containers[0].GetName(), deployment.Containers[1].GetName()),
			},
		},
		{
			description: "Image change populate from database with removal",
			updates: []*aggregatorPkg.ProcessUpdate{
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[0], set.NewStringSet(), aggregatorPkg.FromDatabase),
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[1], set.NewStringSet(), aggregatorPkg.FromDatabase),
			},
			indicatiors: map[string]indicatorModel{
				containerNames[0]: {
					DeploymentID:  deployment.GetId(),
					ContainerName: containerNames[0],
					ImageID:       image.GetId(),
					ExePaths: []string{
						components[1].Executables[0].Path,
						components[1].Executables[1].Path,
					},
				},
				containerNames[1]: {
					DeploymentID:  deployment.GetId(),
					ContainerName: containerNames[1],
					ImageID:       image.GetId(),
					ExePaths: []string{
						components[1].Executables[0].Path,
					},
				},
			},
			existingAcs: map[string]set.StringSet{
				componentsIDs[0]: set.NewStringSet(deployment.Containers[1].GetName()),
				componentsIDs[1]: set.NewStringSet(deployment.Containers[0].GetName()),
			},
			acsToUpdate: map[string]set.StringSet{
				componentsIDs[1]: set.NewStringSet(deployment.Containers[0].GetName(), deployment.Containers[1].GetName()),
			},
			acsToDelete: []string{componentsIDs[0]},
		},
		{
			description: "Image change populate from database with removal request",
			updates: []*aggregatorPkg.ProcessUpdate{
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[0], set.NewStringSet(), aggregatorPkg.ToBeRemoved),
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[1], set.NewStringSet(), aggregatorPkg.FromDatabase),
			},
			indicatiors: map[string]indicatorModel{
				containerNames[0]: {
					DeploymentID:  deployment.GetId(),
					ContainerName: containerNames[0],
					ImageID:       image.GetId(),
					ExePaths: []string{
						components[1].Executables[0].Path,
						components[1].Executables[1].Path,
					},
				},
				containerNames[1]: {
					DeploymentID:  deployment.GetId(),
					ContainerName: containerNames[1],
					ImageID:       image.GetId(),
					ExePaths: []string{
						components[0].Executables[0].Path,
					},
				},
			},
			existingAcs: map[string]set.StringSet{
				componentsIDs[0]: set.NewStringSet(deployment.Containers[1].GetName()),
				componentsIDs[1]: set.NewStringSet(deployment.Containers[0].GetName()),
			},
			acsToDelete: []string{componentsIDs[1]},
		},
		{
			description: "Update from cache adding new contexts",
			updates: []*aggregatorPkg.ProcessUpdate{
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[0], set.NewStringSet(components[0].Executables[0].Path), aggregatorPkg.FromCache),
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[1], set.NewStringSet(components[1].Executables[0].Path, components[1].Executables[1].Path), aggregatorPkg.FromCache),
			},
			existingAcs: map[string]set.StringSet{
				componentsIDs[0]: set.NewStringSet(deployment.Containers[1].GetName()),
				componentsIDs[1]: set.NewStringSet(deployment.Containers[0].GetName()),
			},
			acsToUpdate: map[string]set.StringSet{
				componentsIDs[0]: set.NewStringSet(deployment.Containers[0].GetName(), deployment.Containers[1].GetName()),
				componentsIDs[1]: set.NewStringSet(deployment.Containers[0].GetName(), deployment.Containers[1].GetName()),
			},
		},
		{
			description: "update from cache no new change and no updates",
			updates: []*aggregatorPkg.ProcessUpdate{
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[0], set.NewStringSet(components[0].Executables[0].Path), aggregatorPkg.FromCache),
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[1], set.NewStringSet(components[0].Executables[1].Path, components[1].Executables[0].Path, components[1].Executables[1].Path), aggregatorPkg.FromCache),
			},
			existingAcs: map[string]set.StringSet{
				componentsIDs[0]: set.NewStringSet(deployment.Containers[0].GetName(), deployment.Containers[1].GetName()),
				componentsIDs[1]: set.NewStringSet(deployment.Containers[0].GetName(), deployment.Containers[1].GetName()),
			},
			acsToUpdate: map[string]set.StringSet{},
		},
		{
			description: "update from cache with removal request",
			updates: []*aggregatorPkg.ProcessUpdate{
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[0], set.NewStringSet(), aggregatorPkg.ToBeRemoved),
				aggregatorPkg.NewProcessUpdate(image.GetId(), containerNames[1], set.NewStringSet(components[0].Executables[1].Path), aggregatorPkg.FromCache),
			},
			// This should not be used in this test case.
			indicatiors: map[string]indicatorModel{
				containerNames[0]: {
					DeploymentID:  deployment.GetId(),
					ContainerName: containerNames[0],
					ImageID:       image.GetId(),
					ExePaths: []string{
						components[0].Executables[0].Path,
						components[1].Executables[1].Path,
					},
				},
				containerNames[1]: {
					DeploymentID:  deployment.GetId(),
					ContainerName: containerNames[1],
					ImageID:       image.GetId(),
					ExePaths: []string{
						components[0].Executables[0].Path,
						components[1].Executables[1].Path,
					},
				},
			},
			existingAcs: map[string]set.StringSet{
				componentsIDs[1]: set.NewStringSet(deployment.Containers[0].GetName()),
			},
			acsToUpdate: map[string]set.StringSet{
				componentsIDs[0]: set.NewStringSet(deployment.Containers[1].GetName()),
			},
			acsToDelete: []string{componentsIDs[1]},
		},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.description, func(t *testing.T) {
			s.mockAggregator.EXPECT().GetAndPrune(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(_ func(string) bool, deploymentsSet set.StringSet) map[string][]*aggregatorPkg.ProcessUpdate {
					return map[string][]*aggregatorPkg.ProcessUpdate{
						deployment.GetId(): testCase.updates,
					}
				})
			var databaseFetchCount int
			for _, update := range testCase.updates {
				if update.FromDatabase() {
					databaseFetchCount++
				}
			}
			if databaseFetchCount > 0 {
				s.mockProcessIndicatorDataStore.EXPECT().SearchRawProcessIndicators(gomock.Any(), gomock.Any()).Times(databaseFetchCount).DoAndReturn(
					func(ctx context.Context, query *v1.Query) ([]*storage.ProcessIndicator, error) {
						queries := query.GetConjunction().Queries
						s.Assert().Len(queries, 2)
						var containerName string
						for _, q := range queries {
							mf := q.GetBaseQuery().GetMatchFieldQuery()

							switch mf.GetField() {
							case search.DeploymentID.String():
								assert.Equal(t, strconv.Quote(deployment.GetId()), mf.GetValue())
							case search.ContainerName.String():
								containerName = stripQuotes(mf.GetValue())
							default:
								s.Assert().Fail("unexpected query")
							}
						}

						var ret []*storage.ProcessIndicator

						for _, exec := range testCase.indicatiors[containerName].ExePaths {
							ret = append(ret, &storage.ProcessIndicator{
								Id:            uuid.NewV4().String(),
								ImageId:       testCase.indicatiors[containerName].ImageID,
								DeploymentId:  deployment.GetId(),
								ContainerName: containerName,
								Signal:        &storage.ProcessSignal{ExecFilePath: exec}},
							)
						}
						return ret, nil
					})
			}
			s.mockActiveComponentDataStore.EXPECT().SearchRawActiveComponents(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
				func(ctx context.Context, query *v1.Query) ([]*storage.ActiveComponent, error) {
					// Verify query
					assert.Equal(t, search.DeploymentID.String(), query.GetBaseQuery().GetMatchFieldQuery().GetField())
					assert.Equal(t, strconv.Quote(deployment.GetId()), query.GetBaseQuery().GetMatchFieldQuery().GetValue())
					var ret []*storage.ActiveComponent
					for componentID, containerNames := range testCase.existingAcs {
						edge := edges.EdgeID{ParentID: deployment.GetId(), ChildID: componentID}
						ac := &storage.ActiveComponent{
							Id:             edge.ToString(),
							ActiveContexts: make(map[string]*storage.ActiveComponent_ActiveContext),
						}
						for containerName := range containerNames {
							ac.ActiveContexts[containerName] = &storage.ActiveComponent_ActiveContext{ContainerName: containerName}
						}
						ret = append(ret, ac)
					}
					return ret, nil
				})
			s.mockActiveComponentDataStore.EXPECT().GetBatch(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
				func(ctx context.Context, ids []string) ([]*storage.ActiveComponent, error) {
					var ret []*storage.ActiveComponent
					requestedIds := set.NewStringSet(ids...)
					for componentID, containerNames := range testCase.existingAcs {
						edge := edges.EdgeID{ParentID: deployment.GetId(), ChildID: componentID}
						if !requestedIds.Contains(edge.ToString()) {
							continue
						}
						ac := &storage.ActiveComponent{
							Id:             edge.ToString(),
							ActiveContexts: make(map[string]*storage.ActiveComponent_ActiveContext),
						}
						for containerName := range containerNames {
							ac.ActiveContexts[containerName] = &storage.ActiveComponent_ActiveContext{ContainerName: containerName}
						}
						ret = append(ret, ac)
					}
					return ret, nil
				})

			// Verify active components to be updated or deleted
			if len(testCase.acsToDelete) > 0 {
				s.mockActiveComponentDataStore.EXPECT().DeleteBatch(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
					func(ctx context.Context, ids ...string) error {
						expectedToDelete := set.NewStringSet()
						for _, componentID := range testCase.acsToDelete {
							expectedToDelete.Add(edges.EdgeID{ParentID: deployment.GetId(), ChildID: componentID}.ToString())
						}
						assert.Equal(t, expectedToDelete, set.NewStringSet(ids...))
						return nil
					})
			}
			if len(testCase.acsToUpdate) > 0 {
				s.mockActiveComponentDataStore.EXPECT().UpsertBatch(gomock.Any(), gomock.Any()).Times(1).Return(nil).Do(func(_ context.Context, acs []*acConverter.CompleteActiveComponent) {
					// Verify active components
					assert.Equal(t, len(testCase.acsToUpdate), len(acs))
					actualAcs := make(map[string]*acConverter.CompleteActiveComponent, len(acs))
					for _, ac := range acs {
						edge, err := edges.FromString(ac.ActiveComponent.GetId())
						assert.NoError(t, err)
						actualAcs[edge.ToString()] = ac
					}

					for componentID, expectedContexts := range testCase.acsToUpdate {
						edge := edges.EdgeID{ParentID: deployment.GetId(), ChildID: componentID}
						assert.Contains(t, actualAcs, edge.ToString())
						assert.Equal(t, deployment.GetId(), actualAcs[edge.ToString()].DeploymentID)
						assert.Equal(t, componentID, actualAcs[edge.ToString()].ComponentID)
						assert.Equal(t, edge.ToString(), actualAcs[edge.ToString()].ActiveComponent.GetId())
						assert.Equal(t, expectedContexts.Cardinality(), len(actualAcs[edge.ToString()].ActiveComponent.ActiveContexts))
						for containerName, context := range actualAcs[edge.ToString()].ActiveComponent.ActiveContexts {
							assert.Contains(t, expectedContexts, containerName)
							assert.Equal(t, containerName, context.ContainerName)
						}
					}
				})
			}
			updater.Update()
		})
	}
}

func stripQuotes(value string) string {
	return value[1 : len(value)-1]
}