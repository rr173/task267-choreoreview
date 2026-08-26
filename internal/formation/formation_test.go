package formation

import (
	"testing"

	"task267-choreoreview/internal/model"
)

func TestFormationConflict(t *testing.T) {
	edges := []*model.FormationRelation{
		{ID: 1, FromDancer: 1, ToDancer: 2, Relation: "line", BeatStart: 1, BeatEnd: 4, Status: model.FormationStatusRaw},
		{ID: 2, FromDancer: 2, ToDancer: 1, Relation: "diagonal", BeatStart: 1, BeatEnd: 4, Status: model.FormationStatusRaw},
	}
	check := New().Check(edges)
	if check.ConflictCount != 1 {
		t.Fatalf("conflicts=%d want 1", check.ConflictCount)
	}
}

func TestFormationNoConflictDifferentBeats(t *testing.T) {
	edges := []*model.FormationRelation{
		{ID: 1, FromDancer: 1, ToDancer: 2, Relation: "line", BeatStart: 1, BeatEnd: 2, Status: model.FormationStatusRaw},
		{ID: 2, FromDancer: 2, ToDancer: 1, Relation: "diagonal", BeatStart: 3, BeatEnd: 4, Status: model.FormationStatusRaw},
	}
	check := New().Check(edges)
	if check.ConflictCount != 0 {
		t.Fatalf("conflicts=%d want 0", check.ConflictCount)
	}
}

func TestFormationSkipsRejected(t *testing.T) {
	edges := []*model.FormationRelation{
		{ID: 1, FromDancer: 1, ToDancer: 2, Relation: "line", BeatStart: 1, BeatEnd: 4, Status: model.FormationStatusRejected},
		{ID: 2, FromDancer: 2, ToDancer: 1, Relation: "diagonal", BeatStart: 1, BeatEnd: 4, Status: model.FormationStatusRaw},
	}
	check := New().Check(edges)
	if check.ConflictCount != 0 {
		t.Fatalf("rejected edge should be ignored, conflicts=%d", check.ConflictCount)
	}
}
