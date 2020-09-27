package golombfilter

import (
	"github.com/spaolacci/murmur3"
	"hash"
	"sort"
)

type Filter struct {
	hashes map[uint32]bool
	power  int
	hasher hash.Hash32
}

//New creates a new Golomb Rice Encoding Filter/Set using M=2^power
func New(hashes []uint32, power int, hasher hash.Hash32) *Filter {
	if power < 1 {
		panic("power > 0 is required")
	}
	if power > 31 {
		panic("power < 32 is required")
	}
	hs := make(map[uint32]bool)
	m := uint32(len(hashes) * (1 << power))
	for _, h := range hashes {
		hs[h%m] = true
	}
	return &Filter{

		hashes: hs,
		power:  power,
		hasher: hasher,
	}
}

func (f *Filter) ContainsHash(hash uint32) bool {
	x := uint32(len(f.hashes) * (1 << f.power))
	_, ok := f.hashes[hash%x]
	return ok
}

func (f *Filter) Contains(value []byte) bool {
	f.hasher.Reset()
	f.hasher.Write(value)
	return f.ContainsHash(f.hasher.Sum32())
}

func Encode(f *Filter) []int {
	values := make([]uint32, len(f.hashes))
	i := 0
	for h := range f.hashes {
		values[i] = h
		i++
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	bits := make([]int, 0)
	remainderMask := uint32(1<<f.power - 1)
	bits = appendCoding(bits, values[0], remainderMask, f.power)
	for i := 0; i < len(values)-1; i++ {
		bits = appendCoding(bits, values[i+1]-values[i], remainderMask, f.power)
	}
	return bits
}

func appendCoding(bits []int, value, remainderMask uint32, power int) []int {
	//S & (M - 1) = 18 & (16 - 1) = 0b00010010 & 0b1111 = 0b0010
	//S >> K = 18 >> 4 = 0b00010010 >> 4 = 0b0001 (10 in unary)
	remainder := value & remainderMask
	uniaryCount := value >> power

	for j := uint32(0); j < uniaryCount; j++ {
		bits = append(bits, 1)
	}
	bits = append(bits, 0)

	for j := power - 1; 0 <= j; j-- {
		bits = append(bits, int(remainder>>j&1))
	}
	return bits
}

func Decode(bits []int, power int, hasher hash.Hash32) *Filter {
	var diff uint32
	values := make(map[uint32]bool)
	value := uint32(0)
	index := 0
	for index < len(bits) {
		diff, index = extractValue(bits, index, power)
		value = value + diff
		values[value] = true
	}
	return &Filter{
		hashes: values,
		power:  power,
		hasher: hasher,
	}
}

//extractValue extracts a golomb rice encoded value starting at index based on the M=2^power
func extractValue(bits []int, index, power int) (value uint32, nextIndex int) {
	uniaryCount := 0
	for i := index; i < len(bits); i++ {
		if bits[i] == 0 {
			break
		}
		uniaryCount++
	}
	remainderIndex := index + uniaryCount + 1
	remainder := 0
	for i := 0; i < power; i++ {
		remainder += bits[remainderIndex+i] << (power - 1 - i)
	}
	return uint32(uniaryCount<<power) | uint32(remainder), index + uniaryCount + 1 + power
}

func Builder(power int) *builder {
	return BuilderHash(power, murmur3.New32())
}

func BuilderHash(power int, specificHash hash.Hash32) *builder {
	return &builder{
		power:  power,
		hasher: specificHash,
		data:   make([]uint32, 0),
	}
}

type builder struct {
	power  int
	hasher hash.Hash32
	data   []uint32
}

func (b *builder) AddValue(value []byte) {
	b.hasher.Reset()
	b.hasher.Write(value)
	b.data = append(b.data, b.hasher.Sum32())
}

//Filter returns a Golomb Rice Filter
func (b *builder) Filter() *Filter {
	return New(b.data, b.power, b.hasher)
}
