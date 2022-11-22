package wasm

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/bytecodealliance/wasmtime-go"
)

var ErrNowWasmMemory = errors.New(strings.Join(
	[]string{
		`Unable to find Wasm memory import section.`,
		`  Modules must import memory from the "env" module's`,
		`  "memory" field like so:`,
		`    (import "env" "memory" (memory (;0;) #))`,
	},
	"\n",
))

type WasmInstance struct {
	store    *wasmtime.Store
	linker   *wasmtime.Linker
	instance *wasmtime.Instance
}

func NewInstance(wasm []byte) (*WasmInstance, error) {
	store := wasmtime.NewStore(wasmtime.NewEngine())
	module, err := wasmtime.NewModule(store.Engine, wasm)
	if err != nil {
		return nil, err
	}

	linker := wasmtime.NewLinker(store.Engine)
	memory, err := createMemory(wasm, store)
	if err != nil {
		return nil, err
	}
	createImport(linker, store, memory)
	instance, err := linker.Instantiate(store, module)
	if err != nil {
		return nil, err
	}

	return &WasmInstance{
		store:    store,
		linker:   linker,
		instance: instance,
	}, nil
}

func (w *WasmInstance) WrapInvoke() {
}

func createMemory(wasm []byte, store *wasmtime.Store) (*wasmtime.Memory, error) {
	ENV_MEMORY_IMPORTS_SIGNATURE := []byte{0x65, 0x6e, 0x76, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x02}
	sigIdx := bytes.Index(wasm, ENV_MEMORY_IMPORTS_SIGNATURE)
	if sigIdx < 0 {
		return nil, ErrNowWasmMemory
	}
	memoryInitialLimits := wasm[sigIdx+1+len(ENV_MEMORY_IMPORTS_SIGNATURE)+1]
	memoryType := wasmtime.NewMemoryType(uint32(memoryInitialLimits), false, 0)
	return wasmtime.NewMemory(store, memoryType)
}

func createImport(linker *wasmtime.Linker, store *wasmtime.Store, memory *wasmtime.Memory) {
	linker.FuncWrap("wrap", "__wrap_load_env", func(ptr int32) {
		fmt.Printf("%s\n", memory.UnsafeData(store))
		panic("__wrap_load_env not implemented")
	})
	linker.FuncWrap("wrap", "__wrap_invoke_args", func(methodPtr, argsPtr int32) {
		method := (*[]byte)(unsafe.Pointer(&methodPtr))
		//args := (*[]byte)(unsafe.Pointer(&argsPtr))
		//mem := memory.UnsafeData(store)
		//copy(mem[methodPtr:], (*[]byte)(unsafe.Pointer(&methodPtr)))
		//copy(mem[argsPtr:], (*[]byte)(unsafe.Pointer(&argsPtr)))
		panic(fmt.Sprintf(
			"__wrap_invoke_args not implemented %#v, %#v\n",
			method,
			argsPtr,
		))
	})
	linker.FuncWrap("wrap", "__wrap_invoke_result", func(ptr, len int32) {
		//return memory.UnsafeData(caller)[ptr : ptr+len]
		panic("__wrap_invoke_result not implemented")
	})
	linker.FuncWrap("wrap", "__wrap_invoke_error", func(ptr, len int32) {
		//return memory.UnsafeData(caller)[ptr : ptr+len]
		panic("__wrap_invoke_error not implemented")
	})
	linker.FuncWrap("wrap", "__wrap_abort", func(msgPtr, msgLen, filePtr, fileLen, line, column int32) {
		mem := memory.UnsafeData(store)
		msg := string(mem[msgPtr : msgPtr+msgLen])
		file := string(mem[filePtr : filePtr+fileLen])
		panic(fmt.Sprintf("__wrap_abort: %s\nFile: %s\nLocation: [{%d},{%d}]", msg, file, line, column))
	})
	linker.FuncWrap("wrap", "__wrap_subinvoke", func(uriPtr, uriLen, methodPtr, methodLen, argsPtr, argsLen int32) int32 {
		//mem := memory.UnsafeData(caller)
		//uri := string(mem[uriPtr : uriPtr+uriLen])
		//method := string(mem[methodPtr : methodPtr+methodLen])
		//args := mem[argsPtr : argsPtr+argsLen]
		//panic(fmt.Sprintf(
		//	"Uri: %s\nMethod: %s\nArgs: %x\n  __wrap_subinvoke not implemented",
		//	uri,
		//	method,
		//	args,
		//))
		panic("__wrap_subinvoke not implemented")
	})
	linker.FuncWrap("wrap", "__wrap_subinvoke_result_len", func() int32 {
		panic("__wrap_subinvoke_result_len not implemented")
	})
	linker.FuncWrap("wrap", "__wrap_subinvoke_result", func(ptr int32) {
		panic("__wrap_subinvoke_result not implemented")
	})
	linker.FuncWrap("wrap", "__wrap_subinvoke_error_len", func() int32 {
		panic("__wrap_subinvoke_error_len not implemented")
	})
	linker.FuncWrap("wrap", "__wrap_subinvoke_error", func(ptr int32) {
		panic("__wrap_subinvoke_result not implemented")
	})
	linker.Define("env", "memory", memory)
}
