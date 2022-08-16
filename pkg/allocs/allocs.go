package allocs

import (
	"crypto/sha1"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	SDL_WINDOW = iota
	SDL_RENDERER
	SDL_TEXTURE
	SDL_SURFACE
	SDL_FONT
)

var (
	defaultAllocs = newAllocs()
)

func type2Str(val int) string {
	switch val {
	case SDL_WINDOW:
		return "SDL_WINDOW"
	case SDL_RENDERER:
		return "SDL_RENDERER"
	case SDL_TEXTURE:
		return "SDL_TEXTURE"
	case SDL_SURFACE:
		return "SDL_SURFACE"
	case SDL_FONT:
		return "SDL_FONT"
	}

	panic("bad type")
}

type Allocs struct {
	done       bool
	mu         sync.RWMutex
	cgoMetrics map[int](map[string]string) // type : (id: stackid)
	stacks     map[string][]byte           // stackid: stack
}

func newAllocs() *Allocs {
	allocs := &Allocs{
		cgoMetrics: map[int](map[string]string){},
		stacks:     map[string][]byte{},
	}

	// 提前分配内存
	allocs.cgoMetrics[SDL_WINDOW] = map[string]string{}
	allocs.cgoMetrics[SDL_RENDERER] = map[string]string{}
	allocs.cgoMetrics[SDL_TEXTURE] = map[string]string{}
	allocs.cgoMetrics[SDL_SURFACE] = map[string]string{}
	allocs.cgoMetrics[SDL_FONT] = map[string]string{}

	go func() {
		lastTimes := map[int](map[string]int){}
		//thisTimes := map[int](map[string]int){}

		for !allocs.done {
			fmt.Println("=======", "sum", "=======")

			usedIds := map[string]struct{}{}
			allocs.mu.RLock()
			for intType, kv := range allocs.cgoMetrics {
				fmt.Println(type2Str(intType), ": ", len(kv))

				thisTimes := map[string]int{}

				// 清理
				for _, id := range kv {
					usedIds[id] = struct{}{}

					// 统计
					if _, ok := thisTimes[id]; ok {
						thisTimes[id]++
					} else {
						thisTimes[id] = 1
					}
				}

				for id, times := range thisTimes {
					oldTimes, ok := lastTimes[intType][id]
					if !ok {
						oldTimes = 0
					}
					ret := times - oldTimes
					if ret > 0 {
						fmt.Printf("%s, %d to %d\n", id, oldTimes, times)
					} else if ret < 0 {
						fmt.Printf("%s, %d to %d\n", id, oldTimes, times)
					} else {
						fmt.Printf("%s, keep %d\n", id, times)
					}
				}

				lastTimes[intType] = thisTimes
			}

			//fmt.Printf("remove %d stack ids\n", len(delIds))
			i := 0
			for id, _ := range allocs.stacks {
				if _, ok := usedIds[id]; !ok {
					i++
					delete(allocs.stacks, id)
				}
			}

			fmt.Printf("remove %d stack ids, left: %d \n", i, len(allocs.stacks))
			allocs.mu.RUnlock()
			time.Sleep(30 * time.Second)
		}
	}()

	return allocs
}

func (this *Allocs) PrintStackInfo() {
	this.mu.RLock()
	defer this.mu.RUnlock()

	for type1, kv := range this.cgoMetrics {
		fmt.Println("=====", type2Str(type1), "=====")
		for _, id := range kv {
			fmt.Println(id)
		}
	}
}

func (this *Allocs) PrintAllStack() {
	this.done = true

	this.mu.RLock()
	defer this.mu.RUnlock()

	fmt.Println()
	for intType, kv := range this.cgoMetrics {

		fmt.Println("=======", "stack", "=======")
		fmt.Println(type2Str(intType), ": ", len(kv))

		for key, val := range kv {
			fmt.Println()
			fmt.Printf("(%s %s): %s\n", key, val, this.stacks[val])
		}
	}
}

func (this *Allocs) register(type1 int, id string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	if stackId, ok := this.cgoMetrics[type1][id]; ok {
		panic("freed outerside before this alloc: \n" + string(this.stacks[stackId]))
	}

	byteStack := debug.Stack()
	//hashStack := fmt.Sprintf("%d", tools.HashBytes(byteStack))
	hashStack := sha1.Sum(byteStack)

	var buffer []byte = make([]byte, len(hashStack))
	copy(buffer, hashStack[:len(hashStack)])
	stackId := fmt.Sprintf("%x", buffer)

	// record stack
	this.cgoMetrics[type1][id] = stackId
	if _, ok := this.stacks[stackId]; ok {
		return
	}

	this.stacks[stackId] = byteStack
}

func (this *Allocs) Delete(rawPtr interface{}) {
	if rawPtr == nil {
		return
	}

	switch rawPtr.(type) {
	case *sdl.Window:
		ptr := rawPtr.(*sdl.Window)
		defer ptr.Destroy()
		//id := *(*uint64)(unsafe.Pointer(ptr))
		id := fmt.Sprintf("%p", ptr)

		this.mu.Lock()
		defer this.mu.Unlock()
		delete(this.cgoMetrics[SDL_WINDOW], id)
	case *sdl.Renderer:
		ptr := rawPtr.(*sdl.Renderer)
		//id := *(*uint64)(unsafe.Pointer(ptr))
		id := fmt.Sprintf("%p", ptr)
		ptr.Destroy()
		this.mu.Lock()
		defer this.mu.Unlock()
		delete(this.cgoMetrics[SDL_RENDERER], id)
	case *sdl.Texture:
		ptr := rawPtr.(*sdl.Texture)
		defer ptr.Destroy()
		//id := *(*uint64)(unsafe.Pointer(ptr))
		id := fmt.Sprintf("%p", ptr)
		this.mu.Lock()
		defer this.mu.Unlock()
		delete(this.cgoMetrics[SDL_TEXTURE], id)
	case *sdl.Surface:
		ptr := rawPtr.(*sdl.Surface)
		//id := *(*uint64)(unsafe.Pointer(ptr))
		id := fmt.Sprintf("%p", ptr)
		ptr.Free()
		this.mu.Lock()
		defer this.mu.Unlock()
		delete(this.cgoMetrics[SDL_SURFACE], id)
	case *ttf.Font:
		ptr := rawPtr.(*ttf.Font)
		//id := *(*uint64)(unsafe.Pointer(ptr))
		id := fmt.Sprintf("%p", ptr)
		ptr.Close()
		this.mu.Lock()
		defer this.mu.Unlock()
		delete(this.cgoMetrics[SDL_FONT], id)
	}
}

// default api

func PrintAllStack() {
	defaultAllocs.PrintAllStack()
}

func PrintStackInfo() {
	defaultAllocs.PrintStackInfo()
}

func Delete(rawPtr interface{}) {
	defaultAllocs.Delete(rawPtr)
}
