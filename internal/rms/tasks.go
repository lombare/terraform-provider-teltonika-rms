package rms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
)

// TaskGroup mirrors the /devices/tasks/groups/{id} envelope (metadata only). Task bodies
// are fetched from /devices/tasks?group_id=… — see GetTaskGroup for the composed read.
type TaskGroup struct {
	ID           json.Number     `json:"id"`
	Name         string          `json:"name,omitempty"`
	CompanyID    json.Number     `json:"company_id,omitempty"`
	SuccessTagID []json.Number   `json:"success_tag_id,omitempty"`
	FailedTagID  []json.Number   `json:"failed_tag_id,omitempty"`
	CreatedAt    Timestamp       `json:"created_at,omitempty"`
	UpdatedAt    Timestamp       `json:"updated_at,omitempty"`
	Tasks        []Task          `json:"-"`
	Raw          json.RawMessage `json:"-"`
}

// Task is one entry from /devices/tasks (filterable by group_id).
type Task struct {
	ID            json.Number `json:"id,omitempty"`
	Name          string      `json:"name"`
	Order         int         `json:"order"`
	Type          string      `json:"type"`
	Timeout       int         `json:"timeout"`
	StopOnFailure bool        `json:"stop_on_failure"`
	Data          TaskData    `json:"data"`
	GroupID       json.Number `json:"group_id,omitempty"`
	GroupName     string      `json:"group_name,omitempty"`
	CompanyID     json.Number `json:"company_id,omitempty"`
	CreatedAt     Timestamp   `json:"created_at,omitempty"`
	UpdatedAt     Timestamp   `json:"updated_at,omitempty"`
}

// TaskData carries the per-type payload for a Task. Only the fields matching the task's
// Type are meaningful; the rest are zero-valued and omitted via omitempty on marshal.
type TaskData struct {
	Command         string `json:"command,omitempty"`
	AcceptableCodes []int  `json:"acceptable_codes,omitempty"`
	FileID          int64  `json:"file_id,omitempty"`
	TargetPath      string `json:"target_path,omitempty"`
	File            string `json:"file,omitempty"`
}

// TaskGroupCreate is the POST /devices/tasks/groups body.
type TaskGroupCreate struct {
	Name         string  `json:"name"`
	CompanyID    int64   `json:"company_id,omitempty"`
	Tasks        []Task  `json:"tasks"`
	SuccessTagID []int64 `json:"success_tag_id,omitempty"`
	FailedTagID  []int64 `json:"failed_tag_id,omitempty"`
}

// TaskGroupUpdate is the PUT /devices/tasks/groups/{id} body.
type TaskGroupUpdate struct {
	Name  string `json:"name,omitempty"`
	Tasks []Task `json:"tasks,omitempty"`
}

// ListTaskGroups paginates /devices/tasks/groups.
func (c *Client) ListTaskGroups(ctx context.Context, q url.Values) ([]TaskGroup, error) {
	return ListAll[TaskGroup](ctx, c, "devices/tasks/groups", q)
}

// GetTaskGroup composes the metadata call with the group's task list.
func (c *Client) GetTaskGroup(ctx context.Context, id string) (*TaskGroup, error) {
	var env Envelope[TaskGroup]
	if err := c.Get(ctx, "devices/tasks/groups/"+id, nil, &env); err != nil {
		return nil, err
	}
	group := env.Data
	tasks, err := c.ListTasks(ctx, url.Values{"group_id": []string{id}})
	if err != nil {
		return nil, fmt.Errorf("rms: fetch tasks for group %s: %w", id, err)
	}
	group.Tasks = tasks
	return &group, nil
}

func (c *Client) CreateTaskGroup(ctx context.Context, in TaskGroupCreate) (*TaskGroup, error) {
	var env Envelope[TaskGroup]
	if err := c.Post(ctx, "devices/tasks/groups", in, &env); err != nil {
		return nil, err
	}
	if env.Data.ID.String() == "" {
		return nil, fmt.Errorf("rms: task group create returned empty id")
	}
	return c.GetTaskGroup(ctx, env.Data.ID.String())
}

func (c *Client) UpdateTaskGroup(ctx context.Context, id string, in TaskGroupUpdate) error {
	return c.Put(ctx, "devices/tasks/groups/"+id, in, nil)
}

func (c *Client) DeleteTaskGroup(ctx context.Context, id string) error {
	return c.Delete(ctx, "devices/tasks/groups/"+id, nil)
}

// ListTasks paginates /devices/tasks; use group_id / company_id / type in q to filter.
func (c *Client) ListTasks(ctx context.Context, q url.Values) ([]Task, error) {
	return ListAll[Task](ctx, c, "devices/tasks", q)
}

// SortTasks returns the tasks ordered by their `order` field.
func SortTasks(tasks []Task) []Task {
	out := append([]Task(nil), tasks...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	return out
}
