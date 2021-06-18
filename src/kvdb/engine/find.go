package engine

func (t *Tree) Find(key uint64) (string, error) {
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
	defer t.putNodePool(node)

	for i, nkey := range node.Keys {
		if nkey == key {
			return node.Records[i], nil
		}
	}
	return "", NotFoundKey
}
