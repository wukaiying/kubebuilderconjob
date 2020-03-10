/*

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

package controllers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/go-logr/logr"
	"github.com/robfig/cron"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1 "wukaiying/kubebuilderconjob/api/v1"
)

//cronjob 的结构就是schedule 元素 + job 结构体组合成cronjob
// CronJobReconciler reconciles a CronJob object
//这个结构体里面可以添加子需要的字段
type CronJobReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Clock
}

type realClock struct{}

func (_ realClock) Now() time.Time { return time.Now() }

// clock knows how to get the current time.
// It can be used to fake out timing for testing.
type Clock interface {
	Now() time.Time
}

var (
	scheduledTimeAnnotation = "batch.tutorial.kubebuilder.io/scheduled-at"
)

// +kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=cronjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get
func (r *CronJobReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cronjob", req.NamespacedName)

	var cronJob batchv1.CronJob
	if err := r.Get(ctx, req.NamespacedName, &cronJob); err != nil { //获取集群中cronjob,如果没有找到，则直接return  ctrl.Result{}，不入队列
		log.Error(err, "can not fetch cronjob")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	//列出所有和conjob相关的childjobs,就是在jobtemplate中定义的多个job--> spec.containers
	var childJobs kbatch.JobList
	if err := r.List(ctx, &childJobs, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err != nil {
		log.Error(err, "unable to list child Jobs")
		return ctrl.Result{}, err
	}

	//找出正在运行的，成功的，失败的job,通过job的condition 的type来判断
	var activeJobs []*kbatch.Job
	var successJobs []*kbatch.Job
	var failedJobs []*kbatch.Job
	var mostRecentTime *time.Time // find the last run so we can update the status

	isJobFinished := func(job *kbatch.Job) (bool, kbatch.JobConditionType) {
		for _, c := range job.Status.Conditions {
			if (c.Type == kbatch.JobComplete || c.Type == kbatch.JobFailed) && c.Status == corev1.ConditionTrue {
				return true, c.Type
			}
		}
		return false, ""
	}
	getScheduledTimeForJob := func(job *kbatch.Job) (*time.Time, error) {
		timeRaw := job.Annotations[scheduledTimeAnnotation]
		if len(timeRaw) == 0 { //没有设置schedule time
			return nil, nil
		}
		timeParsed, err := time.Parse(time.RFC3339, timeRaw)
		if err != nil {
			return nil, err
		}
		return &timeParsed, nil
	}

	for _, job := range childJobs.Items {
		_, finishedType := isJobFinished(&job)
		switch finishedType {
		case "": //说明正在运行
			activeJobs = append(activeJobs, &job)
		case kbatch.JobComplete:
			successJobs = append(successJobs, &job)
		case kbatch.JobFailed:
			failedJobs = append(failedJobs, &job)
		}

		scheduledTimeForJob, err := getScheduledTimeForJob(&job)
		if err != nil {
			log.Error(err, "unable to parse schedule time for child job", "job", &job)
			continue
		}
		if scheduledTimeForJob != nil { //更新mostRecentTime,如果mostRecentTime为空，则把scheduledTimeForJob给他，如果mostRecentTime时间小于scheduledTimeForJob则更新mostRecentTime为scheduledTimeForJob
			if mostRecentTime == nil {
				mostRecentTime = scheduledTimeForJob
			} else if mostRecentTime.Before(*scheduledTimeForJob) {
				mostRecentTime = scheduledTimeForJob
			}
		}

		//mostRecentTime每次reconcil都会被更新，如果mostRecentTime不为空则把它放到cronJob.Status.LastScheduleTime
		if mostRecentTime != nil {
			cronJob.Status.LastScheduleTime = &metav1.Time{Time: *mostRecentTime}
		} else {
			cronJob.Status.LastScheduleTime = nil
		}

		cronJob.Status.Active = nil
		for _, activeJob := range activeJobs {
			jobRef, err := reference.GetReference(r.Scheme, activeJob)
			if err != nil {
				log.Error(err, "unable to make reference to active job", "job", activeJob)
				continue
			}
			cronJob.Status.Active = append(cronJob.Status.Active, *jobRef)
			log.V(1).Info("job count", "active jobs", len(activeJobs), "successful jobs", len(successJobs), "failed jobs", len(failedJobs))
		}
	}

	//此时我们已经拼出了一个conjob结构体，里面包含了调度时间和statsus，我们使用update方法来更新cronjob内容
	//我们可以只更新status这部分内容
	if err := r.Status().Update(ctx, &cronJob); err != nil {
		log.Error(err, "unable to update CronJob status")
		return ctrl.Result{}, err
	}

	//失败的job,重复执行几次还不成功就删除,这是cronjob中spec的一个参数
	if cronJob.Spec.FailedJobsHistoryLimit != nil {
		//按照start time进行排序
		sort.Slice(failedJobs, func(i, j int) bool {
			if failedJobs[i].Status.StartTime == nil {
				return failedJobs[j].Status.StartTime != nil
			}
			return failedJobs[i].Status.StartTime.Before(failedJobs[j].Status.StartTime)
		})

		for i, failedJob := range failedJobs {
			if int32(i) <= int32(len(failedJobs))-*cronJob.Spec.FailedJobsHistoryLimit { //也就是failedjob数量没有达到limit,break
				break
			}
			///否则进行删除
			if err := r.Delete(ctx, failedJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
				log.Error(err, "unable to delete old failed job", "job", failedJob)
			} else {
				log.V(0).Info("deleted old failed job", "job", failedJob)
			}
		}
	}

	//同理删除多余的success job
	if cronJob.Spec.SuccessfulJobsHistoryLimit != nil {
		//按照starttime进行排序
		sort.Slice(successJobs, func(i, j int) bool {
			if successJobs[i].Status.StartTime == nil {
				return successJobs[j].Status.StartTime != nil
			}
			return successJobs[i].Status.StartTime.Before(successJobs[j].Status.StartTime)
		})

		for i, successJob := range successJobs {
			if int32(i) <= int32(len(successJobs))-*cronJob.Spec.SuccessfulJobsHistoryLimit { //也就是failedjob数量没有达到limit,break
				break
			}
			///否则进行删除
			if err := r.Delete(ctx, successJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
				log.Error(err, "unable to delete old failed job", "job", successJob)
			} else {
				log.V(0).Info("deleted old failed job", "job", successJob)
			}
		}
	}

	//判断spec 参数suspend，如果suspend我们不执行任何job
	if cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend {
		log.V(1).Info("cronjob suspended, skipping")
		return ctrl.Result{}, nil
	}

	getNextSchedule := func(cronJob *batchv1.CronJob, now time.Time) (lastMissed time.Time, next time.Time, err error) {
		sched, err := cron.ParseStandard(cronJob.Spec.Schedule)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("Unparseable schedule %q: %v", cronJob.Spec.Schedule, err)
		}
		// for optimization purposes, cheat a bit and start from our last observed run time
		// we could reconstitute this here, but there's not much point, since we've
		// just updated it.
		var earliestTime time.Time
		if cronJob.Status.LastScheduleTime != nil {
			earliestTime = cronJob.Status.LastScheduleTime.Time
		} else {
			earliestTime = cronJob.ObjectMeta.CreationTimestamp.Time
		}
		if cronJob.Spec.StartingDeadlineSeconds != nil {
			// controller is not going to schedule anything below this point
			schedulingDeadline := now.Add(-time.Second * time.Duration(*cronJob.Spec.StartingDeadlineSeconds))

			if schedulingDeadline.After(earliestTime) {
				earliestTime = schedulingDeadline
			}
			if earliestTime.After(now) {
				return time.Time{}, sched.Next(now), nil
			}
			starts := 0
			for t := sched.Next(earliestTime); !t.After(now); t = sched.Next(t) {
				lastMissed = t
				// An object might miss several starts. For example, if
				// controller gets wedged on Friday at 5:01pm when everyone has
				// gone home, and someone comes in on Tuesday AM and discovers
				// the problem and restarts the controller, then all the hourly
				// jobs, more than 80 of them for one hourly scheduledJob, should
				// all start running with no further intervention (if the scheduledJob
				// allows concurrency and late starts).
				//
				// However, if there is a bug somewhere, or incorrect clock
				// on controller's server or apiservers (for setting creationTimestamp)
				// then there could be so many missed start times (it could be off
				// by decades or more), that it would eat up all the CPU and memory
				// of this controller. In that case, we want to not try to list
				// all the missed start times.
				starts++
				if starts > 100 {
					// We can't get the most recent times so just return an empty slice
					return time.Time{}, time.Time{}, fmt.Errorf("too many missed start times (> 100). Set or decrease .spec.startingDeadlineSeconds or check clock skew")
				}
			}
		}
		return lastMissed, sched.Next(now), nil
	}
	// figure out the next times that we need to create
	// jobs at (or anything we missed).
	missedRun, nextRun, err := getNextSchedule(&cronJob, r.Now())
	if err != nil {
		log.Error(err, "unable to figure out CronJob schedule")
		// we don't really care about requeuing until we get an update that
		// fixes the schedule, so don't return an error
		return ctrl.Result{}, nil
	}

	//ctrl.Result RequeueAfter如果有值，不管是不是true,都会重新入队列
	scheduledResult := ctrl.Result{RequeueAfter: nextRun.Sub(r.Now())}
	log = log.WithValues("now", r.Now(), "next run", nextRun)
	//我们需要根据下次调度时间来计算什么时候真正被调度运行
	if missedRun.IsZero() {
		log.V(1).Info("no upcoming scheduled times, sleeping until next")
		return scheduledResult, nil
	}

	// make sure we're not too late to start the run
	log = log.WithValues("current run", missedRun)
	tooLate := false
	if cronJob.Spec.StartingDeadlineSeconds != nil {
		tooLate = missedRun.Add(time.Duration(*cronJob.Spec.StartingDeadlineSeconds) * time.Second).Before(r.Now())
	}
	if tooLate {
		log.V(1).Info("missed starting deadline for last run, sleeping till next")
		// TODO(directxman12): events
		return scheduledResult, nil
	}

	// figure out how to run this job -- concurrency policy might forbid us from running
	// multiple at the same time...
	if cronJob.Spec.ConcurrencyPolicy == batchv1.ForbidConcurrent && len(activeJobs) > 0 {
		log.V(1).Info("concurrency policy blocks concurrent runs, skipping", "num active", len(activeJobs))
		return scheduledResult, nil
	}

	// ...or instruct us to replace existing ones...
	if cronJob.Spec.ConcurrencyPolicy == batchv1.ReplaceConcurrent {
		for _, activeJob := range activeJobs {
			// we don't care if the job was already deleted
			if err := r.Delete(ctx, activeJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
				log.Error(err, "unable to delete active job", "job", activeJob)
				return ctrl.Result{}, err
			}
		}
	}

	//组装job,并apply到集群中,相当于把数据Deepcopy到job中，我们写定时调度逻辑来执行它
	constructJobForCronJob := func(cronJob *batchv1.CronJob, scheduledTime time.Time) (*kbatch.Job, error) {
		// We want job names for a given nominal start time to have a deterministic name to avoid the same job being created twice
		name := fmt.Sprintf("%s-%d", cronJob.Name, scheduledTime.Unix())

		job := &kbatch.Job{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
				Name:        name,
				Namespace:   cronJob.Namespace,
			},
			Spec: *cronJob.Spec.JobTemplate.Spec.DeepCopy(),
		}
		for k, v := range cronJob.Spec.JobTemplate.Annotations {
			job.Annotations[k] = v
		}
		job.Annotations[scheduledTimeAnnotation] = scheduledTime.Format(time.RFC3339)
		for k, v := range cronJob.Spec.JobTemplate.Labels {
			job.Labels[k] = v
		}
		if err := ctrl.SetControllerReference(cronJob, job, r.Scheme); err != nil {
			return nil, err
		}

		return job, nil
	}

	// actually make the job...
	job, err := constructJobForCronJob(&cronJob, missedRun)
	if err != nil {
		log.Error(err, "unable to construct job from template")
		// don't bother requeuing until we get a change to the spec
		return scheduledResult, nil
	}

	// ...and create it on the cluster
	if err := r.Create(ctx, job); err != nil {
		log.Error(err, "unable to create Job for CronJob", "job", job)
		return ctrl.Result{}, err
	}
	log.V(1).Info("created Job for CronJob run", "job", job)

	return scheduledResult, nil
}

var (
	jobOwnerKey = ".metadata.controller"
	apiGVStr    = batchv1.GroupVersion.String()
)

func (r *CronJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// set up a real clock, since we're not in a test
	if r.Clock == nil {
		r.Clock = realClock{}
	}

	if err := mgr.GetFieldIndexer().IndexField(&kbatch.Job{}, jobOwnerKey, func(rawObj runtime.Object) []string {
		// grab the job object, extract the owner...
		job := rawObj.(*kbatch.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		// ...make sure it's a CronJob...
		if owner.APIVersion != apiGVStr || owner.Kind != "CronJob" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.CronJob{}).
		Complete(r)
}
