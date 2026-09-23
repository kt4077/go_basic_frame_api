package tree

import "sort"

// TreeNode 通用树节点接口：菜单、部门、角色均满足。
type TreeNode interface {
	GetID() uint
	GetParentID() uint
}

// TreeItem 树形返回结构。
type TreeItem[T TreeNode] struct {
	Data     T              `json:"data" comment:"节点数据"`
	Children []*TreeItem[T] `json:"children,omitempty" comment:"子节点列表"`
}

// BuildTree 将平铺列表组装成树（parent_id 为 0 或父节点不存在的作为根节点）。
func BuildTree[T TreeNode](items []T) []*TreeItem[T] {
	nodeMap := make(map[uint]*TreeItem[T], len(items))
	for _, it := range items {
		nodeMap[it.GetID()] = &TreeItem[T]{Data: it}
	}
	var link func(parentID uint) []*TreeItem[T]
	link = func(parentID uint) []*TreeItem[T] {
		var out []*TreeItem[T]
		for _, it := range items {
			if it.GetParentID() != parentID {
				continue
			}
			node := nodeMap[it.GetID()]
			node.Children = link(it.GetID())
			out = append(out, node)
		}
		return out
	}
	return link(0)
}

// CollectSelfAndDescendants 收集指定 ID 及其所有子孙节点的 ID（如：角色及子角色、部门及子部门）。
func CollectSelfAndDescendants[T TreeNode](items []T, rootID uint) []uint {
	byParent := make(map[uint][]uint, len(items))
	idSet := make(map[uint]struct{}, len(items))
	for _, it := range items {
		byParent[it.GetParentID()] = append(byParent[it.GetParentID()], it.GetID())
		idSet[it.GetID()] = struct{}{}
	}
	out := []uint{rootID}
	queue := []uint{rootID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range byParent[cur] {
			if _, ok := idSet[child]; ok { // 防御循环引用
				delete(idSet, child)
				out = append(out, child)
				queue = append(queue, child)
			}
		}
	}
	return out
}

// SortUintSlice 去重并升序排序。
func SortUintSlice(list []uint) []uint {
	seen := make(map[uint]struct{}, len(list))
	out := make([]uint, 0, len(list))
	for _, v := range list {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
