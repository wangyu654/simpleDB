package bptree

// func (t *Tree) NodeLock(off OFFTYPE) bool {
// 	t.nodeCond.L.Lock()
// 	if _, ok := t.nodeMuMap.Load(off); ok {
// 		return false
// 	} else {
// 		t.nodeMuMap.Store(off, 1)
// 		return true
// 	}
// }

// func (t *Tree) NodeUnlock(off OFFTYPE) {
// 	t.nodeMuMap.Delete(off)
// 	t.nodeCond.Broadcast()
// 	t.nodeCond.L.Unlock()
// }
