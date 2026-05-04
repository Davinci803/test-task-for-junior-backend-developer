package task

import "testing"

func TestStatusValid(t *testing.T) {
	if !StatusNew.Valid() || !StatusInProgress.Valid() || !StatusDone.Valid() {
		t.Fatalf("expected known statuses to be valid")
	}
	if Status("bad").Valid() {
		t.Fatalf("expected unknown status to be invalid")
	}
}
