// store 层端到端测试：建表、CRUD、唯一约束、状态机。
package store

import (
	"os"
	"path/filepath"
	"testing"

	"task267-choreoreview/internal/model"
)

func openTest(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	st, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestDanceStoreCRUD(t *testing.T) {
	st := openTest(t)
	d, err := st.Dances.Create("花灯", "云南", 3, "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if d.Status != model.DanceStatusOrganizing {
		t.Fatalf("status=%s", d.Status)
	}
	if _, err := st.Dances.Create("花灯", "云南", 3, ""); err == nil {
		t.Fatal("expected duplicate name error")
	}
	got, err := st.Dances.Get(d.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "花灯" {
		t.Fatalf("name=%s", got.Name)
	}
}

func TestDanceStateMachine(t *testing.T) {
	st := openTest(t)
	d, _ := st.Dances.Create("烟盒舞", "云南", 4, "")
	// 合法流转
	for _, to := range []string{model.DanceStatusAligning, model.DanceStatusReviewing, model.DanceStatusPublished} {
		if _, err := st.Dances.SetStatus(d.ID, to); err != nil {
			t.Fatalf("transit %s: %v", to, err)
		}
	}
	// 非法回退
	if _, err := st.Dances.SetStatus(d.ID, model.DanceStatusOrganizing); err == nil {
		t.Fatal("expected invalid transition error")
	}
	// 封存
	sealed, err := st.Dances.Seal(d.ID)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	if sealed.Status != model.DanceStatusSealed {
		t.Fatalf("status=%s", sealed.Status)
	}
	if err := st.Dances.EnsureMutable(d.ID); err == nil {
		t.Fatal("expected sealed error")
	}
}

func TestMovementUnique(t *testing.T) {
	st := openTest(t)
	d, _ := st.Dances.Create("打歌", "云南", 2, "")
	m := &model.MovementUnit{DanceID: d.ID, DancerNo: 1, ActionName: "踏步", StartBeat: 1, EndBeat: 2,
		Connection: model.ConnectionContinuous, Status: model.MovementStatusCandidate}
	if _, err := st.Movements.Create(m); err != nil {
		t.Fatalf("create: %v", err)
	}
	m2 := &model.MovementUnit{DanceID: d.ID, DancerNo: 1, ActionName: "跳步", StartBeat: 1, EndBeat: 3,
		Connection: model.ConnectionContinuous, Status: model.MovementStatusCandidate}
	if _, err := st.Movements.Create(m2); err == nil {
		t.Fatal("expected duplicate movement error")
	}
}

func TestBeatAnchorUnique(t *testing.T) {
	st := openTest(t)
	d, _ := st.Dances.Create("左脚舞", "云南", 3, "")
	b := &model.BeatAnchor{DanceID: d.ID, BeatNo: 1, ImageTime: 0.5, Confidence: 0.9, Source: "cam"}
	if _, err := st.Beats.Create(b); err != nil {
		t.Fatalf("create: %v", err)
	}
	b2 := &model.BeatAnchor{DanceID: d.ID, BeatNo: 1, ImageTime: 0.6, Confidence: 0.8, Source: "cam"}
	if _, err := st.Beats.Create(b2); err == nil {
		t.Fatal("expected duplicate beat error")
	}
}

func TestVariantAdjudicate(t *testing.T) {
	st := openTest(t)
	d, _ := st.Dances.Create("霸王鞭", "云南", 4, "")
	v := &model.VariantCandidate{DanceID: d.ID, Type: model.VariantTypeMovement, RefID: 1,
		Detail: "break", Status: model.VariantStatusCandidate}
	created, err := st.Variants.Create(v)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	decided, err := st.Variants.Adjudicate(created.ID, model.VariantStatusConfirmed, "ok")
	if err != nil {
		t.Fatalf("adjudicate: %v", err)
	}
	if decided.Status != model.VariantStatusConfirmed || decided.DecidedAt == nil {
		t.Fatalf("decided=%+v", decided)
	}
	if _, err := st.Variants.Adjudicate(created.ID, model.VariantStatusRejected, "again"); err == nil {
		t.Fatal("expected already-decided error")
	}
}

func TestVersionStateMachine(t *testing.T) {
	st := openTest(t)
	d, _ := st.Dances.Create("孔雀舞", "云南", 3, "")
	v, err := st.Versions.Create(d.ID, "summary", "evidence")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if v.VersionNo != 1 {
		t.Fatalf("version_no=%d", v.VersionNo)
	}
	// draft -> shared
	shared, err := st.Versions.SetStatus(v.ID, model.VersionStatusShared)
	if err != nil {
		t.Fatalf("share: %v", err)
	}
	if shared.Status != model.VersionStatusShared {
		t.Fatalf("status=%s", shared.Status)
	}
	// shared -> frozen
	frozen, err := st.Versions.Freeze(v.ID)
	if err != nil {
		t.Fatalf("freeze: %v", err)
	}
	if frozen.Status != model.VersionStatusFrozen {
		t.Fatalf("status=%s", frozen.Status)
	}
	// frozen -> superseded
	sup, err := st.Versions.SetStatus(v.ID, model.VersionStatusSuperseded)
	if err != nil {
		t.Fatalf("supersede: %v", err)
	}
	if sup.Status != model.VersionStatusSuperseded {
		t.Fatalf("status=%s", sup.Status)
	}
	// 非法：superseded -> draft
	if _, err := st.Versions.SetStatus(v.ID, model.VersionStatusDraft); err == nil {
		t.Fatal("expected invalid transition")
	}
}

func TestPersistAndReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "persist.db")
	st1, err := Open(path)
	if err != nil {
		t.Fatalf("open1: %v", err)
	}
	d, _ := st1.Dances.Create("跳月", "云南", 5, "")
	m := &model.MovementUnit{DanceID: d.ID, DancerNo: 1, ActionName: "踩脚", StartBeat: 1, EndBeat: 2,
		Connection: model.ConnectionContinuous, Status: model.MovementStatusAligned}
	if _, err := st1.Movements.Create(m); err != nil {
		t.Fatalf("create movement: %v", err)
	}
	if err := st1.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	st2, err := Open(path)
	if err != nil {
		t.Fatalf("open2: %v", err)
	}
	defer st2.Close()
	got, err := st2.Dances.Get(d.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "跳月" {
		t.Fatalf("name=%s", got.Name)
	}
	mvs, err := st2.Movements.ListByDance(d.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(mvs) != 1 {
		t.Fatalf("movements=%d", len(mvs))
	}
}

var _ = os.Remove
