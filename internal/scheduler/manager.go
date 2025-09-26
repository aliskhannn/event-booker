package scheduler

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/wb-go/wbf/zlog"
)

// Job interface defines the structure for any job that will be scheduled.
type Job interface {
	// Name returns the name of the job.
	Name() string

	// Schedule returns the cron schedule for the job.
	Schedule() string

	// Run executes the job logic.
	Run(ctx context.Context) error
}

// JobManager manages all scheduled jobs.
type JobManager struct {
	cron    *cron.Cron
	jobs    []Job
	context context.Context
}

// NewJobManager creates a new JobManager instance.
func NewJobManager(ctx context.Context) *JobManager {
	return &JobManager{
		cron:    cron.New(cron.WithSeconds()), // enable seconds precision
		jobs:    []Job{},
		context: ctx,
	}
}

// RegisterJob adds a job to the job manager.
func (jm *JobManager) RegisterJob(job Job) {
	jm.jobs = append(jm.jobs, job)
}

// StartScheduler starts the cron scheduler to execute jobs at their scheduled times.
func (jm *JobManager) StartScheduler() {
	for _, job := range jm.jobs {
		schedule := job.Schedule()

		if _, err := jm.cron.AddFunc(schedule, func() {
			if err := job.Run(jm.context); err != nil {
				zlog.Logger.Error().Err(err).Str("job", job.Name()).Msg("failed to execute job")
			} else {
				zlog.Logger.Printf("job %s executed successfully", job.Name())
			}
		}); err != nil {
			zlog.Logger.Error().Err(err).Str("job", job.Name()).Msg("failed to schedule job")
		}
	}

	jm.cron.Start()
}
