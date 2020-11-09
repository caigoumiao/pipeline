package pipeline

import (
	"context"
)

type (
	// 工作节点的处理方法
	//WorkerFunc interface {
	//    Process(ctx context.Context, in *rawData) (out *rawData, err error)
	//}
	WorkerFunc func(ctx context.Context, in *rawData) (out *rawData, err error)
	// 划分节点的处理方法
	//DividerFunc interface {
	//    Divide(ctx context.Context, in *rawData) (out []*rawData, err error)
	//}
	DividerFunc func(ctx context.Context, in *rawData) (out []*rawData, err error)
	// 合并节点的处理方法
	//MergerFunc interface {
	//    Merge(ctx context.Context, in []*rawData) (out *rawData, err error)
	//}
	MergerFunc func(ctx context.Context, in []*rawData) (out *rawData, err error)
	// 判断节点的处理方法
	//JudgerFunc interface {
	//    Judge(ctx context.Context, in *rawData) (pipeIndex int)
	//}
	JudgerFunc func(ctx context.Context, in *rawData) (pipeIndex int)
)

type HandlerStatus int

const (
	HandlerStatusSuccess HandlerStatus = 0
	HandlerStatusError   HandlerStatus = 1
	HandlerStatusTimeout HandlerStatus = 2
)

// 节点之间传递的数据
type rawData struct {
	Status HandlerStatus
	Data   interface{}
	Meta   map[string]interface{}
}

type NodeTyp string

const (
	NodeTypHead    = "head"
	NodeTypWorker  = "worker"
	NodeTypDivider = "divider"
	NodeTypMerger  = "merger"
	NodeTypJudger  = "judger"
	NodeTypTail    = "tail"
)

type (
	Node struct {
		Typ      NodeTyp
		nodeName string
		actionId string
		Next     []*Node
	}
)
