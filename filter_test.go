package golombfilter

import (
	"encoding/binary"
	"github.com/spaolacci/murmur3"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

func inttobytes(v int) []byte {
	n := make([]byte, 4)
	binary.LittleEndian.PutUint32(n, uint32(v))
	return n
}

func TestBuilder_AddValue(t *testing.T) {
	tests := []struct {
		values []int
		power  int
	}{
		{[]int{1, 3, 5}, 4},
		{rand.Perm(100), 4},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {

			expected := make(map[uint32]bool)
			mur := murmur3.New32()
			for _, v := range test.values {

				mur.Reset()
				mur.Write(inttobytes(v))
				v := mur.Sum32() % uint32(len(test.values)*(1<<test.power))
				expected[v] = true
			}

			builder := Builder(test.power)
			for _, num := range test.values {
				builder.AddValue(inttobytes(num))
			}
			f := builder.Filter()

			if !reflect.DeepEqual(f.hashes, expected) {
				t.Fatalf("expected \n%v\n but found \n%v\n", expected, f.hashes)
			}
		})
	}
}

func TestFilter_Contains(t *testing.T) {
	tests := []struct {
		values      []int
		power       int
		expectTrue  []int
		expectFalse []int
	}{
		{[]int{1, 2, 4, 5}, 4, []int{1, 2, 4, 5}, []int{3}},
		{[]int{1, 2, 4, 5}, 2, []int{1, 2, 4, 5}, []int{3}},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			b := Builder(test.power)
			for _, v := range test.values {
				b.AddValue(inttobytes(v))
			}
			f := b.Filter()

			for _, tv := range test.expectTrue {
				if !f.Contains(inttobytes(tv)) {
					t.Fatalf("expected to contain %v", tv)
				}
			}

			for _, tv := range test.expectFalse {
				if f.Contains(inttobytes(tv)) {
					t.Fatalf("expected not to contain %v", tv)
				}
			}
		})
	}
}

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		values []int
		power  int
	}{
		{[]int{1, 3, 4, 5}, 4},
		{rand.Perm(1000), 4},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			b := Builder(test.power)
			for _, v := range test.values {
				b.AddValue(inttobytes(v))
			}
			f := b.Filter()

			bits := Encode(f)

			f2 := Decode(bits, f.power, f.hasher)

			if !reflect.DeepEqual(f.hashes, f2.hashes) {
				t.Fatalf("expected \n%v\n but found \n%v\n", f.hashes, f2.hashes)
			}

			for _, tv := range test.values {
				btv := inttobytes(tv)
				if f.Contains(btv) != f2.Contains(btv) {
					t.Fatalf("expected equality for %v", tv)
				}
			}
		})
	}
}
