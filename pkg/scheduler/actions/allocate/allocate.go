/*
Copyright 2017 The Kubernetes Authors.

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

package allocate

import (
	"github.com/golang/glog"

	"volcano.sh/volcano/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/util"
)

type allocateAction struct {
	ssn *framework.Session
}

func New() *allocateAction {
	return &allocateAction{}
}

func (alloc *allocateAction) Name() string {
	return "allocate"
}

func (alloc *allocateAction) Initialize() {}

func (alloc *allocateAction) Execute(ssn *framework.Session) {
	glog.V(3).Infof("Enter Allocate ...")
	defer glog.V(3).Infof("Leaving Allocate ...")

	// the allocation for pod may have many stages
	// 1. pick a namespace named N (using ssn.NamespaceOrderFn)
	// 2. pick a queue named Q from N (using ssn.QueueOrderFn)
	// 3. pick a job named J from Q (using ssn.JobOrderFn)
	// 4. pick a task T from J (using ssn.TaskOrderFn)
	// 5. use predicateFn to filter out node that T can not be allocated on.
	// 6. use ssn.NodeOrderFn to judge the best node and assign it to T

	namespaces := util.NewPriorityQueue(ssn.NamespaceOrderFn)

	// jobsMap is map[api.NamespaceName]map[api.QueueID]PriorityQueue(*api.JobInfo)
	// used to find job with highest priority in given queue and namespace
	jobsMap := map[api.NamespaceName]map[api.QueueID]*util.PriorityQueue{}

	for _, job := range ssn.Jobs {
		if job.PodGroup.Status.Phase == scheduling.PodGroupPending {
			continue
		}
		if vr := ssn.JobValid(job); vr != nil && !vr.Pass {
			glog.V(4).Infof("Job <%s/%s> Queue <%s> skip allocate, reason: %v, message %v", job.Namespace, job.Name, job.Queue, vr.Reason, vr.Message)
			continue
		}

		if _, found := ssn.Queues[job.Queue]; !found {
			glog.Warningf("Skip adding Job <%s/%s> because its queue %s is not found",
				job.Namespace, job.Name, job.Queue)
			continue
		}

		namespace := api.NamespaceName(job.Namespace)
		queueMap, found := jobsMap[namespace]
		if !found {
			namespaces.Push(namespace)

			queueMap = make(map[api.QueueID]*util.PriorityQueue)
			jobsMap[namespace] = queueMap
		}

		jobs, found := queueMap[job.Queue]
		if !found {
			jobs = util.NewPriorityQueue(ssn.JobOrderFn)
			queueMap[job.Queue] = jobs
		}

		glog.V(4).Infof("Added Job <%s/%s> into Queue <%s>", job.Namespace, job.Name, job.Queue)
		jobs.Push(job)
	}

	glog.V(3).Infof("Try to allocate resource to %d Namespaces", len(jobsMap))

	pendingTasks := map[api.JobID]*util.PriorityQueue{}

	allNodes := util.GetNodeList(ssn.Nodes)

	predicateFn := func(task *api.TaskInfo, node *api.NodeInfo) error {
		// Check for Resource Predicate
		// TODO: We could not allocate resource to task from both node.Idle and node.Releasing now,
		// after it is done, we could change the following compare to:
		// clonedNode := node.Idle.Clone()
		// if !task.InitResreq.LessEqual(clonedNode.Add(node.Releasing)) {
		//    ...
		// }
		if !task.InitResreq.LessEqual(node.Idle) && !task.InitResreq.LessEqual(node.Releasing) {
			return api.NewFitError(task, node, api.NodeResourceFitFailed)
		}

		return ssn.PredicateFn(task, node)
	}

	// To pick <namespace, queue> tuple for job, we choose to pick namespace firstly.
	// Because we believe that number of queues would less than namespaces in most case.
	// And, this action would make the resource usage among namespace balanced.
	for {
		if namespaces.Empty() {
			break
		}

		// pick namespace from namespaces PriorityQueue
		namespace := namespaces.Pop().(api.NamespaceName)

		queueInNamespace := jobsMap[namespace]

		// pick queue for given namespace
		//
		// This block use a algorithm with time complex O(n).
		// But at least PriorityQueue could not be used here,
		// because the allocation of job would change the priority of queue among all namespaces,
		// and the PriorityQueue have no ability to update priority for a special queue.
		var queue *api.QueueInfo
		for queueId := range queueInNamespace {
			currentQueue := ssn.Queues[queueId]
			if ssn.Overused(currentQueue) {
				glog.V(3).Infof("Namespace <%s> Queue <%s> is overused, ignore it.", namespace, currentQueue.Name)
				delete(queueInNamespace, queueId)
				continue
			}

			if queue == nil || ssn.QueueOrderFn(currentQueue, queue) {
				queue = currentQueue
			}
		}

		if queue == nil {
			glog.V(3).Infof("Namespace <%s> have no queue, skip it", namespace)
			continue
		}

		glog.V(3).Infof("Try to allocate resource to Jobs in Namespace <%s> Queue <%v>", namespace, queue.Name)

		jobs, found := queueInNamespace[queue.UID]
		if !found || jobs.Empty() {
			glog.V(4).Infof("Can not find jobs for queue %s.", queue.Name)
			continue
		}

		job := jobs.Pop().(*api.JobInfo)
		if _, found := pendingTasks[job.UID]; !found {
			tasks := util.NewPriorityQueue(ssn.TaskOrderFn)
			for _, task := range job.TaskStatusIndex[api.Pending] {
				// Skip BestEffort task in 'allocate' action.
				if task.Resreq.IsEmpty() {
					glog.V(4).Infof("Task <%v/%v> is BestEffort task, skip it.",
						task.Namespace, task.Name)
					continue
				}

				tasks.Push(task)
			}
			pendingTasks[job.UID] = tasks
		}
		tasks := pendingTasks[job.UID]

		glog.V(3).Infof("Try to allocate resource to %d tasks of Job <%v/%v>",
			tasks.Len(), job.Namespace, job.Name)

		stmt := ssn.Statement()

		for !tasks.Empty() {
			task := tasks.Pop().(*api.TaskInfo)

			glog.V(3).Infof("There are <%d> nodes for Job <%v/%v>",
				len(ssn.Nodes), job.Namespace, job.Name)

			//any task that doesn't fit will be the last processed
			//within this loop context so any existing contents of
			//NodesFitDelta are for tasks that eventually did fit on a
			//node
			if len(job.NodesFitDelta) > 0 {
				job.NodesFitDelta = make(api.NodeResourceMap)
			}

			predicateNodes, fitErrors := util.PredicateNodes(task, allNodes, predicateFn)
			if len(predicateNodes) == 0 {
				job.NodesFitErrors[task.UID] = fitErrors
				break
			}

			nodeScores := util.PrioritizeNodes(task, predicateNodes, ssn.BatchNodeOrderFn, ssn.NodeOrderMapFn, ssn.NodeOrderReduceFn)
			//Instead of choosing the best node for one tasks we choose the best K nodes for N tasks at a time
			//specify the number of nodes based on the number of tasks
			nodes := util.SelectTopNodes(nodeScores, tasks.Len()+1)
			for i, node := range nodes {
				// Allocate idle resource to the task.
				if task.InitResreq.LessEqual(node.Idle) {
					glog.V(3).Infof("Binding Task <%v/%v> to node <%v>",
						task.Namespace, task.Name, node.Name)
					if err := stmt.Allocate(task, node.Name); err != nil {
						glog.Errorf("Failed to bind Task %v on %v in Session %v, err: %v",
							task.UID, node.Name, ssn.UID, err)
					}
				} else {
					//store information about missing resources
					job.NodesFitDelta[node.Name] = node.Idle.Clone()
					job.NodesFitDelta[node.Name].FitDelta(task.InitResreq)
					glog.V(3).Infof("Predicates failed for task <%s/%s> on node <%s> with limited resources",
						task.Namespace, task.Name, node.Name)

					// Allocate releasing resource to the task if any.
					if task.InitResreq.LessEqual(node.Releasing) {
						glog.V(3).Infof("Pipelining Task <%v/%v> to node <%v> for <%v> on <%v>",
							task.Namespace, task.Name, node.Name, task.InitResreq, node.Releasing)
						if err := stmt.Pipeline(task, node.Name); err != nil {
							glog.Errorf("Failed to pipeline Task %v on %v",
								task.UID, node.Name)
						}
					}
				}
				//If the task queue is empty, stop resource allocation
				if tasks.Empty() {
					break
				}
				//If there are still tasks and nodes, continue to allocate
				if i < len(nodes)-1 {
					task = tasks.Pop().(*api.TaskInfo)
				}
			}
		}

		if ssn.JobReady(job) {
			stmt.Commit()
		} else {
			stmt.Discard()
		}

		// Added Namespace back until no job in Namespace.
		namespaces.Push(namespace)
	}
}

func (alloc *allocateAction) UnInitialize() {}
