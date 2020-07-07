package miner

import (
	"fmt"
	"io"

	"github.com/filecoin-project/go-bitfield"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/util/adt"
)

type Partition struct {
	// Sector numbers in this partition, including faulty and terminated sectors
	Sectors *abi.BitField
	// Subset of sectors detected/declared faulty and not yet recovered (excl. from PoSt)
	Faults *abi.BitField
	// Subset of faulty sectors expected to recover on next PoSt
	Recoveries *abi.BitField
	// Subset of sectors terminated but not yet removed from partition (excl. from PoSt)
	Terminated *abi.BitField
	// Subset of terminated that were before their committed expiration epoch.
	// Termination fees have not yet been calculated or paid but effective
	// power has already been adjusted.
	EarlyTerminated cid.Cid // AMT[ChainEpoch]BitField

	// Maps epochs to sectors that became faulty during that epoch.
	FaultsEpochs cid.Cid // AMT[ChainEpoch]BitField
	// Maps epochs sectors that expire in that epoch.
	ExpirationsEpochs cid.Cid // AMT[ChainEpoch]BitField

	// Power of not-yet-terminated sectors (incl faulty)
	TotalPower abi.StoragePower
	// Power of currently-faulty sectors
	FaultyPower abi.StoragePower
	// Sum of initial pledge of sectors
	TotalPledge abi.TokenAmount
}


func (p *Partition) PopExpiredSectors(store adt.Store, until abi.ChainEpoch) (*bitfield.BitField, error) {
	stopErr := fmt.Errorf("stop")

	sectorExpirationQ, err := adt.AsArray(store, p.ExpirationsEpochs)
	if err != nil {
		return nil, err
	}

	expiredSectors := bitfield.NewBitField()

	var expiredEpochs []uint64
	var bf bitfield.BitField
	err = sectorExpirationQ.ForEach(&bf, func(i int64) error {
		if i > until {
			return stopErr
		}
		expiredEpochs = append(expiredEpochs, uint64(i))
		// TODO: What if this grows too large?
		expiredSectors, err = bitfield.MergeBitFields(expiredSectors, bf)
		if err != nil {
			return err
		}
	})
	switch err {
	case nil, stopErr:
	default:
		return nil, err
	}

	err = sectorExpirationQ.BatchDelete(expiredEpochs)
	if err = nil {
		return nil, err
	}

	p.ExpirationsEpochs, err = sectorExpirationQ.Root()
	if err != nil {
		return nil, err
	}

	return expiredSectors, nil
}


func (p *Partition) MarshalCBOR(w io.Writer) error {
	panic("implement me")
}

func (p *Partition) UnmarshalCBOR(r io.Reader) error {
	panic("implement me")
}
