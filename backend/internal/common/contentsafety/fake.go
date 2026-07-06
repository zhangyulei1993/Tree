package contentsafety

import (
	"context"
	"sync"
)

type FakeClient struct {
	mu sync.Mutex

	Verdict string
	Err     error
	Calls   []FakeCall
}

type FakeCall struct {
	OpenID  string
	Scene   int
	Content string
}

func NewFakeClient(verdict string) *FakeClient {
	if verdict == "" {
		verdict = SuggestPass
	}
	return &FakeClient{Verdict: verdict}
}

func (f *FakeClient) MsgSecCheck(ctx context.Context, openid string, scene int, content string) (Result, error) {
	_ = ctx
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, FakeCall{OpenID: openid, Scene: scene, Content: content})
	if f.Err != nil {
		return Result{}, f.Err
	}
	return Result{
		Suggest: f.Verdict,
		Label:   100,
		TraceID: "fake-trace-id",
	}, nil
}

func (f *FakeClient) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.Calls)
}

func (f *FakeClient) LastCall() (FakeCall, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Calls) == 0 {
		return FakeCall{}, false
	}
	return f.Calls[len(f.Calls)-1], true
}
