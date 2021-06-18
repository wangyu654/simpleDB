package engine

import "fmt"

func (t *Tree) Delete(key uint64) error {
	if t.rootOff == INVALID_OFFSET {
		return fmt.Errorf("not found key:%d", key)
	}
	return t.deleteKeyFromLeaf(key)
}

func (t *Tree) deleteKeyFromLeaf(key uint64) error {
	var (
		leaf     *Node
		prevLeaf *Node
		nextLeaf *Node
		err      error
		idx      int
	)
	if leaf, err = t.newMappingNodeFromPool(INVALID_OFFSET); err != nil {
		return err
	}

	if err = t.findLeaf(leaf, key); err != nil {
		return err
	}

	idx = getIndex(leaf.Keys, key)
	if idx == len(leaf.Keys) || leaf.Keys[idx] != key {
		t.putNodePool(leaf)
		return fmt.Errorf("not found key:%d", key)

	}
	removeKeyFromLeaf(leaf, idx)

	// if leaf is root
	if leaf.Self == t.rootOff {
		return t.flushNodesAndPutNodesPool(leaf)
	}

	// update the last key of parent's if necessary
	if idx == len(leaf.Keys) {
		if err = t.mayUpdatedLastParentKey(leaf, idx-1); err != nil {
			return err
		}
	}

	// if satisfied len
	if len(leaf.Keys) >= order/2 {
		return t.flushNodesAndPutNodesPool(leaf)
	}

	if leaf.Next != INVALID_OFFSET {
		if nextLeaf, err = t.newMappingNodeFromPool(leaf.Next); err != nil {
			return err
		}
		// lease from next leaf
		if len(nextLeaf.Keys) > order/2 {
			key := nextLeaf.Keys[0]
			rec := nextLeaf.Records[0]
			removeKeyFromLeaf(nextLeaf, 0)
			if idx, err = insertKeyValIntoLeaf(leaf, key, rec); err != nil {
				return err
			}
			// update the last key of parent's if necessy
			if err = t.mayUpdatedLastParentKey(leaf, idx); err != nil {
				return err
			}
			return t.flushNodesAndPutNodesPool(nextLeaf, leaf)
		}

		// merge nextLeaf and curleaf
		if leaf.Prev != INVALID_OFFSET {
			if prevLeaf, err = t.newMappingNodeFromPool(leaf.Prev); err != nil {
				return err
			}
			prevLeaf.Next = nextLeaf.Self
			nextLeaf.Prev = prevLeaf.Self
			if err = t.flushNodesAndPutNodesPool(prevLeaf); err != nil {
				return err
			}
		} else {
			nextLeaf.Prev = INVALID_OFFSET
		}

		nextLeaf.Keys = append(leaf.Keys, nextLeaf.Keys...)
		nextLeaf.Records = append(leaf.Records, nextLeaf.Records...)

		leaf.IsActive = false
		t.putFreeBlocks(leaf.Self)

		if err = t.flushNodesAndPutNodesPool(leaf, nextLeaf); err != nil {
			return err
		}

		return t.deleteKeyFromNode(leaf.Parent, leaf.Keys[len(leaf.Keys)-1])
	}

	// come here because leaf.Next = INVALID_OFFSET
	if leaf.Prev != INVALID_OFFSET {
		if prevLeaf, err = t.newMappingNodeFromPool(leaf.Prev); err != nil {
			return err
		}
		// lease from prev leaf
		if len(prevLeaf.Keys) > order/2 {
			key := prevLeaf.Keys[len(prevLeaf.Keys)-1]
			rec := prevLeaf.Records[len(prevLeaf.Records)-1]
			removeKeyFromLeaf(prevLeaf, len(prevLeaf.Keys)-1)
			if idx, err = insertKeyValIntoLeaf(leaf, key, rec); err != nil {
				return err
			}
			// update the last key of parent's if necessy
			if err = t.mayUpdatedLastParentKey(prevLeaf, len(prevLeaf.Keys)-1); err != nil {
				return err
			}
			return t.flushNodesAndPutNodesPool(prevLeaf, leaf)
		}
		// merge prevleaf and curleaf
		prevLeaf.Next = INVALID_OFFSET
		prevLeaf.Keys = append(prevLeaf.Keys, leaf.Keys...)
		prevLeaf.Records = append(prevLeaf.Records, leaf.Records...)

		leaf.IsActive = false
		t.putFreeBlocks(leaf.Self)

		if err = t.flushNodesAndPutNodesPool(leaf, prevLeaf); err != nil {
			return err
		}

		return t.deleteKeyFromNode(leaf.Parent, leaf.Keys[len(leaf.Keys)-2])
	}

	return nil
}
