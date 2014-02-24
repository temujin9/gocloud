package cloudformation

import (
	"fmt"
	"sort"
	"time"

	"github.com/dynport/dgtk/cli"
	"github.com/dynport/gocli"
	"github.com/dynport/gocloud/aws/cloudformation"
)

type StacksList struct {
	IncludeDeleted bool `cli:"opt --deleted"`
	Full           bool `cli:"opt --full"`
}

var client = cloudformation.NewFromEnv()

func (list *StacksList) Run() error {
	rsp, e := client.ListStacks(nil)
	if e != nil {
		return e
	}
	t := gocli.NewTable()
	for _, s := range rsp.ListStacksResult.Stacks {
		if !list.IncludeDeleted && s.StackStatus == "DELETE_COMPLETE" {
			continue
		}
		parts := []interface{}{s.StackName, s.StackStatus, s.CreationTime.Format(time.RFC3339)}
		if list.Full {
			parts = append(parts, s.StackId)
		}
		t.Add(parts...)
	}
	fmt.Println(t)
	return nil
}

type StackResources struct {
	Name string `cli:"arg required"`
}

func (r *StackResources) Run() error {
	rsp, e := client.DescribeStackResources(cloudformation.DescribeStackResourcesParameters{
		StackName: r.Name,
	})
	if e != nil {
		return e
	}
	t := gocli.NewTable()
	for _, r := range rsp.DescribeStackResourcesResult.StackResources {
		t.Add(r.LogicalResourceId, r.PhysicalResourceId)
	}
	fmt.Println(t)
	return nil
}

type StacksWatch struct {
	Name string `cli:"arg required"`
}

type StackEventsList []*cloudformation.StackEvent

func (list StackEventsList) Len() int {
	return len(list)
}

func (list StackEventsList) Swap(a, b int) {
	list[a], list[b] = list[b], list[a]
}

func (list StackEventsList) Less(a, b int) bool {
	return list[a].Timestamp.Before(list[b].Timestamp)
}

var (
	ReasonUserInitiated             = "User Initiated"
	ReasonResourceCreationInitiated = "Resource creation Initiated"
)

func (s *StacksWatch) Run() error {
	rsp, e := client.DescribeStackEvents(&cloudformation.DescribeStackEventsParameters{StackName: s.Name})
	if e != nil {
		return e
	}
	max := 0
	events := StackEventsList(rsp.DescribeStackEventsResult.StackEvents)
	sort.Sort(events)
	for _, e := range events {
		ph := e.PhysicalResourceId
		fmt.Printf("%s %-24s %-32s %s\n", e.Timestamp.Format(time.RFC3339), maxLen(e.LogicalResourceId, 24), maxLen(ph, 32), e.ResourceStatus)
		switch e.ResourceStatusReason {
		case "", ReasonUserInitiated, ReasonResourceCreationInitiated:
			//
		default:
			fmt.Printf("%20s %s\n", "", gocli.Red(e.ResourceStatusReason))
		}
	}
	fmt.Printf("max: %d", max)
	return nil
}

func maxLen(s string, i int) string {
	if len(s) < i {
		return s
	}
	return s[0:i]
}

func Register(router *cli.Router) {
	router.Register("aws/cloudformation/stacks/list", &StacksList{}, "List Cloudformation stacks")
	router.Register("aws/cloudformation/stacks/watch", &StacksWatch{}, "Watch Stacks")
	router.Register("aws/cloudformation/stacks/resources", &StackResources{}, "Describe Stack Resources")
}