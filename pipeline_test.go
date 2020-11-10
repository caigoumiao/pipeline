package pipeline

import (
	"context"
	"fmt"
	"testing"
)

// 测试pipeline的构建
func TestManager_BuildPipeline(t *testing.T) {
	m := NewManager()
	if err := m.AddWorkerNode("work1", func(ctx context.Context, in *rawData) (out *rawData, err error) {
		if a, ok := in.Data.(int); !ok {
			err = fmt.Errorf("type of in.Data is not int")
		} else {
			fmt.Println(a)
			in.Data = a + 2
			out = in
		}
		return
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.AddWorkerNode("work2", func(ctx context.Context, in *rawData) (out *rawData, err error) {
		if a, ok := in.Data.(int); !ok {
			err = fmt.Errorf("type of in.Data is not int")
		} else {
			fmt.Println(a)
			in.Data = a * 3
			out = in
		}
		return
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.BuildPipeline([][]string{
		{"head000", "work1"},
		{"work1", "work2"},
		{"work2", "tail111"},
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
}

// 测试情况1：
// 只有工作节点，一条直线执行
// 计算：(a+2)*5
func TestManager_Handle1(t *testing.T) {
	m := NewManager()
	if err := m.AddWorkerNode("work1", func(ctx context.Context, in *rawData) (out *rawData, err error) {
		if a, ok := in.Data.(int); !ok {
			err = fmt.Errorf("type of in.Data is not int")
		} else {
			fmt.Println(a)
			in.Data = a + 2
			out = in
		}
		return
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.AddWorkerNode("work2", func(ctx context.Context, in *rawData) (out *rawData, err error) {
		if a, ok := in.Data.(int); !ok {
			err = fmt.Errorf("type of in.Data is not int")
		} else {
			fmt.Println(a)
			in.Data = a * 3
			out = in
		}
		return
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.BuildPipeline([][]string{
		{"head000", "work1"},
		{"work1", "work2"},
		{"work2", "tail111"},
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if out, err := m.Handle(&rawData{
		Data: 3,
	}); err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		if out.Data.(int) != 15 {
			t.Errorf("out=%d not eq 15", out.Data.(int))
			t.FailNow()
		}
	}
}

// 测试情况2：
// 带有分裂合并节点
// 计算bool值：(a+2)*5 < (a+3)*4
func TestManager_Handle2(t *testing.T) {
	m := NewManager()
	if err := m.AddDividerNode("divider1", func(ctx context.Context, in *rawData) (out []*rawData, err error) {
		out = append(out, in)
		out = append(out, &rawData{
			Data: in.Data,
		})
		return
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.AddWorkerNode("w1", func(ctx context.Context, in *rawData) (out *rawData, err error) {
		if a, ok := in.Data.(int); !ok {
			err = fmt.Errorf("type of in.Data is not int")
			return
		} else {
			in.Data = a + 2
			out = in
			return
		}
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.AddWorkerNode("w2", func(ctx context.Context, in *rawData) (out *rawData, err error) {
		if a, ok := in.Data.(int); !ok {
			err = fmt.Errorf("type of in.Data is not int")
			return
		} else {
			in.Data = a * 5
			out = in
			return
		}
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.AddWorkerNode("w3", func(ctx context.Context, in *rawData) (out *rawData, err error) {
		if a, ok := in.Data.(int); !ok {
			err = fmt.Errorf("type of in.Data is not int")
			return
		} else {
			in.Data = a + 3
			out = in
			return
		}
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.AddWorkerNode("w4", func(ctx context.Context, in *rawData) (out *rawData, err error) {
		if a, ok := in.Data.(int); !ok {
			err = fmt.Errorf("type of in.Data is not int")
			return
		} else {
			in.Data = a * 4
			out = in
			return
		}
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.AddMergerNode("m1", func(ctx context.Context, in []*rawData) (out *rawData, err error) {
		if len(in) != 2 {
			err = fmt.Errorf("inData length wrong")
			return
		}
		out = &rawData{
			Meta: make(map[string]interface{}),
		}
		out.Meta["res"] = in[0].Data.(int) < in[1].Data.(int)
		return
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.BuildPipeline([][]string{
		{"head000", "divider1"},
		{"divider1", "w1"},
		{"divider1", "w3"},
		{"w1", "w2"},
		{"w3", "w4"},
		{"w2", "m1"},
		{"w4", "m1"},
		{"m1", "tail111"},
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	var a = 1
	if out, err := m.Handle(&rawData{
		Data: a,
	}); err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		if !out.Meta["res"].(bool) {
			t.Errorf("wrong! res=false, ans=true")
		}
	}
}

// 测试情况：
// 带有判断节点
//
func TestManager_Handle3(t *testing.T) {
	m := NewManager()
	if err := m.AddJudgerNode("j1", func(ctx context.Context, in *rawData) (pipeIndex int) {
		a := in.Data.(int)
		if a < 100 {
			return 0
		} else if a < 200 {
			return 1
		} else {
			return 2
		}
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.AddWorkerNode("w1", func(ctx context.Context, in *rawData) (out *rawData, err error) {
		if a, ok := in.Data.(int); !ok {
			err = fmt.Errorf("type of in.Data is not int")
			return
		} else {
			in.Data = a + 5
			out = in
			return
		}
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.AddWorkerNode("w2", func(ctx context.Context, in *rawData) (out *rawData, err error) {
		if a, ok := in.Data.(int); !ok {
			err = fmt.Errorf("type of in.Data is not int")
			return
		} else {
			in.Data = a * 2
			out = in
			return
		}
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.AddWorkerNode("w3", func(ctx context.Context, in *rawData) (out *rawData, err error) {
		err = fmt.Errorf("data out bound")
		return
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := m.BuildPipeline([][]string{
		{"head000", "j1"},
		{"j1", "w1"},
		{"j1", "w2"},
		{"j1", "w3"},
		{"w1", "tail111"},
		{"w2", "tail111"},
		{"w3", "tail111"},
	}); err != nil {
		t.Error(err)
		t.FailNow()
	}
	var a = 1
	if out, err := m.Handle(&rawData{
		Data: a,
	}); err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		if out.Data.(int) != 6 {
			t.Errorf("res=%d, trueAnswer=%d", out.Data.(int), 6)
			t.FailNow()
		}
		t.Log("test1 passed")
	}

	a = 150
	if out, err := m.Handle(&rawData{
		Data: a,
	}); err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		if out.Data.(int) != 300 {
			t.Errorf("res=%d, trueAnswer=%d", out.Data.(int), 300)
			t.FailNow()
		}
		t.Log("test2 passed")
	}

	a = 203
	if _, err := m.Handle(&rawData{
		Data: a,
	}); err == nil {
		t.Errorf("predict error occurs, but not")
		t.FailNow()
	} else {
		t.Log("test3 passed")
		fmt.Println(err.Error())
	}
}
