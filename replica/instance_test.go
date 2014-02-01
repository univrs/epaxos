package replica

import (
	"fmt"
	"testing"

	"github.com/go-epaxos/epaxos/data"
	"github.com/go-epaxos/epaxos/test"
	"github.com/stretchr/testify/assert"
)

var _ = fmt.Printf

// **************************
// **** COMMON ROUTINE ******
// **************************

func commonTestlibExampleCommands() data.Commands {
	return data.Commands{
		data.Command("hello"),
	}
}

func commonTestlibExampleInstance() *Instance {
	r := New(0, 5, new(test.DummySM))
	i := NewInstance(r, conflictNotFound+1)
	i.cmds = commonTestlibExampleCommands()
	return i
}

func commonTestlibExampleNilStatusInstance() *Instance {
	i := commonTestlibExampleInstance()
	i.status = nilStatus
	return i
}
func commonTestlibExamplePreAcceptedInstance() *Instance {
	i := commonTestlibExampleInstance()
	i.status = preAccepted
	i.ballot = i.replica.makeInitialBallot()
	return i
}
func commonTestlibExampleAcceptedInstance() *Instance {
	i := commonTestlibExampleInstance()
	i.status = accepted
	i.ballot = i.replica.makeInitialBallot()
	return i
}
func commonTestlibExampleCommittedInstance() *Instance {
	i := commonTestlibExampleInstance()
	i.status = committed
	i.ballot = i.replica.makeInitialBallot()
	return i
}
func commonTestlibExamplePreParingInstance() *Instance {
	i := commonTestlibExampleInstance()
	i.status = preparing
	i.ballot = i.replica.makeInitialBallot()
	return i
}

func TestNewInstance(t *testing.T) {
	expectedReplicaId := uint8(0)
	expectedInstanceId := uint64(1)
	r := New(expectedReplicaId, 5, new(test.DummySM))
	i := NewInstance(r, expectedInstanceId)
	assert.Equal(t, i.replica.Id, expectedReplicaId)
	assert.Equal(t, i.id, expectedInstanceId)
	assert.Equal(t, i.deps, i.replica.makeInitialDeps())
}

// ************************
// ****** Nil Status ******
// ************************

// If a nilstatus instance receives propose, it should change its status to
// preaccepted, return (broadcastAction, pre-accept message) and setup relevant
// information.
// The instance should also be ready to receive pre-accept reply. That means the
// relevant info should be set.
func TestNilStatusProcessPropose(t *testing.T) {
	p := &data.Propose{
		Cmds: commonTestlibExampleCommands(),
	}

	instWithBallot := commonTestlibExampleNilStatusInstance()
	instWithBallot.ballot = instWithBallot.replica.makeInitialBallot()
	// test panics not freshly created nilStatus instance
	assert.Panics(t, func() { instWithBallot.nilStatusProcess(p) })

	// test panics instance's status is not nilStatus
	preAcceptedInstance := commonTestlibExamplePreAcceptedInstance()
	assert.Panics(t, func() { preAcceptedInstance.nilStatusProcess(p) })

	i := commonTestlibExampleNilStatusInstance()
	// test panics empty propose
	assert.Panics(t, func() { i.nilStatusProcess(&data.Propose{}) })

	action, m := i.nilStatusProcess(p)
	if !assert.IsType(t, &data.PreAccept{}, m) {
		t.Fatal("")
	}

	pa := m.(*data.PreAccept)
	assert.Equal(t, i.status, preAccepted)
	assert.Equal(t, action, fastQuorumAction)

	assert.True(t, assert.ObjectsAreEqual(pa, &data.PreAccept{
		ReplicaId:  i.replica.Id,
		InstanceId: i.id,
		Cmds:       commonTestlibExampleCommands(),
		Seq:        0,
		Deps:       i.deps,
		Ballot:     i.replica.makeInitialBallot(),
	}))

	assert.Equal(t, i.info.preAcceptCount, 0)
	assert.True(t, i.info.isFastPath)
}

func TestNilStatusProcessPreAccept(t *testing.T) {
}

func TestNilStatusProcessAccept(t *testing.T) {
}

func TestNilStatusProcessCommit(t *testing.T) {
}

func TestNilStatusOnCommitDependency(t *testing.T) {
}

// ************************
// ****** PREACCEPTED *****
// ************************

// TestPreAcceptedProcessStatus tests
// if preAcceptedProcess panics as expected
func TestPreAcceptedProcessStatus(t *testing.T) {
	inst := commonTestlibExampleAcceptedInstance()
	ac := &data.Accept{}
	assert.Panics(t, func() { inst.preAcceptedProcess(ac) })
}

// When preAccepted instance receives a pre-accept,
// If ballot > self ballot,
// if preAcceptedProcess accepts or
// rejects the PreAccept message correctly
func TestPreAcceptedProcessPreAccept(t *testing.T) {
	inst := commonTestlibExamplePreAcceptedInstance()
	instanceBallot := data.NewBallot(2, 3, inst.replica.Id)
	smallerBallot := data.NewBallot(2, 2, inst.replica.Id)
	largerBallot := data.NewBallot(2, 4, inst.replica.Id)

	inst.ballot = instanceBallot

	// PreAccept with smaller ballot
	p := &data.PreAccept{
		Ballot: smallerBallot,
	}
	action, reply := inst.preAcceptedProcess(p)

	assert.Equal(t, action, replyAction)
	assert.Equal(t, reply, &data.PreAcceptReply{
		Ok:         false,
		ReplicaId:  inst.replica.Id,
		InstanceId: inst.id,
		Ballot:     instanceBallot,
	})

	expectedSeq := uint32(42)
	expectedDeps := data.Dependencies{1, 0, 0, 8, 6}

	// PreAccept with larger ballot
	p = &data.PreAccept{
		Cmds:   commonTestlibExampleCommands(),
		Deps:   expectedDeps,
		Seq:    expectedSeq,
		Ballot: largerBallot,
	}
	action, reply = inst.preAcceptedProcess(p)

	assert.Equal(t, action, replyAction)
	assert.Equal(t, reply, &data.PreAcceptReply{
		Ok:         true,
		ReplicaId:  inst.replica.Id,
		InstanceId: inst.id,
		Seq:        expectedSeq,
		Deps:       expectedDeps,
		Ballot:     largerBallot,
	})
}

// **********************
// *****  ACCEPTED ******
// **********************

func TestAcceptedProcessPrepare(t *testing.T) {
}

// **********************
// ***** COMMITTED ******
// **********************

// When a committed instance receives:
// * pre-accept reply,
// it should ignore it
func TestCommittedProcessWithNoAction(t *testing.T) {
	// create a committed instance
	i := commonTestlibExampleCommittedInstance()
	// send a pre-accept message to it
	pa := &data.PreAcceptReply{}
	action, m := i.committedProcess(pa)

	// expect:
	// - action: NoAction
	// - message: nil
	// - instance not changed
	assert.Equal(t, action, noAction)
	assert.Nil(t, m)
}

// If a committed instance receives accept, it will reject it.
func TestCommittedProcessWithRejcetAccept(t *testing.T) {
	// create a committed instance
	inst := commonTestlibExampleCommittedInstance()
	// send an Accept message to it
	a := &data.Accept{}
	action, m := inst.committedProcess(a)

	// expect:
	// - action: replyAction
	// - message: AcceptReply with ok == false, ballot == i.ballot
	assert.Equal(t, action, replyAction)
	assert.Equal(t, m, &data.AcceptReply{
		Ok:         false,
		ReplicaId:  inst.replica.Id,
		InstanceId: inst.id,
		Ballot:     inst.ballot.GetCopy(),
	})
}

// if a committed instance receives prepare with
// - larger ballot, reply ok = true with large ballot
// - smaller ballot, reply ok = true with small ballot
func TestCommittedProcessWithHandlePrepare(t *testing.T) {
	// create a committed instance
	inst := commonTestlibExampleCommittedInstance()

	// create small and large ballots
	smallBallot := inst.replica.makeInitialBallot()
	largeBallot := smallBallot.GetIncNumCopy()

	// send a Prepare message to it
	p := &data.Prepare{
		Ballot:          largeBallot,
		NeedCmdsInReply: true,
	}
	expectedReply := &data.PrepareReply{
		Ok:         true,
		ReplicaId:  inst.replica.Id,
		InstanceId: inst.id,
		Status:     committed,
		Cmds:       inst.cmds,
		Deps:       inst.deps,
	}

	// expect:
	// - action: replyAction
	// - message: AcceptReply with ok == true, ballot == message ballot

	// handle larger ballot
	action, m := inst.committedProcess(p)
	assert.Equal(t, action, replyAction)

	expectedReply.Ballot = largeBallot.GetCopy()
	expectedReply.FromInitialLeader = true
	assert.Equal(t, m, expectedReply)

	// handle smaller ballot
	inst.ballot = largeBallot
	p.Ballot = smallBallot
	_, m = inst.committedProcess(p)

	expectedReply.Ballot = smallBallot.GetCopy()
	expectedReply.FromInitialLeader = false
	assert.Equal(t, m, expectedReply)
}

// committed instance should reject pre-accept
func TestCommittedProcessWithRejectPreAccept(t *testing.T) {
	// create a committed instance
	inst := commonTestlibExampleCommittedInstance()

	// send a PreAccept message to it
	p := &data.PreAccept{}
	action, m := inst.committedProcess(p)

	// expect:
	// - action: replyAction
	// - message: PreAcceptReply with ok == false
	assert.Equal(t, action, replyAction)
	assert.Equal(t, m, &data.PreAcceptReply{
		Ok:         false,
		ReplicaId:  inst.replica.Id,
		InstanceId: inst.id,
		Ballot:     inst.ballot,
	})
}

func TestCommittedProccessWithPanic(t *testing.T) {
	// create a accepted instance
	inst := commonTestlibExampleAcceptedInstance()
	p := &data.Propose{}
	// expect:
	// - action: will panic if is not at committed status
	assert.Panics(t, func() { inst.committedProcess(p) })

	// create a committed instance
	inst = commonTestlibExampleCommittedInstance()

	// expect:
	// - action: will panic if receiving propose
	assert.Panics(t, func() { inst.committedProcess(p) })
}

// **********************
// ***** PREPARING ******
// **********************

// **********************
// ***** REJECTIONS *****
// **********************

// TestRejections tests correctness of all rejection functions.
// These rejection functions have reply fields in common:
// {
//   ok: false
//   ballot: self ballot
//   Ids
// }
func TestRejections(t *testing.T) {
	inst := commonTestlibExampleInstance()
	expectedBallot := data.NewBallot(1, 2, inst.replica.Id)
	inst.ballot = expectedBallot.GetCopy()

	// reject with PreAcceptReply
	action, par := inst.rejectPreAccept()
	assert.Equal(t, action, replyAction)

	assert.True(t, assert.ObjectsAreEqual(par, &data.PreAcceptReply{
		Ok:         false,
		ReplicaId:  inst.replica.Id,
		InstanceId: inst.id,
		Ballot:     expectedBallot,
	}))

	// reject with AcceptReply
	action, ar := inst.rejectAccept()
	assert.Equal(t, action, replyAction)

	assert.True(t, assert.ObjectsAreEqual(ar, &data.AcceptReply{
		Ok:         false,
		ReplicaId:  inst.replica.Id,
		InstanceId: inst.id,
		Ballot:     expectedBallot,
	}))

	// reject with PrepareReply
	action, ppr := inst.rejectPrepare()
	assert.Equal(t, action, replyAction)
	assert.True(t, assert.ObjectsAreEqual(ppr, &data.PrepareReply{
		Ok:         false,
		ReplicaId:  inst.replica.Id,
		InstanceId: inst.id,
		Ballot:     expectedBallot,
	}))
}

// ******************************
// ******* HANDLE MESSAGE *******
// ******************************

// It's testing `handleprepare` will return (replyaction, correct prepare-reply)
// If we send prepare which sets `needcmdsinreply` true, it should return cmds in reply.
func TestHandlePrepare(t *testing.T) {
	i := commonTestlibExampleCommittedInstance()
	i.ballot = i.replica.makeInitialBallot()
	i.deps = data.Dependencies{3, 4, 5, 6, 7}

	largerBallot := i.ballot.GetIncNumCopy()

	// NeedCmdsInReply == false
	prepare := &data.Prepare{
		ReplicaId:       i.replica.Id,
		InstanceId:      i.id,
		Ballot:          largerBallot,
		NeedCmdsInReply: false,
	}

	action, reply := i.handlePrepare(prepare)

	assert.Equal(t, action, replyAction)
	assert.Equal(t, reply, &data.PrepareReply{
		Ok:                true,
		ReplicaId:         0,
		InstanceId:        1,
		Status:            committed,
		Cmds:              nil,
		Deps:              i.deps.GetCopy(),
		Ballot:            prepare.Ballot,
		FromInitialLeader: true,
	})

	// NeedCmdsInReply == true
	prepare.NeedCmdsInReply = true
	i.cmds = commonTestlibExampleCommands()
	i.ballot = i.replica.makeInitialBallot()

	action, reply = i.handlePrepare(prepare)
	assert.Equal(t, action, replyAction)
	// test the reply
	assert.Equal(t, reply.Cmds, i.cmds)
}

// TestHandleCommit tests the functionality of handleCommit
// on success: handleCommit returns a no act and nil message,
// besides, the instances' status is set to commited.
// on failure: otherwise
func TestHandleCommit(t *testing.T) {
}

// TestCheckStatus tests the behaviour of checkStatus,
// - If instance is not at any status listed in checking function, it should panic.
// - If instance is at status listed, it should not panic.
func TestCheckStatus(t *testing.T) {
	i := &Instance{
		status: committed,
	}

	assert.Panics(t, func() { i.checkStatus(preAccepted, accepted, preparing) })
	assert.NotPanics(t, func() { i.checkStatus(committed) })
}