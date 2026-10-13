package pglock

import (
	"context"
	"testing"

	"github.com/atbeta/picfast/internal/testutil"
)

func TestAcquireExcludesConcurrentHolder(t *testing.T) {
	pool, _ := testutil.SetupDB(t)
	ctx := context.Background()

	release, ok, err := Acquire(ctx, pool, KeyRecalcPHash)
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	if !ok {
		t.Fatal("first Acquire should hold the lock")
	}

	if _, ok, err := Acquire(ctx, pool, KeyRecalcPHash); err != nil {
		t.Fatalf("second Acquire: %v", err)
	} else if ok {
		t.Fatal("second Acquire should not acquire while the lock is held")
	}

	release()
	release() // idempotent — must not panic or error

	release3, ok, err := Acquire(ctx, pool, KeyRecalcPHash)
	if err != nil {
		t.Fatalf("third Acquire: %v", err)
	}
	if !ok {
		t.Fatal("lock should be free after release")
	}
	release3()
}

func TestTryWithLockReportsContention(t *testing.T) {
	pool, _ := testutil.SetupDB(t)
	ctx := context.Background()

	release, ok, err := Acquire(ctx, pool, KeyJobsCleanup)
	if err != nil || !ok {
		t.Fatalf("Acquire: ok=%v err=%v", ok, err)
	}
	defer release()

	ran := false
	acquired, err := TryWithLock(ctx, pool, KeyJobsCleanup, func(context.Context) error {
		ran = true
		return nil
	})
	if err != nil {
		t.Fatalf("TryWithLock: %v", err)
	}
	if acquired || ran {
		t.Fatal("TryWithLock ran fn while the lock was held elsewhere")
	}
}
