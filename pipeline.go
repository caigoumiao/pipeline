package pipeline

import (
	"context"
	"errors"
	"fmt"
)

type Manager struct {
	nodes          map[string]*Node
	edges          [][]string
	actionMap      map[string]interface{}
	inEdgeOfMerger map[string]int
}

var (
	ErrorsNodeNameDuplicate      = errors.New("node name is duplicate")
	ErrorsEdgesNotSetVirtualHead = errors.New("edges doesn't set virtual head")
	ErrorsNodesOrEdgesEmpty      = errors.New("edges or nodes is empty")
	ErrorsNodeNil                = errors.New("node is nil")
	ErrorsCannotReachTail        = errors.New("pipeline cannot reach tail")
)

func NewManager() *Manager {
	return &Manager{
		nodes:          make(map[string]*Node),
		edges:          nil,
		actionMap:      make(map[string]interface{}),
		inEdgeOfMerger: make(map[string]int),
	}
}

// 添加一个工作节点
func (m *Manager) AddWorkerNode(name string, f func(ctx context.Context, in *rawData) (out *rawData, err error)) error {
	if _, ok := m.nodes[name]; ok {
		return ErrorsNodeNameDuplicate
	}
	actionId := fmt.Sprintf("worker-%d", len(m.actionMap)+1)
	m.actionMap[actionId] = WorkerFunc(f)
	m.nodes[name] = &Node{
		Typ:      NodeTypWorker,
		actionId: actionId,
		nodeName: name,
	}
	return nil
}

// 添加一个分裂节点
func (m *Manager) AddDividerNode(name string, f func(ctx context.Context, in *rawData) (out []*rawData, err error)) error {
	if _, ok := m.nodes[name]; ok {
		return ErrorsNodeNameDuplicate
	}
	actionId := fmt.Sprintf("divider-%d", len(m.actionMap)+1)
	m.actionMap[actionId] = DividerFunc(f)
	m.nodes[name] = &Node{
		Typ:      NodeTypDivider,
		actionId: actionId,
		nodeName: name,
	}
	return nil
}

// 添加一个合并节点
func (m *Manager) AddMergerNode(name string, f func(ctx context.Context, in []*rawData) (out *rawData, err error)) error {
	if _, ok := m.nodes[name]; ok {
		return ErrorsNodeNameDuplicate
	}
	actionId := fmt.Sprintf("merger-%d", len(m.actionMap)+1)
	m.actionMap[actionId] = MergerFunc(f)
	m.nodes[name] = &Node{
		Typ:      NodeTypMerger,
		nodeName: name,
		actionId: actionId,
	}
	return nil
}

// 添加一个判断节点
func (m *Manager) AddJudgerNode(name string, f func(ctx context.Context, in *rawData) (pipeIndex int)) error {
	if _, ok := m.nodes[name]; ok {
		return ErrorsNodeNameDuplicate
	}
	actionId := fmt.Sprintf("judger-%d", len(m.actionMap)+1)
	m.actionMap[actionId] = JudgerFunc(f)
	m.nodes[name] = &Node{
		Typ:      NodeTypJudger,
		nodeName: name,
		actionId: actionId,
	}
	return nil
}

const (
	headNodeName = "head000"
	tailNodeName = "tail111"
)

func (m *Manager) BuildPipeline(e [][]string) (err error) {
	m.edges = e
	if err = m.connectNodes(); err != nil {
		return
	}
	if err = m.validate(); err != nil {
		return
	}
	m.calInEdgeOfMerger()
	return
}

// 将节点连成链表
func (m *Manager) connectNodes() error {
	if len(m.edges) == 0 || len(m.nodes) == 0 {
		return ErrorsNodesOrEdgesEmpty
	}
	// 添加虚拟头、尾节点
	m.nodes[headNodeName] = &Node{
		Typ:      NodeTypHead,
		nodeName: headNodeName,
	}
	m.nodes[tailNodeName] = &Node{
		Typ:      NodeTypTail,
		nodeName: tailNodeName,
	}
	// 尝试连接节点
	for _, edge := range m.edges {
		frontNode := m.nodes[edge[0]]
		if frontNode == nil || frontNode.Typ == NodeTypTail {
			continue
		}
		forwardNode := m.nodes[edge[1]]
		if forwardNode == nil {
			continue
		}
		frontNode.Next = append(frontNode.Next, forwardNode)
	}
	return nil
}

// todo:
// 检查节点以及连接的正确性
// 检查标准如下：
// 1、分裂节点的子节点不能为空，最后一定在合并节点汇合
// 2、edges 数组也需要检查
// 3、
func (m *Manager) validate() error {
	return nil
}

// 计算每个合并节点的入度
func (m *Manager) calInEdgeOfMerger() {
	for i := 0; i < len(m.edges); i++ {
		if m.nodes[m.edges[i][1]].Typ == NodeTypMerger {
			m.inEdgeOfMerger[m.edges[i][1]]++
		}
	}
}

// 流水线执行需要用到的结构体
type nodeDataWrapper struct {
	node *Node
	in   *rawData
}

// 执行整个流水线
func (m *Manager) Handle(in *rawData) (out *rawData, err error) {
	head := m.nodes[headNodeName]
	p := head.Next[0]
	mergerNodeInDataMap := make(map[string][]*rawData)
	var queue []*nodeDataWrapper
	queue = append(queue, &nodeDataWrapper{
		node: p,
		in:   in,
	})
	for len(queue) > 0 {
		nw := queue[0]
		queue = queue[1:]
		switch nw.node.Typ {
		case NodeTypDivider:
			// 处理分裂节点
			// divide 方法的到的数据列表依次分给每个子节点
			action := m.actionMap[nw.node.actionId].(DividerFunc)
			if outs, err := action(context.Background(), nw.in); err != nil {
				return nil, err
			} else {
				if len(outs) == 0 || len(outs) != len(nw.node.Next) {
					err = fmt.Errorf("divider node[%s] outs null or length of outs and Next is not match", nw.node.nodeName)
					return nil, err
				}
				for i := 0; i < len(nw.node.Next); i++ {
					queue = append(queue, &nodeDataWrapper{
						node: nw.node.Next[i],
						in:   outs[i],
					})
				}
			}
		case NodeTypMerger:
			// 处理合并节点
			thre := m.inEdgeOfMerger[nw.node.nodeName]
			if thre <= 1 {
				// 报错
				err = fmt.Errorf("merger node[%s] inEdges=%d", nw.node.nodeName, thre)
				return
			}
			mergerNodeInDataMap[nw.node.nodeName] = append(mergerNodeInDataMap[nw.node.nodeName], nw.in)
			if len(mergerNodeInDataMap[nw.node.nodeName]) == thre {
				// 执行merge 方法
				action := m.actionMap[nw.node.actionId].(MergerFunc)
				if out, err = action(context.Background(), mergerNodeInDataMap[nw.node.nodeName]); err != nil {
					return
				} else {
					if len(nw.node.Next) == 0 || nw.node.Next[0] == nil {
						err = fmt.Errorf("merger node[%s] next node is nil", nw.node.nodeName)
						return
					}
					// 将下一个节点加入队列
					queue = append(queue, &nodeDataWrapper{
						node: nw.node.Next[0],
						in:   out,
					})
				}
			}
		case NodeTypJudger:
			// 处理判断节点的情况
			action := m.actionMap[nw.node.actionId].(JudgerFunc)
			pIndex := action(context.Background(), nw.in)
			if pIndex >= len(nw.node.Next) {
				err = fmt.Errorf("judger node[%s] pIndex outbound %d>=%d", nw.node.nodeName, pIndex, len(nw.node.Next))
				return
			}
			queue = append(queue, &nodeDataWrapper{
				node: nw.node.Next[pIndex],
				in:   nw.in,
			})
		case NodeTypWorker:
			// 如果是worker节点则一直往下执行
			p := nw.node
			in = nw.in
			for p != nil && p.Typ == NodeTypWorker {
				action := m.actionMap[p.actionId].(WorkerFunc)
				if out, err = action(context.Background(), in); err != nil {
					return nil, err
				} else {
					in = out
					if len(p.Next) <= 0 {
						err = fmt.Errorf("node[%s] Next is nil", p.nodeName)
						return
					}
					p = p.Next[0]
				}
			}
			// 特殊情况，报错
			if p == nil {
				err = ErrorsNodeNil
				return
			}
			// 其他类型的节点直接加入队列
			queue = append(queue, &nodeDataWrapper{
				node: p,
				in:   in,
			})
		case NodeTypTail:
			// 如果执行到末尾则返回结果
			return nw.in, nil
		}
	}
	err = ErrorsCannotReachTail
	return
}
