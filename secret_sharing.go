package mpc

import (
	"crypto/rand"
	"fmt"
	"math/big"
	mr "math/rand"
)

var one = new(big.Int).SetInt64(1)

// GenerateShares generates `n` shares such that (`s_1` + ... + `s_n`) mod `M` = `secret`
func GenerateShares(secret int64, n int, M *big.Int) []*big.Int {
	s := new(big.Int).SetInt64(secret)
	sum := new(big.Int)
	shares := make([]*big.Int, n)
	j := mr.Intn(n)
	for i := 0; i < n; i++ {
		if i != j {
			shares[i] = getRandom(M)
			sum.Add(sum, shares[i])
		}
	}
	shares[j] = s.Sub(s, sum).Mod(s, M)
	return shares
}

// GenerateBeaverTriplet generate three numbers `w` = `uv` mod `N`
func GenerateBeaverTriplet(N *big.Int) [3]*big.Int {
	var triplet [3]*big.Int

	x, y := getRandom(N), getRandom(N)

	if mr.Intn(2) > 0 {
		// u = x, v = y, w = xy mod N
		w := new(big.Int).Mul(x, y)
		triplet[0] = w.Mod(w, N)
		triplet[1] = x
		triplet[2] = y
	} else {
		// v = x, w = y, u = (xy^-1)^-1 mod N
		u := new(big.Int).Mul(x, new(big.Int).ModInverse(y, N))
		triplet[0] = y
		triplet[1] = x
		triplet[2] = u.ModInverse(u, N)
	}

	return triplet
}

// Message format exchanged between protocol parties
type Message struct {
	id int        // party id in the protocol round/instance
	v  []*big.Int // intermediate values
}

// BroadcastAgent used as single message exchange channel
type BroadcastAgent struct {
	parties  map[int]chan<- []Message
	messages []Message
}

// NewBroadcastAgent returns the pointer to a new BroadcastAgent instance
func NewBroadcastAgent(n int) *BroadcastAgent {
	return &BroadcastAgent{}
}

// Broadcast a message to all prototol parties
func (ba *BroadcastAgent) Broadcast(round int, m Message) {
	ba.messages = append(ba.messages, m)
	if len(ba.messages) == len(ba.parties) {
		for _, ch := range ba.parties {
			ch <- ba.messages
		}
	}
}

// Subscribe to the BA in order to receive protocol messages
func (ba *BroadcastAgent) Subscribe(id int, ch chan chan<- []Message) {
	c, ok := <-ch
	if ok {
		ba.parties[id] = c
	}
}

// Parameters needed to run the secret sharing protocols on the commodity model
type Parameters struct {
	n       int         // number of computing parties
	M       *big.Int    // modulus for circular group operations
	triplet [3]*big.Int // beaver's triplep, for multiplication optmization (this is the commodity model)
	assinc  int         // assincronous bit, to mark the protocol party with different multiplication setup on the commodity model
}

// NewParameters creates default protocol parameters for testing
func NewParameters(bitlen int) *Parameters {
	m, err := rand.Prime(rand.Reader, bitlen)
	if err != nil {
		panic("Could not generate randon prime")
	}
	n := 2
	return &Parameters{
		n:       n,
		M:       m,
		triplet: GenerateBeaverTriplet(m),
		assinc:  mr.Intn(n),
	}
}

// IntProtocol represents the default structure for secret-sharing protocol
// to perform computations of integers
type IntProtocol interface {
	Setup(param *Parameters, args []int64) error
	Run() error
	Output() int64
}

// Party models a protocol party
type Party interface {
	//Instanciate(id int, param *Parameters, ba *BroadcastAgent) error
	Run()
}

// getRandom generates a random Int `r` such that `r < N` and `gcd(r, N) = 1`
func getRandom(N *big.Int) *big.Int {
	gcd := new(big.Int)
	r := new(big.Int)
	err := fmt.Errorf("")

	for gcd.Cmp(one) != 0 {
		r, err = rand.Int(rand.Reader, N)
		if err != nil {
			panic("Error while reading crypto/rand")
		}

		gcd = new(big.Int).GCD(nil, nil, r, N)
	}
	return r
}
