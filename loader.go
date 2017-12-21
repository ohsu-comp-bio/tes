package tes

import (
	"context"
	"sync"
)

// BulkClient makes easy it call client.GetTask in parallel.
type BulkClient struct {
	*Client
	// Number of requests to make in parallel.
	// Defaults to 5.
	Threads int
}

// GetTasksByID is a simple wrapper around GetTasks that automatically
// creates the []*GetTaskRequest with the same view for each given task ID.
func (bc *BulkClient) GetTasksByID(ctx context.Context, ids []string, view TaskView) []GetTasksResponse {
	reqs := make([]*GetTaskRequest, len(ids), len(ids))
	for i, id := range ids {
		reqs[i] = &GetTaskRequest{
			Id:   id,
			View: view,
		}
	}
	return bc.GetTasks(ctx, reqs)
}

// GetTasks calls client.GetTask in parallel.
func (bc *BulkClient) GetTasks(ctx context.Context, reqs []*GetTaskRequest) []GetTasksResponse {
	var wg sync.WaitGroup
	results := make([]GetTasksResponse, len(reqs), len(reqs))

	threads := 5
	if bc.Threads > 0 {
		threads = bc.Threads
	}

	resch := make(chan *GetTasksResponse)

	go func() {
		for i, req := range reqs {
			res := results[i]
			res.req = req
			resch <- &res
		}
		close(resch)
	}()

	// start a worker threads.
	wg.Add(threads)
	for i := 0; i < threads; i++ {
		go func() {
			for r := range resch {
				task, err := bc.Client.GetTask(ctx, r.req)
				r.Task = task
				r.Error = err
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
	}()

	return results
}

// GetTasksResponse holds the response (Task + error) for each request
// in a call to BulkClient.GetTasks.
type GetTasksResponse struct {
	*Task
	Error error
	req   *GetTaskRequest
}
