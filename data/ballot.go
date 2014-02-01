package data

const (
	// Ballot has a format like:
	// Epoch   | Number  | ReplicaId
	// 20 bits | 36 bits | 8 bits
	ballotEpochWidth     uint = 20
	ballotNumberWidth    uint = 36
	ballotReplicaIdWidth uint = 8

	ballotEpochMask     uint64 = ((1 << ballotEpochWidth) - 1) << (ballotNumberWidth + ballotReplicaIdWidth)
	ballotNumberMask    uint64 = ((1 << ballotNumberWidth) - 1) << (ballotReplicaIdWidth)
	ballotReplicaIdMask uint64 = (1 << ballotReplicaIdWidth) - 1
)

type Ballot struct {
	epoch     uint32
	number    uint64
	replicaId uint8
}

func NewBallot(epoch uint32, number uint64, replicId uint8) *Ballot {
	return &Ballot{
		epoch,
		number,
		replicId,
	}
}

func (b *Ballot) ToUint64() uint64 {
	return ((uint64(b.epoch) << (ballotNumberWidth + ballotReplicaIdWidth)) |
		(b.number << ballotReplicaIdWidth) |
		uint64(b.replicaId))
}

func (b *Ballot) FromUint64(num uint64) {
	b.epoch = uint32((num & ballotEpochMask) >> (ballotNumberWidth + ballotReplicaIdWidth))
	b.number = ((num & ballotNumberMask) >> ballotReplicaIdWidth)
	b.replicaId = uint8(num & ballotReplicaIdMask)
}

func (b *Ballot) Compare(other *Ballot) int {
	if b == nil || other == nil {
		panic("Compare: ballot should not be nil")
	}
	if b.epoch > other.epoch {
		return 1
	}
	if b.epoch < other.epoch {
		return -1
	}
	if b.number > other.number {
		return 1
	}
	if b.number < other.number {
		return -1
	}
	if b.replicaId > other.replicaId {
		return 1
	}
	if b.replicaId < other.replicaId {
		return -1
	}

	return 0
}

func (b *Ballot) IncNumber() {
	b.number++
}

func (b *Ballot) GetNumber() uint64 {
	return b.number
}

func (b *Ballot) SetReplicaId(rId int) {
	b.replicaId = uint8(rId)
}

func (b *Ballot) GetIncNumCopy() *Ballot {
	return &Ballot{
		b.epoch,
		b.number + 1,
		b.replicaId,
	}
}

func (b *Ballot) IsInitialBallot() bool {
	return b.number == 0
}

func (b *Ballot) GetCopy() *Ballot {
	return &Ballot{
		b.epoch,
		b.number,
		b.replicaId,
	}
}