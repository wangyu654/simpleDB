package engine

func (t *Tree) Insert(key uint64, val string) error {
	var (
		err  error
		node *Node
	)

	if len(val) > 8 {
		return TooLargeVal
	}

	if t.rootOff == INVALID_OFFSET {
		// 没有根结点
		if node, err = t.newNodeFromDisk(); err != nil {
			return err
		}
		t.rootOff = node.Self
		node.IsActive = true
		node.Keys = append(node.Keys, key)
		node.Records = append(node.Records, val)
		node.IsLeaf = true
		return t.flushNodeAndPutNodePool(node)
	}
	return t.insertIntoLeaf(key, val)
}

func (t *Tree) insertIntoLeaf(key uint64, rec string) error {
	t.rwMu.Lock()
	var (
		leaf     *Node
		err      error
		idx      int
		new_leaf *Node
	)
	// 空节点作为叶子节点 等待装配
	if leaf, err = t.newMappingNodeFromPool(INVALID_OFFSET); err != nil {
		return err
	}
	// 树遍历 节点内二分 查找对应叶子节点 进行装配
	if err = t.findLeaf(leaf, key); err != nil {
		return err
	}

	t.rwMu.Unlock()
	defer t.NodeUnlock(leaf.Self)
	
	// 插入到对应位置
	if idx, err = insertKeyValIntoLeaf(leaf, key, rec); err != nil {
		return err
	}

	// 递归更新父节点的相应键
	if err = t.mayUpdatedLastParentKey(leaf, idx); err != nil {
		return err
	}

	if len(leaf.Keys) <= order {
		// 刷盘
		return t.flushNodeAndPutNodePool(leaf)
	}

	// 页裂分
	if new_leaf, err = t.newNodeFromDisk(); err != nil {
		return err
	}
	new_leaf.IsLeaf = true
	if err = t.splitLeafIntoTwoleaves(leaf, new_leaf); err != nil {
		return err
	}

	if err = t.flushNodesAndPutNodesPool(new_leaf, leaf); err != nil {
		return err
	}

	// 组织裂分后的两个页和父节点的关系
	return t.insertIntoParent(leaf.Parent, leaf.Self, leaf.Keys[len(leaf.Keys)-1], new_leaf.Self)
}
