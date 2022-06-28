// Code generated by counterfeiter. DO NOT EDIT.
package fake

import (
	"sync"

	"code.cloudfoundry.org/korifi/api/repositories"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type ImagePusher struct {
	PushStub        func(string, v1.Image, remote.Option, remote.Option) (string, error)
	pushMutex       sync.RWMutex
	pushArgsForCall []struct {
		arg1 string
		arg2 v1.Image
		arg3 remote.Option
		arg4 remote.Option
	}
	pushReturns struct {
		result1 string
		result2 error
	}
	pushReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ImagePusher) Push(arg1 string, arg2 v1.Image, arg3 remote.Option, arg4 remote.Option) (string, error) {
	fake.pushMutex.Lock()
	ret, specificReturn := fake.pushReturnsOnCall[len(fake.pushArgsForCall)]
	fake.pushArgsForCall = append(fake.pushArgsForCall, struct {
		arg1 string
		arg2 v1.Image
		arg3 remote.Option
		arg4 remote.Option
	}{arg1, arg2, arg3, arg4})
	stub := fake.PushStub
	fakeReturns := fake.pushReturns
	fake.recordInvocation("Push", []interface{}{arg1, arg2, arg3, arg4})
	fake.pushMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ImagePusher) PushCallCount() int {
	fake.pushMutex.RLock()
	defer fake.pushMutex.RUnlock()
	return len(fake.pushArgsForCall)
}

func (fake *ImagePusher) PushCalls(stub func(string, v1.Image, remote.Option, remote.Option) (string, error)) {
	fake.pushMutex.Lock()
	defer fake.pushMutex.Unlock()
	fake.PushStub = stub
}

func (fake *ImagePusher) PushArgsForCall(i int) (string, v1.Image, remote.Option, remote.Option) {
	fake.pushMutex.RLock()
	defer fake.pushMutex.RUnlock()
	argsForCall := fake.pushArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *ImagePusher) PushReturns(result1 string, result2 error) {
	fake.pushMutex.Lock()
	defer fake.pushMutex.Unlock()
	fake.PushStub = nil
	fake.pushReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *ImagePusher) PushReturnsOnCall(i int, result1 string, result2 error) {
	fake.pushMutex.Lock()
	defer fake.pushMutex.Unlock()
	fake.PushStub = nil
	if fake.pushReturnsOnCall == nil {
		fake.pushReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.pushReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *ImagePusher) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.pushMutex.RLock()
	defer fake.pushMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ImagePusher) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ repositories.ImagePusher = new(ImagePusher)
