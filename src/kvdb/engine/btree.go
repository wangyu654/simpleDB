package engine

import (
	"errors"
	"os"
	"sync"
	"syscall"
)

var (
	err   error
	order = 128
)

const (
	INVALID_OFFSET = 0xdeadbeef
	MAX_FREEBLOCKS = 100 //100个block组成一个file
)

type OFFTYPE uint64

type Tree struct {
	rootOff    OFFTYPE
	nodePool   *sync.Pool
	freeBlocks []OFFTYPE
	file       *os.File
	blockSize  uint64
	fileSize   uint64
}

type Node struct {
	IsActive bool      //1 节点所在的磁盘空间是否在当前b+树内
	Children []OFFTYPE //
	Self     OFFTYPE   //8
	Next     OFFTYPE   //8
	Prev     OFFTYPE   //8
	Parent   OFFTYPE   //8
	Keys     []uint64  // 8字节的key
	Records  []string  // 最大8字节的字符串
	IsLeaf   bool      //1
}

func NewTree(filename string) (*Tree, error) {
	var (
		stat  syscall.Statfs_t
		fstat os.FileInfo
		err   error
	)

	t := &Tree{}

	t.rootOff = INVALID_OFFSET
	t.nodePool = &sync.Pool{
		//get 返回一个空node
		New: func() interface{} {
			return &Node{}
		},
	}
	//	freeBlock 缓存
	t.freeBlocks = make([]OFFTYPE, 0, MAX_FREEBLOCKS)
	//	btree 对应的文件
	if t.file, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644); err != nil {
		return nil, err
	}
	// 获取磁盘的信息
	if err = syscall.Statfs(filename, &stat); err != nil {
		return nil, err
	}
	// 磁盘的一个块大小 stat.Bsize    4096byte 4kb
	t.blockSize = uint64(stat.Bsize)
	if t.blockSize == 0 {
		return nil, errors.New("blockSize should be zero")
	}
	// 获取文件信息
	if fstat, err = t.file.Stat(); err != nil {
		return nil, err
	}

	t.fileSize = uint64(fstat.Size())
	if t.fileSize != 0 {
		//
		if err = t.restructRootNode(); err != nil {
			return nil, err
		}
		//
		if err = t.checkDiskBlockForFreeNodeList(); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (t *Tree) restructRootNode() error {
	var (
		err error
	)
	node := &Node{}

	for off := uint64(0); off < t.fileSize; off += t.blockSize {
		//
		if err = t.seekNode(node, OFFTYPE(off)); err != nil {
			return err
		}
		if node.IsActive {
			break
		}
	}
	if !node.IsActive {
		return InvalidDBFormat
	}
	for node.Parent != INVALID_OFFSET {
		if err = t.seekNode(node, node.Parent); err != nil {
			return err
		}
	}

	t.rootOff = node.Self

	return nil
}

func (t *Tree) checkDiskBlockForFreeNodeList() error {
	var (
		err error
	)
	node := &Node{}
	bs := t.blockSize
	for off := uint64(0); off < t.fileSize && len(t.freeBlocks) < MAX_FREEBLOCKS; off += bs {
		if off+bs > t.fileSize {
			break
		}
		if err = t.seekNode(node, OFFTYPE(off)); err != nil {
			return err
		}
		if !node.IsActive {
			t.freeBlocks = append(t.freeBlocks, OFFTYPE(off))
		}
	}
	next_file := ((t.fileSize + 4095) / 4096) * 4096
	for len(t.freeBlocks) < MAX_FREEBLOCKS {
		//
		t.freeBlocks = append(t.freeBlocks, OFFTYPE(next_file))
		next_file += bs
	}
	t.fileSize = next_file
	return nil
}

func (t *Tree) Close() error {
	if t.file != nil {
		t.file.Sync()
		return t.file.Close()
	}
	return nil
}
