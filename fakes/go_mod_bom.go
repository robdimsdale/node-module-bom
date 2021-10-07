package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit"
)

type GoModBOM struct {
	GenerateCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir string
		}
		Returns struct {
			BOMEntrySlice []packit.BOMEntry
			Error         error
		}
		Stub func(string) ([]packit.BOMEntry, error)
	}
}

func (f *GoModBOM) Generate(param1 string) ([]packit.BOMEntry, error) {
	f.GenerateCall.mutex.Lock()
	defer f.GenerateCall.mutex.Unlock()
	f.GenerateCall.CallCount++
	f.GenerateCall.Receives.WorkingDir = param1
	if f.GenerateCall.Stub != nil {
		return f.GenerateCall.Stub(param1)
	}
	return f.GenerateCall.Returns.BOMEntrySlice, f.GenerateCall.Returns.Error
}
