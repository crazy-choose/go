package custom

import (
	"fmt"
	"github.com/crazy-choose/go/model_bak/meta"
	"testing"
)

func TestMap_AnyMap(t *testing.T) {
	m := Map[string, meta.Position]{}
	m.Store("a", meta.Position{InstrumentID: "a"})
	m.Store("b", meta.Position{InstrumentID: "b"})
	m.Store("c", meta.Position{InstrumentID: "c"})
	m.Store("d", meta.Position{InstrumentID: "d"})
	m.Store("e", meta.Position{InstrumentID: "e"})
	fmt.Println(m)

	rm := m.Map()
	for k, v := range rm {
		fmt.Printf("k: %s, type: %T\n", k, k)
		fmt.Printf("v: %v, type: %T\n", v, v)
	}
}

func TestSet_Set(t *testing.T) {
	s := Set[int]{}
	s.Add(1)
	s.Add(2)
	s.Add(3)
	s.Add(4)
	s.Add(5)
	fmt.Println(s.Has(3))
	fmt.Println(s.Has(6))

	ss := Set[string]{}
	ss.Add("a")
	ss.Add("b")
	ss.Add("c")
	ss.Add("d")
	ss.Add("e")
	fmt.Println(ss.Has("f"))
	fmt.Println(ss.Has("d"))

}
