package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
)

var (
	debug = flag.Bool("debug", false, "debug mode")
	split = flag.Bool("split", false, "split files during defrag")
)

type Block struct {
	id   int // -1 represents free space
	file *File
}

func (b *Block) Free() bool {
	return b.id < 0
}

func (b *Block) Swap(ob *Block) {
	b.id, ob.id = ob.id, b.id
	b.file, ob.file = ob.file, b.file
}

type File struct {
	id     int
	blocks []*Block // invariant: all blocks have block.id = file.id
}

func (f *File) Size() int {
	return len(f.blocks)
}

func (f *File) Free() bool {
	return f.id < 0
}

type Map struct {
	files  []*File
	blocks []*Block
}

func (m *Map) Dense() string {
	var s strings.Builder
	id := m.blocks[0].id
	count := 1
	free := m.blocks[0].Free()
	for _, block := range m.blocks[1:] {
		if block.id == id {
			count++
		} else {
			for count > 9 {
				s.WriteString("90")
				count -= 9
			}
			s.WriteString(strconv.Itoa(count))
			// Two consecutive non-free blocks need an interstitial zero-sized
			// free block.
			if !free && !block.Free() {
				s.WriteString("0")
			}
			id = block.id
			count = 1
			free = id < 0
		}
	}
	for count > 9 {
		s.WriteString("90")
		count -= 9
	}
	s.WriteString(strconv.Itoa(count))
	return s.String()
}

func (m *Map) Sparse() string {
	var s strings.Builder
	for _, block := range m.blocks {
		if block.Free() {
			s.WriteString(".")
		} else {
			s.WriteString(strconv.Itoa(block.id))
		}
	}
	return s.String()
}

func (m *Map) SparseLimited(start, end int) string {
	var s strings.Builder
	for i, block := range m.blocks {
		if i < start || i > end {
			continue
		}
		if block.Free() {
			s.WriteString(".")
		} else {
			s.WriteString(strconv.Itoa(block.id))
		}
		s.WriteString(" ")
	}
	return s.String()
}

func (m *Map) DefragBlocks(cb func(i, j int)) {
	i, j := 0, len(m.blocks)-1

	for m.blocks[j].Free() {
		j--
	}

	for i < j {
		if m.blocks[i].Free() {
			m.blocks[i].Swap(m.blocks[j])
			for m.blocks[j].Free() {
				j--
			}
		}
		i++
		cb(i, j)
	}
}

func (m *Map) DefragFiles(cb func()) {
	cb()
	// Process files right to left
	for j := len(m.files) - 1; j > 0; j-- {
		// We only care about moving files, not free space
		if m.files[j].Free() {
			continue
		}

		for i := 0; i < j; i++ {
			// Look for a free space
			if !m.files[i].Free() {
				continue
			}
			// Free space must be big enough for j
			if m.files[i].Size() < m.files[j].Size() {
				continue
			}
			// Move the file, possibly updating j if we need to insert a split
			// file due to free space fragmentation
			j = m.MoveFile(j, i)
			cb()
			break
		}
	}
}

// MoveFile moves a file from its index to the start of a free block's index,
// and updates the corresponding blocks. If the free pseudo-file is larger than
// the file, it creates a new free pseudo-file with the leftover blocks and
// inserts it after the newly-moved file.
//
// The return value is the index of the new free space where the file used to
// be. It will be different from the original file index if the move resulted in
// a new free pseudo-file being created.
func (m *Map) MoveFile(fileIdx, freeIdx int) (j int) {
	file, free := m.files[fileIdx], m.files[freeIdx]
	free.id, file.id = file.id, free.id
	freeSize, fileSize := free.Size(), file.Size()
	for i := 0; i < fileSize; i++ {
		// We don't want to use Block.Swap here because we've already moved the
		// entire File. Block.Swap swaps the contents of each Block, which would
		// mean that we move it back where it came from, and associate it with
		// the wrong File.
		free.blocks[i].id, file.blocks[i].id = file.blocks[i].id, free.blocks[i].id
	}
	leftover := freeSize - fileSize
	if leftover == 0 {
		return fileIdx
	}
	newFree := &File{
		id:     -1,
		blocks: append([]*Block{}, free.blocks[fileSize:]...),
	}
	for _, block := range newFree.blocks {
		block.file = newFree
	}
	free.blocks = free.blocks[:fileSize]
	m.files = slices.Insert(m.files, freeIdx+1, newFree)
	return fileIdx + 1
}

func (m *Map) Checksum() int {
	sum := 0
	for i, b := range m.blocks {
		if b.Free() {
			continue
		}
		sum += i * b.id
	}
	return sum
}

func main() {
	flag.Parse()
	filename := "example.txt"
	if len(flag.Args()) > 0 {
		filename = flag.Args()[0]
	}
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	m := &Map{}

	s := bufio.NewScanner(f)
	for s.Scan() {
		entries := s.Text()
		for i, e := range entries {
			if e < '0' || e > '9' {
				log.Fatalf("invalid entry %q", e)
			}
			blocks := int(e - '0')
			id := -1
			if i%2 == 0 {
				id = i / 2
			}
			file := &File{id: id}
			m.files = append(m.files, file)
			for b := 0; b < blocks; b++ {
				block := &Block{id: id, file: file}
				m.blocks = append(m.blocks, block)
				file.blocks = append(file.blocks, block)
			}
		}
	}
	fmt.Printf("dense: %s\n", m.Dense())
	fmt.Printf("sparse: %s\n", m.Sparse())
	fmt.Printf("defragging...\n")
	if *split {
		m.DefragBlocks(func(i, j int) {
			if *debug {
				if j-i < 20 {
					fmt.Println(m.SparseLimited(i-40, j+40))
				}
				fmt.Printf("i=%d j=%d\n", i, j)
			}
		})
	} else {
		m.DefragFiles(func() {
			if *debug {
				fmt.Println(m.Sparse())
			}
		})
	}
	fmt.Printf("dense: %s\n", m.Dense())
	fmt.Printf("sparse: %s\n", m.Sparse())
	fmt.Printf("checksum: %d\n", m.Checksum())
}
