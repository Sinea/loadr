package backend

import (
	"errors"
	"github.com/Sinea/loadr/pkg/loadr"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockProgressHandler struct {
	loadr.ProgressHandler
	mock.Mock
}

func (m *mockProgressHandler) Delete(t loadr.Token) error {
	args := m.Called(t)
	var e error
	if args.Error(0) != nil {
		e = args.Error(0)
	}
	return e
}

func (m *mockProgressHandler) Set(t loadr.Token, p *loadr.Progress, g uint) error {
	args := m.Called(t, p, g)
	var e error
	if args.Error(0) != nil {
		e = args.Error(0)
	}
	return e
}

func TestBackend_Run(t *testing.T) {
	handler := &mockProgressHandler{}
	b := New(loadr.NetConfig{Address: "0.0.0.0:9191"}).(*backend)
	go b.Run(handler)
}

func TestBackend_UpdateProgress_BindError(t *testing.T) {
	handler := &mockProgressHandler{}
	b := &backend{handler: handler}
	progressJSON := `{"asd"":0}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(progressJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:token")
	c.SetParamNames("token")
	c.SetParamValues("abc")
	err := b.updateProgress(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestBackend_UpdateProgress_InvalidProgressJOSN(t *testing.T) {
	handler := &mockProgressHandler{}
	b := &backend{handler: handler}
	progressJSON := `{"guarantee": 9, "progress": {"stage": "asd_", "progress": 0}}`
	req := httptest.NewRequest("POST", "/abc", strings.NewReader(progressJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := b.updateProgress(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestBackend_UpdateProgress_NoError(t *testing.T) {
	handler := &mockProgressHandler{}
	handler.
		On("Set", loadr.Token("abc"), &loadr.Progress{Stage: "asd", Progress: 0}, uint(9)).
		Return(nil)
	b := &backend{handler: handler}
	progressJSON := `{"guarantee": 9, "progress": {"stage": "asd", "progress": 0}}`
	req := httptest.NewRequest("POST", "/abc", strings.NewReader(progressJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:token")
	c.SetParamNames("token")
	c.SetParamValues("abc")
	err := b.updateProgress(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestBackend_UpdateProgress_SetErrors(t *testing.T) {
	handler := &mockProgressHandler{}
	handler.
		On("Set", loadr.Token("abc"), &loadr.Progress{Stage: "asd", Progress: 0}, uint(9)).
		Return(errors.New("set error"))
	b := &backend{handler: handler}
	progressJSON := `{"guarantee": 9, "progress": {"stage": "asd", "progress": 0}}`
	req := httptest.NewRequest("POST", "/abc", strings.NewReader(progressJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:token")
	c.SetParamNames("token")
	c.SetParamValues("abc")
	err := b.updateProgress(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestBackend_Delete_Error(t *testing.T) {
	handler := &mockProgressHandler{}
	handler.
		On("Delete", loadr.Token("abc")).
		Return(errors.New("set error"))
	b := &backend{handler: handler}
	req := httptest.NewRequest("DELETE", "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:token")
	c.SetParamNames("token")
	c.SetParamValues("abc")
	err := b.deleteProgress(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestBackend_Delete_NoError(t *testing.T) {
	handler := &mockProgressHandler{}
	handler.
		On("Delete", loadr.Token("abc")).
		Once().
		Return(nil)
	b := &backend{handler: handler}
	req := httptest.NewRequest("DELETE", "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:token")
	c.SetParamNames("token")
	c.SetParamValues("abc")
	err := b.deleteProgress(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
