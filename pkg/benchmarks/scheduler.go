package benchmarks

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"bitbucket.org/stack-rox/apollo/generated/api/v1"
	"bitbucket.org/stack-rox/apollo/pkg/clientconn"
	"bitbucket.org/stack-rox/apollo/pkg/env"
	"bitbucket.org/stack-rox/apollo/pkg/logging"
	"bitbucket.org/stack-rox/apollo/pkg/orchestrators"
	"bitbucket.org/stack-rox/apollo/pkg/protoconv"
	"bitbucket.org/stack-rox/apollo/pkg/uuid"
	"github.com/deckarep/golang-set"
	"github.com/golang/protobuf/ptypes"
)

var (
	log = logging.LoggerForModule()
)

const (
	cleanupTimeout = 1 * time.Minute
	retries        = 5
	updateInterval = 30 * time.Second
	// triggerTimespan is how long we should check for unfired triggers
	triggerTimespan = 5 * time.Minute

	benchmarkServiceName = "benchmark"
)

var (
	replaceRegex = regexp.MustCompile(`(\.|\s)`)

	staticIDNamespace = uuid.FromStringOrPanic("0a41c738-16d8-4e82-8e1b-921e5bb3d1c5")
)

type scheduleMetadata struct {
	*v1.BenchmarkSchedule
	NextScanTime time.Time
}

// SchedulerClient schedules the docker benchmark
type SchedulerClient struct {
	updateTicker   *time.Ticker
	scheduleTicker *time.Ticker
	orchestrator   orchestrators.Orchestrator

	advertisedEndpoint string
	centralEndpoint    string
	clusterID          string
	image              string

	started bool
	done    chan struct{}

	schedules map[string]*scheduleMetadata
	triggers  map[string]*v1.BenchmarkTrigger

	// Channel for enqueuing Scans (note: checks will be populated by the consumer)
	benchmarkChan chan *v1.BenchmarkScanMetadata
}

// NewSchedulerClient returns a new scheduler
func NewSchedulerClient(orchestrator orchestrators.Orchestrator, centralEndpoint, advertisedEndpoint, image string, clusterID string) (*SchedulerClient, error) {
	return &SchedulerClient{
		updateTicker:       time.NewTicker(updateInterval),
		orchestrator:       orchestrator,
		done:               make(chan struct{}),
		clusterID:          clusterID,
		centralEndpoint:    centralEndpoint,
		advertisedEndpoint: advertisedEndpoint,
		image:              image,

		schedules: make(map[string]*scheduleMetadata),
		triggers:  make(map[string]*v1.BenchmarkTrigger),

		benchmarkChan: make(chan *v1.BenchmarkScanMetadata, 512),
	}, nil
}

func grpcContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), requestTimeout)
}

func (s *SchedulerClient) getSchedules() ([]*v1.BenchmarkSchedule, error) {
	conn, err := clientconn.GRPCConnection(s.centralEndpoint)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := grpcContext()
	defer cancel()
	scheduleResp, err := v1.NewBenchmarkScheduleServiceClient(conn).GetBenchmarkSchedules(ctx, &v1.GetBenchmarkSchedulesRequest{
		Cluster: s.clusterID,
	})
	if err != nil {
		return nil, fmt.Errorf("Error checking schedule: %s", err)
	}
	return scheduleResp.Schedules, nil
}

func (s *SchedulerClient) benchmarkScanExists(scanID, benchmarkName string) (bool, error) {
	conn, err := clientconn.GRPCConnection(s.centralEndpoint)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	ctx, cancel := grpcContext()
	defer cancel()
	log.Infof("Fetching benchmark scan: %v", scanID)
	scan, err := v1.NewBenchmarkScanServiceClient(conn).GetBenchmarkScan(ctx, &v1.GetBenchmarkScanRequest{
		ScanId:     scanID,
		ClusterIds: []string{s.clusterID},
	})
	if err != nil {
		return false, fmt.Errorf("error checking results: %s", err)
	}
	return len(scan.GetChecks()) > 0, nil
}

func (s *SchedulerClient) getTriggers() ([]*v1.BenchmarkTrigger, error) {
	conn, err := clientconn.GRPCConnection(s.centralEndpoint)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := grpcContext()
	defer cancel()

	ts := ptypes.TimestampNow()
	ts.Seconds -= int64(triggerTimespan.Seconds())
	triggerResp, err := v1.NewBenchmarkTriggerServiceClient(conn).GetTriggers(ctx, &v1.GetBenchmarkTriggersRequest{
		ClusterIds: []string{s.clusterID},
		FromTime:   ts,
	})
	if err != nil {
		return nil, err
	}
	return triggerResp.Triggers, err
}

// Need to see if we have launched a trigger before
func (s *SchedulerClient) initializeTriggers() {
	triggers, err := s.getTriggers()
	if err != nil {
		log.Error(err)
		return
	}
	for _, trigger := range triggers {
		triggered, err := ptypes.Timestamp(trigger.GetTime())
		if err != nil {
			log.Errorf("Could not convert triggered time %v to golang type", trigger.GetTime())
			continue
		}
		scanID := uniqueScanID(triggered, trigger.GetName(), "triggered")
		exists, err := s.benchmarkScanExists(scanID, trigger.GetName())
		if err != nil {
			log.Errorf("Error getting benchmark results for scan %v", scanID)
			continue
		}
		if exists {
			s.triggers[trigger.Name] = trigger
		}
	}
}

func (s *SchedulerClient) removeService(id string) {
	for i := 1; i < retries+1; i++ {
		if err := s.orchestrator.Kill(id); err != nil {
			log.Errorf("Error removing benchmark service %v: %+v", id, err)
		} else {
			return
		}
		time.Sleep(time.Duration(i) * 2 * time.Second)
	}
	log.Error("Timed out trying to remove benchmark service")
}

func (s *SchedulerClient) waitForBenchmarkToFinish(serviceName string) {
	if err := s.orchestrator.WaitForCompletion(serviceName, cleanupTimeout); err != nil {
		log.Errorf("Error waiting for completion of %v: %+v", serviceName, err)
	}
	s.removeService(serviceName)
}

// Launch triggers a run of the benchmark immediately.
// The stateLock must be held by the caller until this function returns.
func (s *SchedulerClient) Launch(scan *v1.BenchmarkScanMetadata) error {
	service := orchestrators.SystemService{
		Name: benchmarkServiceName,
		Envs: []string{
			env.Combine(env.Image.EnvVar(), s.image),
			env.CombineSetting(env.AdvertisedEndpoint),
			env.Combine(env.ScanID.EnvVar(), scan.GetScanId()),
			env.Combine(env.Checks.EnvVar(), strings.Join(scan.GetChecks(), ",")),
			env.Combine(env.BenchmarkName.EnvVar(), scan.GetBenchmark()),
			env.Combine(env.BenchmarkReason.EnvVar(), scan.GetReason().String()),
		},
		Image:  s.image,
		Global: true,
	}
	_, err := s.orchestrator.LaunchBenchmark(service)
	if err != nil {
		return err
	}
	s.waitForBenchmarkToFinish(benchmarkServiceName)
	return nil
}

// ParseHour parses out a time in the form 03:04 PM
func ParseHour(h string) (time.Time, error) {
	hourTime, err := time.Parse("03:04 PM", h)
	if err != nil {
		return hourTime, fmt.Errorf("could not parse hour '%v'", h)
	}
	return hourTime, nil
}

var dayMap = map[string]struct{}{
	"Sunday":    {},
	"Monday":    {},
	"Tuesday":   {},
	"Wednesday": {},
	"Thursday":  {},
	"Friday":    {},
	"Saturday":  {},
}

// ValidDay makes sure that the string is a valid day of the week
func ValidDay(d string) bool {
	_, ok := dayMap[d]
	return ok
}

func nextScheduledTime(schedule *v1.BenchmarkSchedule) (time.Time, error) {
	hourTime, err := ParseHour(schedule.GetHour())
	if err != nil {
		return hourTime, err
	}

	nowTimezone := time.Now().Add(-time.Duration(schedule.TimezoneOffset) * time.Hour)
	nextTime := time.Date(nowTimezone.Year(), nowTimezone.Month(), nowTimezone.Day(), int(hourTime.Hour()), hourTime.Minute(), 0, 0, nowTimezone.Location())
	for nextTime.Before(nowTimezone) || nextTime.Weekday().String() != schedule.GetDay() {
		nextTime = nextTime.AddDate(0, 0, 1)
	}
	// Move nextTime back into UTC
	nextTime = nextTime.Add(time.Duration(schedule.TimezoneOffset) * time.Hour)
	log.Infof("Next time: %v", nextTime)
	return nextTime, nil
}

func (s *SchedulerClient) updateTriggers() {
	triggers, err := s.getTriggers()
	if err != nil {
		log.Error(err)
		return
	}
	for _, trigger := range triggers {
		key := trigger.GetTime().String()
		if _, ok := s.triggers[key]; !ok {
			t, err := ptypes.Timestamp(trigger.GetTime())
			if err != nil {
				log.Error(err)
				continue
			}
			scanID := uniqueScanID(t, trigger.GetName(), "triggered")
			log.Infof("Adding %v to the benchmark queue", scanID)

			s.benchmarkChan <- &v1.BenchmarkScanMetadata{
				ScanId:     scanID,
				Benchmark:  trigger.GetName(),
				ClusterIds: trigger.GetClusterIds(),
				Time:       trigger.GetTime(),
				Reason:     v1.BenchmarkReason_TRIGGERED,
			}
			s.triggers[key] = trigger
		}
	}
}

func (s *SchedulerClient) updateSchedules() {
	schedules, err := s.getSchedules()
	if err != nil {
		log.Error(err)
		return
	}
	currentSchedules := mapset.NewSet()
	for _, schedule := range schedules {
		oldSchedule, exists := s.schedules[schedule.Name]
		// If the schedule doesn't exist or has been updated then start scheduling for it
		if !exists || protoconv.CompareProtoTimestamps(schedule.LastUpdated, oldSchedule.LastUpdated) != 0 {
			nextTime, err := nextScheduledTime(schedule)
			if err != nil {
				log.Error(err)
				continue
			}
			s.schedules[schedule.Name] = &scheduleMetadata{
				BenchmarkSchedule: schedule,
				NextScanTime:      nextTime,
			}
		}
		currentSchedules.Add(schedule.Name)
	}

	for name := range s.schedules {
		if !currentSchedules.Contains(name) {
			delete(s.schedules, name)
		}
	}
	// Run through the schedules and run their benchmarks if they have expired
	now := time.Now()
	for benchmarkName, scheduleMetadata := range s.schedules {
		nextScanTime := scheduleMetadata.NextScanTime
		protoTime, err := ptypes.TimestampProto(nextScanTime)
		if err != nil {
			log.Errorf("Could not convert golang time %v to proto time", nextScanTime)
		}
		if nextScanTime.Before(now) {
			scanID := uniqueScanID(nextScanTime, benchmarkName, "scheduled")
			// Add benchmark to the queue to be scheduled
			log.Infof("Adding %v to the benchmark queue", scanID)
			s.benchmarkChan <- &v1.BenchmarkScanMetadata{
				ScanId:     scanID,
				Benchmark:  benchmarkName,
				ClusterIds: scheduleMetadata.GetClusterIds(),
				Time:       protoTime,
				Reason:     v1.BenchmarkReason_SCHEDULED,
			}

			// Schedule the scan next week
			scheduleMetadata.NextScanTime = nextScanTime.Add(7 * 24 * time.Hour)
			log.Infof("Benchmark %v is scheduled to run next week at %v", scheduleMetadata.GetName(), scheduleMetadata.NextScanTime.Format(time.RFC3339))
		}
	}
}

func (s *SchedulerClient) launchBenchmark(scan *v1.BenchmarkScanMetadata) error {
	conn, err := clientconn.GRPCConnection(s.centralEndpoint)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	benchmark, err := v1.NewBenchmarkServiceClient(conn).GetBenchmark(ctx, &v1.GetBenchmarkRequest{Name: scan.GetBenchmark()})
	if err != nil {
		return err
	}
	scan.Checks = benchmark.GetChecks()
	// Send report back to master (may need retries, saying that we are trying to launch)
	ctx, cancel = context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	_, err = v1.NewBenchmarkScanServiceClient(conn).PostBenchmarkScan(ctx, scan)
	if err != nil {
		return err
	}
	if err := s.Launch(scan); err != nil {
		return fmt.Errorf("Error launching benchmark: %s", err)
	}
	return nil
}

func uniqueScanID(t time.Time, benchmarkName, triggerType string) string {
	return uuid.NewV5(staticIDNamespace, t.Format(time.RFC3339)+benchmarkName+triggerType).String()
}

// Start runs the scheduler
func (s *SchedulerClient) Start() {
	// Initialize triggers that have results from this sensor
	s.initializeTriggers()
	for {
		select {
		case <-s.updateTicker.C:
			// Update the schedules and schedule any that need to be run
			s.updateSchedules()
			// Update the triggers and schedule any ones that need to be run
			s.updateTriggers()
		case scan := <-s.benchmarkChan:
			log.Infof("Launching benchmark %v for scan id '%s'", scan.GetBenchmark(), scan.GetScanId())
			if err := s.launchBenchmark(scan); err != nil {
				log.Errorf("Error launching benchmark %v with scan id '%v': %+v", scan.GetBenchmark(), scan.GetScanId(), err)
			}
		case <-s.done:
			s.started = false
			return
		}
	}
}

// Stop stops the scheduler client from triggering any more jobs.
func (s *SchedulerClient) Stop() {
	s.done <- struct{}{}

	// TODO(cg): Also stop any launched benchmark.
}
