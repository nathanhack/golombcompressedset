# golombcompressedset
**golombcompressedset** is Golomb Rice Coded Filter/Set. It is similar to a Bloom Filter but with asymptotically better space characteristics for low error probability. The **golombcompressedset** specifically uses the Golomb Rice encoding scheme.  

## Usage
There are two ways to create a `Filter`.  The first is using a builder.  The second is directly via the `New`.

### Builder
The library contains a builder that can be created by calling `Builder(power)` where `power > 0` and specifies the probability of false positives given by `1/2^power`.

Adding values to the builder can be done by calling `AddValue(value)` where value is a `[]byte`.

After finishing adding values to the builder calling `Filter()` to create a `Filter`. 


### New
If hashes of all the values is known then a `Filter` can be create directly by calling `New(hashes, power, hasher)` where `hashes` is a `[]uint32`, `power` is `>0` and specifies the probability of false positives `1/2^power`,  and `hasher` is the implements the `hash.Hash32` interface.

### Filter
Once a `Filter` has been created then queries can be made to the `Filter` by calling `Contains(value)` where `value` is a `[]byte`.  It can also be queried using the hash of the `value` by calling `ContainsHash(hash)` where `hash` is an `uint32`.


### Encode/Decode
A filter can be encoded or decoded for storage using `Encode` and `Decode` functions.


### Examples
```
import (

	gcs “github.com/nathanhack/golombcompressedset”
)
…

builder := gcs.Builder(3) //  

…
builder.AddValue(value1)
builder.AddValue(value2)

…

filter := builder.Filter()

fmt.Println(filter.Contains(value))

//Output:
// true

```

```
import "github.com/spaolacci/murmur3"


...

bs := gcs.Encode(filter)

filter2 := gcs.Decode(bs, 3, murmur3.New32())
fmt.Println(filter2.Contains(value1))

```






