package loadr

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log"
	"testing"
)

type mockClient struct {
	Client
	mock.Mock
	isAliveReturned chan struct{}
}

func (m *mockClient) Write(p *Progress) error {
	var e error = nil
	args := m.Called()
	if args.Error(0) != nil {
		e = args.Error(0)
	}
	return e
}

func (m *mockClient) Close() error {
	var e error = nil
	args := m.Called()
	if args.Error(0) != nil {
		e = args.Error(0)
	}
	return e
}

func (m *mockClient) IsAlive() bool {
	if m.isAliveReturned != nil {
		defer close(m.isAliveReturned)
	}
	return m.Called().Bool(0)
}

type mockStore struct {
	Store
	mock.Mock
}

func (m *mockStore) Delete(t Token) error {
	return m.Called().Error(0)
}

func (m *mockStore) Set(Token, *Progress) error {
	return m.Called().Error(0)
}

func (m *mockStore) Get(t Token) (*Progress, error) {
	args := m.Called()
	var p *Progress = nil
	var e error = nil
	if args.Get(0) != nil {
		p = args.Get(0).(*Progress)
	}
	if args.Get(1) != nil {
		e = args.Error(1)
	}
	return p, e
}

type mockChannel struct {
	Channel
	mock.Mock
}

func (m *mockChannel) Progresses() <-chan MetaProgress {
	args := m.Called()
	return args.Get(0).(chan MetaProgress)
}

func (m *mockChannel) Push(MetaProgress) error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockChannel) Errors() <-chan error {
	args := m.Called()

	return args.Get(0).(chan error)
}

type backendListenerMock struct {
	BackendListener
	mock.Mock
}

func (m *backendListenerMock) Run(handler ProgressHandler) {

}

type mockClientsListener struct {
	ClientListener
	mock.Mock
}

func (m *mockClientsListener) Wait() <-chan *Subscription {
	return m.Called().Get(0).(chan *Subscription)
}

func TestService_HandleSubscription_WithStoreError(t *testing.T) {
	store := &mockStore{}
	store.On("Get").Once().Return(nil, errors.New("asd"))

	progressChan := make(chan MetaProgress)
	channel := &mockChannel{}
	channel.On("Progresses").Return(progressChan)
	channel.On("Errors").Return(make(chan error))

	client := &mockClient{}
	client.On("Write").Once().Return(nil)
	bb := new(bytes.Buffer)
	testLogger := log.New(bb, "", 0)
	s := New(store, channel, testLogger)

	s.Subscribe(Token("x"), client)
}

func TestService_HandleSubscription_WithoutStoreError(t *testing.T) {
	progress := &Progress{Stage: "", Progress: 0}
	store := &mockStore{}
	store.On("Get").Once().Return(progress, nil)

	progressChan := make(chan MetaProgress, 1)
	channel := &mockChannel{}
	channel.On("Progresses").Return(progressChan)
	channel.On("Errors").Return(make(chan error))

	client := &mockClient{}
	client.On("Write").Once().Return(nil)

	bb := new(bytes.Buffer)
	testLogger := log.New(bb, "", 0)
	s := New(store, channel, testLogger)
	s.Subscribe(Token("x"), client)
}

func TestService_HandleSubscription_WithoutStoreErrorAndClientError(t *testing.T) {
	progress := &Progress{Stage: "", Progress: 0}
	store := &mockStore{}
	store.On("Get").Once().Return(progress, nil)

	progressChan := make(chan MetaProgress, 1)
	channel := &mockChannel{}
	channel.On("Progresses").Return(progressChan)
	channel.On("Errors").Return(make(chan error))

	client := &mockClient{}
	client.On("Write").Return(errors.New(""))
	client.On("Close").Return(nil)

	bb := new(bytes.Buffer)
	testLogger := log.New(bb, "", 0)
	s := New(store, channel, testLogger)
	s.Subscribe(Token("x"), client)
}

func TestService_HandleProgress_ClientWriteError(t *testing.T) {
	progress := &Progress{Stage: "", Progress: 0}
	store := &mockStore{}
	store.On("Get").Once().Return(progress, nil)

	progressChan := make(chan MetaProgress, 1)
	channel := &mockChannel{}
	channel.On("Progresses").Return(progressChan)
	channel.On("Errors").Return(make(chan error))

	client := &mockClient{}
	client.On("Write").Return(errors.New(""))
	client.On("Close").Return(nil)

	bb := new(bytes.Buffer)
	testLogger := log.New(bb, "", 0)
	s := New(store, channel, testLogger)
	s.Subscribe(Token("x"), client)
	s.HandleProgress(MetaProgress{Token: Token("x"), Progress: Progress{Stage: "x", Progress: 0}})
}

func TestService_Delete(t *testing.T) {
	progress := &Progress{Stage: "", Progress: 0}
	store := &mockStore{}
	store.On("Get").Once().Return(progress, nil)
	store.On("Delete").Once().Return(nil)

	progressChan := make(chan MetaProgress, 1)
	channel := &mockChannel{}
	channel.On("Progresses").Return(progressChan)
	channel.On("Errors").Return(make(chan error))

	client := &mockClient{}
	client.On("Write").Return(nil)
	client.On("Close").Return(nil)

	bb := new(bytes.Buffer)
	testLogger := log.New(bb, "", 0)
	s := New(store, channel, testLogger)
	err := s.Delete(Token("x"))
	assert.NoError(t, err)
}

func TestService_Delete_Fails(t *testing.T) {
	progress := &Progress{Stage: "", Progress: 0}
	store := &mockStore{}
	store.On("Get").Once().Return(progress, nil)
	store.On("Delete").Once().Return(errors.New("some error"))

	progressChan := make(chan MetaProgress, 1)
	channel := &mockChannel{}
	channel.On("Progresses").Return(progressChan)
	channel.On("Errors").Return(make(chan error))

	client := &mockClient{}
	client.On("Close").Once().Return(errors.New("close error"))
	client.On("Write").Return(nil)

	bb := new(bytes.Buffer)
	testLogger := log.New(bb, "", 0)
	s := New(store, channel, testLogger)
	s.Subscribe(Token("x"), client)
	err := s.Delete(Token("x"))

	assert.Error(t, err)
}

func TestService_Set_InvalidProgress(t *testing.T) {
	progress := &Progress{Stage: "", Progress: 0}
	store := &mockStore{}
	store.On("Get").Once().Return(progress, nil)
	store.On("Delete").Once().Return(errors.New("some error"))

	progressChan := make(chan MetaProgress, 1)
	channel := &mockChannel{}
	channel.On("Progresses").Return(progressChan)
	channel.On("Errors").Return(make(chan error))

	client := &mockClient{}
	client.On("Close").Once().Return(errors.New("close error"))
	client.On("Write").Return(nil)

	bb := new(bytes.Buffer)
	testLogger := log.New(bb, "", 0)
	s := New(store, channel, testLogger)
	err := s.Set(Token("x"), &Progress{Stage: "loading_", Progress: 0}, 0)

	assert.Error(t, err)
}

func TestService_Set_StoreError(t *testing.T) {
	store := &mockStore{}
	store.On("Set").Once().Return(errors.New("storage error"))

	channel := &mockChannel{}
	channel.On("Push").Once().Return(nil)

	client := &mockClient{}
	client.On("Close").Once().Return(errors.New("close error"))
	client.On("Write").Return(nil)

	bb := new(bytes.Buffer)
	testLogger := log.New(bb, "", 0)
	s := New(store, channel, testLogger)
	err := s.Set(Token("x"), &Progress{Stage: "loading", Progress: 0}, Storage)

	assert.Error(t, err)
}

func TestService_Set_ChannelError(t *testing.T) {
	store := &mockStore{}
	store.On("Set").Once().Return(nil)

	channel := &mockChannel{}
	channel.On("Push").Once().Return(errors.New("channel error"))

	client := &mockClient{}
	client.On("Close").Once().Return(errors.New("close error"))
	client.On("Write").Return(nil)

	bb := new(bytes.Buffer)
	testLogger := log.New(bb, "", 0)
	s := New(store, channel, testLogger)
	err := s.Set(Token("x"), &Progress{Stage: "loading", Progress: 0}, Broadcast)

	assert.Error(t, err)
}

func TestService_Set_NoError(t *testing.T) {
	store := &mockStore{}
	store.On("Set").Once().Return(nil)

	channel := &mockChannel{}
	channel.On("Push").Once().Return(nil)

	client := &mockClient{}
	client.On("Close").Once().Return(errors.New("close error"))
	client.On("Write").Return(nil)

	bb := new(bytes.Buffer)
	testLogger := log.New(bb, "", 0)
	s := New(store, channel, testLogger)
	err := s.Set(Token("x"), &Progress{Stage: "loading", Progress: 0}, 0)

	assert.NoError(t, err)
}
