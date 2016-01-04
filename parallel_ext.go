package main

import (
	"github.com/wushilin/parallel"
	"github.com/wushilin/stream"
)

type PStream interface {
	PMap(f stream.MapFunc) stream.Stream
}

type basePStream struct {
	stream.Stream
}

func (v *basePStream) PMap(f stream.MapFunc) stream.Stream {
	futures := make([]interface{}, 0)
	iter := v.Map(func(arg interface{}) interface{} {
		return parallel.MakeFuture(parallel.MapFunc(f), arg)
	}).Iterator()

	for v, ok := iter.Next(); ok; v, ok = iter.Next() {
		futures = append(futures, v)
	}

	return stream.FromArray(futures).Map(func(future interface{}) interface{} {
		return future.(parallel.Future).Wait()
	})
}
