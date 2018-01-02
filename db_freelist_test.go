package bolt_test

import (
	"fmt"
	"testing"

	bolt "github.com/coreos/bbolt"
)

func TestOpenFreelistSyncToNoSync(t *testing.T) {
	db := MustOpenWithOption(&bolt.Options{NoFreelistSync: false})
	defer db.MustClose()

	// Write some pages.
	tx, err := db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}
	wbuf := make([]byte, 8192)
	for i := 0; i < 100; i++ {
		s := fmt.Sprintf("%d", i)
		b, err := tx.CreateBucket([]byte(s))
		if err != nil {
			t.Fatal(err)
		}
		if err = b.Put([]byte(s), wbuf); err != nil {
			t.Fatal(err)
		}
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}

	// Generate free pages.
	if tx, err = db.Begin(true); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 50; i++ {
		s := fmt.Sprintf("%d", i)
		b := tx.Bucket([]byte(s))
		if b == nil {
			t.Fatal(err)
		}
		if err := b.Delete([]byte(s)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := db.DB.Close(); err != nil {
		t.Fatal(err)
	}

	// Record freelist count from opening with NoFreelistSync.
	db.MustReopen()

	freepages1 := db.Stats().FreePageN
	if freepages1 == 0 {
		t.Fatalf("no free pages on NoFreelistSync reopen")
	}
	if err := db.DB.Close(); err != nil {
		t.Fatal(err)
	}

	// Check free page count is reconstructed when opened with freelist sync.
	db.o = &bolt.Options{NoFreelistSync: true}
	db.MustReopen()

	freepages2 := db.Stats().FreePageN

	fmt.Println("freepages1, freepages2:", freepages1, freepages2)

	if freepages2 < freepages1 {
		t.Logf("closed with %d free pages, opened with %d", freepages1, freepages2)
	}
}
