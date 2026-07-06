package adapter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ScenarioHeader はリクエスト単位のシナリオ切り替えに使うヘッダ名。
const ScenarioHeader = "X-Mockport-Scenario"

// ErrUnknownScenario は未知のシナリオ名が指定されたときに返るエラー。
var ErrUnknownScenario = errors.New("unknown scenario")

// scenarioContextKey は解決済みシナリオを request context に載せるためのキー。
type scenarioContextKey struct{}

// scenarioHolder は解決済みシナリオをミドルウェアへ受け渡すための可変ホルダ。
//
// net/http はミドルウェアとハンドラで同一の *http.Request を共有するため、
// ミドルウェアが context にホルダを差し込み、アダプタが Resolve 成功時に
// ホルダへ書き込むことで「実際に採用された（検証済みの）シナリオ」だけを
// ミドルウェアが読み取れる。未知シナリオで Resolve が失敗した場合は何も
// 書き込まれないため、不正なヘッダ値がレポートに混入しない。
type scenarioHolder struct {
	scenario string
	set      bool
}

// WithScenarioCapture はレポート記録用に解決済みシナリオを受け取るホルダを
// request context へ差し込んだ新しい Request を返す。ミドルウェアがハンドラ
// 呼び出し前に一度だけ呼び出すことを想定している。
func WithScenarioCapture(req *http.Request) *http.Request {
	holder := &scenarioHolder{}
	ctx := context.WithValue(req.Context(), scenarioContextKey{}, holder)
	return req.WithContext(ctx)
}

// ResolvedScenarioFromContext は WithScenarioCapture で差し込まれたホルダから
// 解決済みシナリオを取り出す。アダプタが値をセットしていれば (scenario, true)、
// そうでなければ ("", false) を返す。
func ResolvedScenarioFromContext(ctx context.Context) (string, bool) {
	holder, ok := ctx.Value(scenarioContextKey{}).(*scenarioHolder)
	if !ok || holder == nil || !holder.set {
		return "", false
	}
	return holder.scenario, true
}

// storeResolvedScenario は解決済みシナリオを context 上のホルダへ記録する。
// ホルダが無い場合（テスト等で WithScenarioCapture を経由していない場合）は
// 何もしない。
func storeResolvedScenario(ctx context.Context, scenario string) {
	if holder, ok := ctx.Value(scenarioContextKey{}).(*scenarioHolder); ok && holder != nil {
		holder.scenario = scenario
		holder.set = true
	}
}

// ScenarioResolver はリクエスト単位のシナリオ解決ロジックを担う。
// 解決順序: ヘッダ > Config.Scenario > defaultName。
type ScenarioResolver struct {
	configured  string
	defaultName string
	known       map[string]bool
}

// NewScenarioResolver は ScenarioResolver を生成する。
func NewScenarioResolver(cfg Config, defaultName string, meta Metadata) *ScenarioResolver {
	known := make(map[string]bool, len(meta.Scenarios))
	for _, s := range meta.Scenarios {
		known[s.Name] = true
	}
	return &ScenarioResolver{
		configured:  cfg.Scenario,
		defaultName: defaultName,
		known:       known,
	}
}

// Resolve はリクエストのヘッダを参照してシナリオ名を解決する。
// ヘッダに未知の値が含まれる場合は ErrUnknownScenario を返す。
//
// 解決に成功した場合のみ、採用されたシナリオを request context のホルダへ
// 記録する（WithScenarioCapture で差し込まれている場合）。これにより
// recordMiddleware は「実際に採用された検証済みのシナリオ」だけをレポートへ
// 記録でき、未知シナリオで弾かれた不正なヘッダ値が混入しない。
func (s *ScenarioResolver) Resolve(req *http.Request) (string, error) {
	scenario, err := s.resolve(req)
	if err != nil {
		return "", err
	}
	storeResolvedScenario(req.Context(), scenario)
	return scenario, nil
}

func (s *ScenarioResolver) resolve(req *http.Request) (string, error) {
	if value := strings.TrimSpace(req.Header.Get(ScenarioHeader)); value != "" {
		if !s.known[value] {
			return "", fmt.Errorf("%w: %s", ErrUnknownScenario, value)
		}
		return value, nil
	}
	if s.configured != "" {
		if !s.known[s.configured] {
			return "", fmt.Errorf("%w: %s", ErrUnknownScenario, s.configured)
		}
		return s.configured, nil
	}
	return s.defaultName, nil
}
