package bptree

func (t *Tree) Find(key uint64) (string, error) {
	t.rwMu.RLock()
	var (
		node *Node
		err  error
	)

	if t.rootOff == INVALID_OFFSET {
		return "", nil
	}

	if node, err = t.newMappingNodeFromPool(INVALID_OFFSET); err != nil {
		return "", err
	}	
	if err = t.findLeaf(node, key); err != nil {
		return "", err
	}
	t.rwMu.RUnlock()
	defer t.putNodePool(node)

	for i, nkey := range node.Keys {
		if nkey == key {
			return node.Records[i], nil
		}
	}
	t.NodeUnlock(node.Self)
	return "", NotFoundKey
}


/*

 version 1

1.   SL(index)
2.   Travel down to the leaf node
3.   SL(leaf)
4.   SU(index)
5.   Read the leaf node
6.   SU(leaf)


*/