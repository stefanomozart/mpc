package mpc

import (
	"fmt"
	"math/big"
	"sync"
	"time"
)

// DistributedIntMean protocol for secure multi-party integer mean computation
type DistributedIntMean struct {
	param      *Parameters       // number of parties, modulus, beaver triplet
	parcels    []int64           // parcels to compose the mean
	parties    []chan []*big.Int // Channels to computing parties
	shares     [][]*big.Int      // Shares of the secret values (parcels)
	ba         *BroadcastAgent   // Broadcast channel among parties
	wg         sync.WaitGroup
	result     int64 // protocol result for output
	setup, run time.Duration
}

// NewDistributedIntMean creates a new Distributed Mean Protocol instance
func NewDistributedIntMean() IntProtocol {
	return &DistributedIntMean{}
}

// Setup the Distributed Integer Mean protocol with given hiper parameters
// (`n` parties, performing modular operations with `N` modulus) and the
// given parcels to be unpacked from the generic varadiac argument
func (dim *DistributedIntMean) Setup(param *Parameters, args []int64) error {
	start := time.Now()

	// Validate parameters
	if param.n < 2 {
		return fmt.Errorf("The number of computing parties must be equal to or greater than 2")
	}

	if len(args) < 1 {
		return fmt.Errorf("At least one number (int64 argument) is required in order to compute a mean")
	}

	dim.param = param
	dim.parcels = args
	dim.shares = make([][]*big.Int, len(args))

	// Generate secret shares to be sent to protocol parties
	for i, p := range dim.parcels {
		dim.shares[i] = GenerateShares(p, dim.param.n, dim.param.M)
	}

	// Instantiate a broacas agent, for protocol parties message exchange
	dim.ba = NewBroadcastAgent(dim.param.n)

	for i := 0; i < dim.param.n; i++ {
		dim.parties = append(dim.parties, make(chan []*big.Int))
		dim.wg.Add(1)
		p := &DistributedIntMeanParty{
			id:     i,
			param:  param,
			ba:     dim.ba,
			client: dim.parties[i],
			wg:     &dim.wg,
		}

		go p.Run()
	}

	dim.setup = time.Since(start)
	return nil
}

// Run the protocol
func (dim *DistributedIntMean) Run() error {
	start := time.Now()

	// Send shares to parties
	for i, p := range dim.parties {
		pShares := make([]*big.Int, len(dim.shares))
		for j, s := range dim.shares {
			pShares[j] = s[i]
		}
		p <- pShares
	}

	// Receive result shares from protocol parties
	zs := make([][]*big.Int, dim.param.n)
	for i, p := range dim.parties {
		zs[i] = <-p
	}

	z := new(big.Int)
	sumMod := new(big.Int)
	for _, zi := range zs {
		z.Add(z, zi[0])
		sumMod.Add(sumMod, zi[1])
	}

	lens := new(big.Int).SetInt64(int64(len(dim.parcels)))
	outroMod := new(big.Int).Div(dim.param.M, lens)
	z.Mod(z, outroMod)
	sumMod.Add(sumMod, z)
	sumMod.Div(sumMod, lens)
	sumMod.Mod(sumMod, outroMod)

	dim.result = z.Int64()

	dim.wg.Wait()

	dim.run = time.Since(start)
	return nil
}

// Output the computed computed mean
func (dim *DistributedIntMean) Output() int64 {
	return dim.result
}

// DistributedIntMeanParty is a Secure Distributes Mean Protocol party
type DistributedIntMeanParty struct {
	id      int             // party id in the protocol round/instance
	param   *Parameters     // hiper paramenters (modulus, beaver triplet)
	ba      *BroadcastAgent // communication with other parties
	client  chan []*big.Int // communication with the client
	wg      *sync.WaitGroup // waitgroup to signal end of execution
	parcels []*big.Int      // parcels to be added
	z       *big.Int        // local result
}

// NewDistributedIntMeanParty returns a news instance of the DM party
func NewDistributedIntMeanParty(id int, param *Parameters, ba *BroadcastAgent) Party {
	return &DistributedIntMeanParty{
		id:    id,
		param: param,
		ba:    ba,
	}
}

// Run the party role in the distributed computation
func (dmp *DistributedIntMeanParty) Run() {
	dmp.parcels = <-dmp.client
	lens := new(big.Int).SetInt64(int64(len(dmp.parcels)))

	sum := new(big.Int)
	for _, p := range dmp.parcels {
		sum.Add(sum, p)
	}
	sum.Mod(sum, dmp.param.M)

	//- sumDiv
	m := new(big.Int)
	sum, m = sum.DivMod(sum, lens, m)
	sum.Mod(sum, dmp.param.M)
	m.Mod(m, dmp.param.M)

	//- Send response to client
	dmp.client <- []*big.Int{sum, m}
	dmp.wg.Done()
}
