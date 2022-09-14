package rkt

import (
	"bytes"
	"errors"
	"testing"

	"github.com/flatcar/mayday/mayday/plugins/rkt/v1alpha"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestTarable(t *testing.T) {
	grpcpod := v1alpha.Pod{Id: "abc123"}
	p := Pod{Pod: &grpcpod}

	assert.Equal(t, p.Header().Name, "rkt/abc123")

	content := new(bytes.Buffer)
	content.ReadFrom(p.Content())

	assert.Contains(t, content.String(), "abc123")
}

func TestGetLogs(t *testing.T) {
	p1 := v1alpha.Pod{Id: "abc123", State: v1alpha.PodState_POD_STATE_RUNNING}
	p2 := v1alpha.Pod{Id: "xyz789", State: v1alpha.PodState_POD_STATE_EXITED}

	pods := []*Pod{{Pod: &p1}, {Pod: &p2}}

	viper.Set("danger", true)
	logs := getLogs(pods)
	assert.Equal(t, len(logs), 1)
	// log command is correct
	assert.EqualValues(t, logs[0].Args(), []string{"journalctl", "-M", "rkt-abc123"})
	// output will be to correct file
	assert.Equal(t, logs[0].Name(), "/rkt/abc123.log")

	viper.Set("danger", false)
	logs1 := getLogs(pods)
	assert.Equal(t, len(logs1), 0)
}

func TestGracefulFail(t *testing.T) {
	// tests that if startApi() fails, podsFromApi() and closeApi() are never called
	startCalled := false
	closedCalled := false
	podsCalled := false

	startApi = func() error {
		startCalled = true
		return errors.New("api fail")
	}

	closeApi = func() error {
		closedCalled = true
		return nil
	}

	podsFromApi = func() ([]*v1alpha.Pod, error) {
		podsCalled = true
		return nil, nil
	}

	GetPods()

	assert.True(t, startCalled)
	assert.False(t, closedCalled)
	assert.False(t, podsCalled)
}

func TestSuccess(t *testing.T) {
	// tests that if startApi() succeeds, other functions are properly called
	startCalled := false
	closedCalled := false
	podsCalled := false

	startApi = func() error {
		startCalled = true
		return nil
	}

	closeApi = func() error {
		closedCalled = true
		return nil
	}

	podsFromApi = func() ([]*v1alpha.Pod, error) {
		podsCalled = true
		return nil, nil
	}

	_, _, err := GetPods()
	assert.Nil(t, err)

	assert.True(t, startCalled)
	assert.True(t, closedCalled)
	assert.True(t, podsCalled)
}
