package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type FSStub struct {
	result func() ([]myDirEntry, error)
}

func (fss *FSStub) MyReadDir(input string) ([]myDirEntry, error) {
	return fss.result()
}

func (fs *FSImpl) foo(input string) ([]myDirEntry, error) {
	return fs.MyReadDir(input)
}

func TestFoo(t *testing.T) {
	stub := &FSStub{func() ([]myDirEntry, error) {
		tmp := []myDirEntry{
			{
				Name:  "test1.txt",
				Size:  1,
				IsDir: false,
			},
			{
				Name:  "test2.txt",
				Size:  2,
				IsDir: false,
			},
		}
		return tmp, nil
	}}
	tmp2 := []myDirEntry{
		{
			Name:  "test1.txt",
			Size:  1,
			IsDir: false,
		},
		{
			Name:  "test2.txt",
			Size:  2,
			IsDir: false,
		},
	}

	tmp3, _ := stub.foo("stub")

	assert.Equal(t, tmp2, tmp3)
}
