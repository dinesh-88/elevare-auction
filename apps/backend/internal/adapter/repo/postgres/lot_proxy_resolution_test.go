package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestResolveProxyOutcomeSingleProxyBidder(t *testing.T) {
	userA := uuid.New()
	participants := []lotBidParticipant{{UserID: userA, EffectiveMax: 10000, FirstAt: time.Now().UTC()}}
	outcome := resolveProxyOutcome(participants, 100, 1100, 1100)

	if outcome.WinnerUserID != userA {
		t.Fatalf("expected winner %s, got %s", userA, outcome.WinnerUserID)
	}
	if outcome.VisiblePrice != 1100 {
		t.Fatalf("expected visible price 1100, got %d", outcome.VisiblePrice)
	}
}

func TestResolveProxyOutcomeTwoProxyBidders(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	now := time.Now().UTC()
	participants := []lotBidParticipant{
		{UserID: userA, EffectiveMax: 10000, FirstAt: now},
		{UserID: userB, EffectiveMax: 9000, FirstAt: now.Add(time.Second)},
	}
	outcome := resolveProxyOutcome(participants, 100, 8100, 8100)

	if outcome.WinnerUserID != userA {
		t.Fatalf("expected winner %s, got %s", userA, outcome.WinnerUserID)
	}
	if outcome.VisiblePrice != 9100 {
		t.Fatalf("expected visible price 9100, got %d", outcome.VisiblePrice)
	}
}

func TestResolveProxyOutcomeManualBidTriggersProxy(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	now := time.Now().UTC()
	participants := []lotBidParticipant{
		{UserID: userA, EffectiveMax: 10000, FirstAt: now},
		{UserID: userB, EffectiveMax: 8500, FirstAt: now.Add(time.Second)},
	}
	outcome := resolveProxyOutcome(participants, 100, 8100, 8500)

	if outcome.WinnerUserID != userA {
		t.Fatalf("expected winner %s, got %s", userA, outcome.WinnerUserID)
	}
	if outcome.VisiblePrice != 8600 {
		t.Fatalf("expected visible price 8600, got %d", outcome.VisiblePrice)
	}
}

func TestResolveProxyOutcomeHigherManualBeatsProxy(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	now := time.Now().UTC()
	participants := []lotBidParticipant{
		{UserID: userB, EffectiveMax: 11000, FirstAt: now.Add(time.Second)},
		{UserID: userA, EffectiveMax: 10000, FirstAt: now},
	}
	outcome := resolveProxyOutcome(participants, 100, 8100, 11000)

	if outcome.WinnerUserID != userB {
		t.Fatalf("expected winner %s, got %s", userB, outcome.WinnerUserID)
	}
	if outcome.VisiblePrice != 11000 {
		t.Fatalf("expected visible price 11000, got %d", outcome.VisiblePrice)
	}
}

func TestResolveProxyOutcomeTieAtMaxDoesNotIncreaseVisiblePrice(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	now := time.Now().UTC()
	participants := []lotBidParticipant{
		{UserID: userA, EffectiveMax: 10000, FirstAt: now},
		{UserID: userB, EffectiveMax: 10000, FirstAt: now.Add(time.Second)},
	}
	outcome := resolveProxyOutcome(participants, 100, 10000, 10000)
	if outcome.WinnerUserID != userA {
		t.Fatalf("expected earlier bidder %s to win tie, got %s", userA, outcome.WinnerUserID)
	}
	if outcome.VisiblePrice != 10000 {
		t.Fatalf("expected visible price 10000, got %d", outcome.VisiblePrice)
	}
}

func TestResolveSoftCloseEndTimeExtendsInsideWindow(t *testing.T) {
	now := time.Now().UTC()
	end := now.Add(20 * time.Second)
	updated, extended := resolveSoftCloseEndTime(&end, now, true, 30, 120)
	if !extended {
		t.Fatal("expected soft close to extend")
	}
	if updated == nil {
		t.Fatal("expected updated end time")
	}
	expected := now.Add(120 * time.Second)
	if !updated.Equal(expected) {
		t.Fatalf("expected end %s, got %s", expected, *updated)
	}
}

func TestResolveSoftCloseEndTimeNoExtensionOutsideWindow(t *testing.T) {
	now := time.Now().UTC()
	end := now.Add(2 * time.Minute)
	updated, extended := resolveSoftCloseEndTime(&end, now, true, 30, 120)
	if extended {
		t.Fatal("did not expect extension")
	}
	if updated == nil || !updated.Equal(end) {
		t.Fatalf("expected unchanged end time %s", end)
	}
}

func TestConcurrentBidSequenceSimulation(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	inc := int64(100)
	current := int64(1000)

	minA := current + inc
	outA := resolveProxyOutcome([]lotBidParticipant{{UserID: userA, EffectiveMax: 10000}}, inc, minA, minA)
	if outA.WinnerUserID != userA || outA.VisiblePrice != 1100 {
		t.Fatalf("unexpected first outcome winner=%s price=%d", outA.WinnerUserID, outA.VisiblePrice)
	}
	current = outA.VisiblePrice

	minB := current + inc
	outB := resolveProxyOutcome([]lotBidParticipant{{UserID: userA, EffectiveMax: 10000}, {UserID: userB, EffectiveMax: 9000}}, inc, minB, minB)
	if outB.WinnerUserID != userA {
		t.Fatalf("expected user A to stay winner, got %s", outB.WinnerUserID)
	}
	if outB.VisiblePrice != 9100 {
		t.Fatalf("expected final price 9100, got %d", outB.VisiblePrice)
	}
}
