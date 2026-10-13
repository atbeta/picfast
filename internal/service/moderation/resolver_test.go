package moderation

import (
	"testing"

	"github.com/atbeta/picfast/internal/testutil"
)

func TestResolverCachesSameMode(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	defer pool.Close()

	r := NewResolver(db)
	m1 := r.For("disabled")
	m2 := r.For("disabled")
	if m1 != m2 {
		t.Fatal("expected same moderator instance for repeated same mode")
	}
	if m1.Name() != "noop" {
		t.Fatalf("disabled mode Name() = %q, want noop", m1.Name())
	}
}

func TestResolverRebuildsOnModeChange(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	defer pool.Close()

	r := NewResolver(db)
	m1 := r.For("disabled")
	m2 := r.For("manual")
	if m1 == m2 {
		t.Fatal("expected a new moderator after mode change")
	}
	if _, ok := m2.(*ManualModerator); !ok {
		t.Fatalf("manual mode moderator type = %T, want *ManualModerator", m2)
	}
	// Switching back rebuilds again.
	m3 := r.For("disabled")
	if _, ok := m3.(*NoopModerator); !ok {
		t.Fatalf("disabled mode moderator type = %T, want *NoopModerator", m3)
	}
}

func TestResolverInvalidModeFallsBackToNoop(t *testing.T) {
	pool, db := testutil.SetupDB(t)
	defer pool.Close()

	r := NewResolver(db)
	m := r.For("bogus")
	if _, ok := m.(*NoopModerator); !ok {
		t.Fatalf("invalid mode moderator type = %T, want *NoopModerator", m)
	}
	// Repeated invalid requests return the cached noop without rebuilding.
	if again := r.For("bogus"); again != m {
		t.Fatal("expected cached noop for repeated invalid mode")
	}
	// Empty mode resolves to disabled/noop too.
	if empty := r.For(""); empty.Name() != "noop" {
		t.Fatalf("empty mode Name() = %q, want noop", empty.Name())
	}
}
