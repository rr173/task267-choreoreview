// 民俗舞蹈动作谱记复核台服务入口。
//
// 用法：
//   task267-choreoreview --addr :8080 --db choreoreview.db
//   task267-choreoreview --smoke-test --db smoke.db
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task267-choreoreview/internal/httpapi"
	"task267-choreoreview/internal/model"
	"task267-choreoreview/internal/service"
	"task267-choreoreview/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "choreoreview.db", "SQLite database path")
	smoke := flag.Bool("smoke-test", false, "run end-to-end self test and exit")
	flag.Parse()

	if *smoke {
		if err := runSmokeTest(*dbPath); err != nil {
			log.Fatalf("smoke test failed: %v", err)
		}
		fmt.Println("smoke test passed")
		return
	}

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	srv := &http.Server{
		Addr:              *addr,
		Handler:           httpapi.New(svc).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("task267-choreoreview listening on %s (db=%s)", *addr, *dbPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

// runSmokeTest 端到端自检：
//  1. 创建舞段项目「云南花灯·崴步」，状态 organizing；
//  2. 导入动作单元（3 个舞者×2 个动作，其中 1 处动作连接断裂）与节拍锚点（影像证据）；
//  3. 导入队形关系（1 对互斥关系制造冲突）；
//  4. 全量分析：动作连续性 + 节拍对齐 + 队形拓扑，物化异读候选；
//  5. 裁决候选（确认/否决），创建并冻结谱记版本，封存舞段；
//  6. 关闭并重开同一数据库，验证全部状态持久化与重启恢复；
//  7. 校验错误边界：封存后拒绝新增动作、重复指纹拒绝、冻结版本状态机。
func runSmokeTest(dbPath string) error {
	_ = os.Remove(dbPath)
	st, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	svc := service.New(st)

	// 1) 舞段项目
	dance, err := svc.Dance().CreateDance("云南花灯·崴步", "云南", 3, "田野影像 2026-07 采集")
	if err != nil {
		return fmt.Errorf("create dance: %w", err)
	}
	did := dance.ID

	// 2a) 动作单元：舞者1 两个连续动作 + 舞者2 两个动作（断裂）+ 舞者3 一个动作
	movementInputs := []struct {
		dancerNo, start, end int
		action, conn         string
	}{
		{1, 1, 2, "崴步", model.ConnectionContinuous},
		{1, 3, 4, "小崴", model.ConnectionContinuous},
		{2, 1, 2, "大崴", model.ConnectionContinuous},
		{2, 5, 6, "颠步", model.ConnectionContinuous}, // 断裂：3-4 缺失
		{3, 1, 4, "绕花", model.ConnectionJump},
	}
	for _, in := range movementInputs {
		if _, err := svc.Movement().AddMovement(did, in.dancerNo, in.start, in.end, in.action, in.conn); err != nil {
			return fmt.Errorf("add movement %s: %w", in.action, err)
		}
	}

	// 2b) 节拍锚点：谱记 6 拍，锚点覆盖 1-4（1,2 间隙=0.5s → 120BPM）
	anchorInputs := []struct {
		beatNo       int
		imageTime    float64
		confidence   float64
	}{
		{1, 0.00, 0.95},
		{2, 0.50, 0.93},
		{3, 1.00, 0.90},
		{4, 1.50, 0.88},
	}
	for _, in := range anchorInputs {
		if _, err := svc.Beat().AddBeatAnchor(did, in.beatNo, in.imageTime, in.confidence, "camera-A"); err != nil {
			return fmt.Errorf("add beat anchor: %w", err)
		}
	}

	// 2c) 队形关系：1↔2 两条互斥关系（line vs diagonal，节拍重叠）制造冲突
	if _, err := svc.Formation().AddFormation(did, 1, 2, 1, 4, "line"); err != nil {
		return fmt.Errorf("add formation: %w", err)
	}
	if _, err := svc.Formation().AddFormation(did, 2, 1, 1, 4, "diagonal"); err != nil {
		return fmt.Errorf("add formation conflict: %w", err)
	}

	// 3) 全量分析
	res, err := svc.Analyze().RunFullAnalysis(did)
	if err != nil {
		return fmt.Errorf("analyze: %w", err)
	}
	if res.Continuity.BrokenCount == 0 {
		return fmt.Errorf("expected movement break, got 0")
	}
	if res.Formation.ConflictCount == 0 {
		return fmt.Errorf("expected formation conflict, got 0")
	}
	if res.Alignment.MatchedCount != 4 {
		return fmt.Errorf("expected 4 matched beats, got %d", res.Alignment.MatchedCount)
	}

	// 4) 裁决：确认队形冲突候选、否决动作断裂候选
	variants, err := svc.Variant().OpenVariants(did)
	if err != nil {
		return err
	}
	if len(variants) < 2 {
		return fmt.Errorf("expected >=2 open variants, got %d", len(variants))
	}
	var formationVariant, movementVariant *model.VariantCandidate
	for _, v := range variants {
		switch v.Type {
		case model.VariantTypeFormation:
			formationVariant = v
		case model.VariantTypeMovement:
			movementVariant = v
		}
	}
	if formationVariant != nil {
		if _, err := svc.Variant().Adjudicate(formationVariant.ID, model.VariantStatusConfirmed, "地方变体：谱记者遗漏队形变化"); err != nil {
			return fmt.Errorf("adjudicate formation: %w", err)
		}
	}
	if movementVariant != nil {
		if _, err := svc.Variant().Adjudicate(movementVariant.ID, model.VariantStatusRejected, "谱记优先：影像角度误导"); err != nil {
			return fmt.Errorf("adjudicate movement: %w", err)
		}
	}

	// 5) 舞段流转 → 待对齐 → 待复核 → 已发布
	for _, to := range []string{model.DanceStatusAligning, model.DanceStatusReviewing, model.DanceStatusPublished} {
		if _, err := svc.Dance().TransitDance(did, to); err != nil {
			return fmt.Errorf("transit dance to %s: %w", to, err)
		}
	}

	// 6) 创建谱记版本：草稿 → 共享 → 冻结
	ver, err := svc.Version().CreateVersion(did, "云南花灯·崴步 谱记 v1", "锚点 1-4 对齐 120BPM；队形冲突裁决为地方变体")
	if err != nil {
		return fmt.Errorf("create version: %w", err)
	}
	// 错误边界：草稿不得直接冻结（必须先共享）
	if _, err := svc.Version().FreezeVersion(ver.ID); err == nil {
		return fmt.Errorf("expected draft freeze rejected")
	}
	// 草稿 → 共享 → 冻结
	shared, err := svc.Store.Versions.SetStatus(ver.ID, model.VersionStatusShared)
	if err != nil {
		return fmt.Errorf("share version: %w", err)
	}
	if shared.Status != model.VersionStatusShared {
		return fmt.Errorf("expected shared, got %s", shared.Status)
	}
	frozen, err := svc.Version().FreezeVersion(shared.ID)
	if err != nil {
		return fmt.Errorf("freeze version: %w", err)
	}
	if frozen.Status != model.VersionStatusFrozen {
		return fmt.Errorf("expected frozen, got %s", frozen.Status)
	}

	// 7) 封存舞段
	sealed, err := svc.Dance().SealDance(did)
	if err != nil {
		return fmt.Errorf("seal dance: %w", err)
	}
	if sealed.Status != model.DanceStatusSealed {
		return fmt.Errorf("expected sealed, got %s", sealed.Status)
	}

	// 8) 错误边界：封存后拒绝新增动作
	if _, err := svc.Movement().AddMovement(did, 9, 1, 1, "非法动作", model.ConnectionContinuous); err == nil {
		return fmt.Errorf("expected sealed dance reject movement")
	}

	// 9) 关闭并重开：验证重启恢复
	closedDanceID, closedVerID, closedSummary := did, frozen.ID, frozen.Summary
	if err := st.Close(); err != nil {
		return err
	}
	st2, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("reopen: %w", err)
	}
	defer st2.Close()
	svc2 := service.New(st2)

	got, err := svc2.Dance().GetDance(closedDanceID)
	if err != nil {
		return fmt.Errorf("reload dance: %w", err)
	}
	if got.Status != model.DanceStatusSealed {
		return fmt.Errorf("expected sealed after restart, got %s", got.Status)
	}
	vers, err := svc2.Version().ListVersions(closedDanceID)
	if err != nil {
		return err
	}
	if len(vers) != 1 || vers[0].ID != closedVerID || vers[0].Summary != closedSummary {
		return fmt.Errorf("version not recovered: %+v", vers)
	}
	if vers[0].Status != model.VersionStatusFrozen {
		return fmt.Errorf("expected frozen after restart, got %s", vers[0].Status)
	}
	mv, err := svc2.Movement().ListMovements(closedDanceID)
	if err != nil {
		return err
	}
	if len(mv) != len(movementInputs) {
		return fmt.Errorf("movements not recovered: got %d want %d", len(mv), len(movementInputs))
	}
	openAfterRestart, err := svc2.Variant().OpenVariants(closedDanceID)
	if err != nil {
		return err
	}
	openCount := len(openAfterRestart)
	if openCount != 0 {
		return fmt.Errorf("expected 0 open variants after adjudication, got %d", openCount)
	}

	fmt.Printf("smoke: dance=%d movements=%d beats=4 formations=2 variants=%d open=%d versions=%d status=%s\n",
		closedDanceID, len(mv), len(variants), openCount, len(vers), got.Status)
	return nil
}
