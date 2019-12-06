package main

import (
	"testing"
	"time"
)

func newFakeState() *terminatorState {
	now := time.Now()
	then := now.Add(-4 * time.Hour)

	i1 := &instance{
		instanceID:   "i-EatTacos",
		terminatedAt: then}

	i2 := &instance{
		instanceID:   "i-LikeBananas",
		terminatedAt: now}

	t := &terminatorState{}

	t.terminated = append(t.terminated, i1)
	t.terminated = append(t.terminated, i2)

	return t
}

func TestNodeTerminator_ExpireTerminatedInstances(t *testing.T) {
	state := newFakeState()

	state.expireTerminatedInstances()

	if len(state.terminated) != 1 {
		t.Errorf("failed to expired already terminated instance")
	}
}

func TestNodeTerminator_OkToTerminate(t *testing.T) {
	state := newFakeState()

	i := &instance{
		instanceID:   "i-EatTacos",
		terminatedAt: time.Now()}

	if state.okToTerminate(i.instanceID) {
		t.Errorf("expected false but got true")
	}
}
