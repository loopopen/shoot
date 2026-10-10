package status

import "testing"

func TestScan(t *testing.T) {
	t.Run("bytes", func(t *testing.T) {
		var status Status
		if err := status.Scan([]byte("Paid")); err != nil {
			t.Fatal(err)
		}
		if status != StatusPaid {
			t.Fatalf("got %v", status)
		}
	})

	t.Run("string", func(t *testing.T) {
		var status Status
		if err := status.Scan("Paid"); err != nil {
			t.Fatal(err)
		}
		if status != StatusPaid {
			t.Fatalf("got %v", status)
		}
	})

	t.Run("null", func(t *testing.T) {
		status := StatusPaid
		if err := status.Scan(nil); err != nil {
			t.Fatal(err)
		}
		if status != StatusPending {
			t.Fatalf("got %v", status)
		}
	})
}
