package httpreq_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/visonhuo/mykit/pkg/httpreq"
)

type request struct {
	method  string
	body    []byte
	headers map[string][]string
	forms   map[string][]string
	query   map[string][]string
}

func TestRequest(t *testing.T) {
	type setupFunc func(t *testing.T) (*httptest.Server, *request)
	type testFunc func(t *testing.T, ts *httptest.Server, recv *request)
	type testSet struct {
		setup setupFunc
		test  testFunc
	}

	for name, set := range map[string]testSet{
		"verifyParamsRaw":      {setupServer, testVerifyParamsRaw},
		"verifyParamsForm":     {setupServer, testVerifyParamsForm},
		"verifyParamJSON":      {setupServer, testVerifyParamsJSON},
		"verifyParamXML":       {setupServer, testVerifyParamsXML},
		"customMarshaller":     {setupServer, testCustomMarshaller},
		"unexpectedStatusCode": {setupServer500, testUnexpectedStatusCode},
		"tlsClient":            {setupServerTLS, testTLSClient},
	} {
		t.Run(name, func(t *testing.T) {
			ts, recv := set.setup(t)
			defer ts.Close()
			set.test(t, ts, recv)
		})
	}
}

func setupServer(t *testing.T) (*httptest.Server, *request) {
	var recv request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		require.NoError(t, req.ParseForm())
		recv.method = req.Method
		recv.body, _ = io.ReadAll(req.Body)
		recv.headers = req.Header
		recv.forms = req.Form
		recv.query = req.URL.Query()
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		require.NoError(t, err)
		return
	}))
	return ts, &recv
}

func setupServer500(t *testing.T) (*httptest.Server, *request) {
	var recv request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		recv.method = req.Method
		recv.body, _ = io.ReadAll(req.Body)
		recv.headers = req.Header
		recv.forms = req.Form
		recv.query = req.URL.Query()
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Server error"))
		require.NoError(t, err)
		return
	}))
	return ts, &recv
}

func setupServerTLS(t *testing.T) (*httptest.Server, *request) {
	var recv request
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		require.NoError(t, req.ParseForm())
		recv.method = req.Method
		recv.body, _ = io.ReadAll(req.Body)
		recv.headers = req.Header
		recv.forms = req.Form
		recv.query = req.URL.Query()
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		require.NoError(t, err)
		return
	}))
	return ts, &recv
}

func testVerifyParamsRaw(t *testing.T, ts *httptest.Server, recv *request) {
	respBody, err := httpreq.NewRequest(context.Background(), http.MethodPost, ts.URL, []byte("body")).
		Header("Header", "h1", "h2").
		Query("query", "q1", "q2").
		Call().Expect200().Response()
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), respBody)
	require.Equal(t, http.MethodPost, recv.method)
	require.Equal(t, []byte("body"), recv.body)
	require.Equal(t, []string{"h1", "h2"}, recv.headers["Header"])
	require.Equal(t, []string{"q1", "q2"}, recv.query["query"])
}

func testVerifyParamsForm(t *testing.T, ts *httptest.Server, recv *request) {
	respBody, err := httpreq.NewRequest(context.Background(), http.MethodPost, ts.URL, url.Values{"form": {"f1", "f2"}}).
		Header("Header", "h1", "h2").
		Query("query", "q1", "q2").
		Call().Expect200().Response()
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), respBody)
	require.Equal(t, http.MethodPost, recv.method)
	require.Equal(t, []string{"f1", "f2"}, recv.forms["form"])
	require.Equal(t, []string{"h1", "h2"}, recv.headers["Header"])
	require.Equal(t, []string{"q1", "q2"}, recv.query["query"])
}

func testVerifyParamsJSON(t *testing.T, ts *httptest.Server, recv *request) {
	respBody, err := httpreq.NewRequest(context.Background(), http.MethodPost, ts.URL, map[string]string{"status": "ok"}).
		Header("Header", "h1", "h2").
		Query("query", "q1", "q2").
		Call().Expect200().Response()
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), respBody)
	require.Equal(t, http.MethodPost, recv.method)
	require.Equal(t, []byte(`{"status":"ok"}`), recv.body)
	require.Equal(t, []string{"h1", "h2"}, recv.headers["Header"])
	require.Equal(t, []string{"q1", "q2"}, recv.query["query"])

	respBody, err = httpreq.NewRequest(context.Background(), http.MethodPost, ts.URL, `{"status":"ok2"}`).
		ContentType(httpreq.ContentTypeJSON).
		Call().Expect200().Response()
	require.NoError(t, err)
	require.Equal(t, []byte(`{"status":"ok2"}`), recv.body)
}

func testVerifyParamsXML(t *testing.T, ts *httptest.Server, recv *request) {
	type obj struct {
		Status bool `xml:"status"`
	}
	respBody, err := httpreq.NewRequest(context.Background(), http.MethodPost, ts.URL, obj{Status: true}).
		ContentType(httpreq.ContentTypeXML).
		Header("Header", "h1", "h2").
		Query("query", "q1", "q2").
		Call().Expect200().Response()
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), respBody)
	require.Equal(t, http.MethodPost, recv.method)
	require.Equal(t, []byte(`<obj><status>true</status></obj>`), recv.body)
	require.Equal(t, []string{"h1", "h2"}, recv.headers["Header"])
	require.Equal(t, []string{"q1", "q2"}, recv.query["query"])

	respBody, err = httpreq.NewRequest(context.Background(), http.MethodPost, ts.URL, `<obj><status>false</status></obj>`).
		ContentType(httpreq.ContentTypeXML).
		Call().Expect200().Response()
	require.NoError(t, err)
	require.Equal(t, []byte(`<obj><status>false</status></obj>`), recv.body)
}

type mockMarshaller struct {
}

func (m mockMarshaller) Marshal(_ interface{}) ([]byte, error) {
	return []byte("mock"), nil
}

func testCustomMarshaller(t *testing.T, ts *httptest.Server, recv *request) {
	mockContentType := "mock"
	httpreq.RegisterMarshaller(mockContentType, mockMarshaller{})

	respBody, err := httpreq.NewRequest(context.Background(), http.MethodPost, ts.URL, []byte("ok")).
		ContentType(mockContentType).
		Call().Expect200().Response()
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), respBody)
	require.Equal(t, http.MethodPost, recv.method)
	require.Equal(t, []byte(`mock`), recv.body)

	_, err = httpreq.NewRequest(context.Background(), http.MethodPost, ts.URL, []byte("ok")).
		ContentType("wrong-content-type").
		Call().Expect200().Response()
	require.Error(t, err)
}

func testUnexpectedStatusCode(t *testing.T, ts *httptest.Server, recv *request) {
	respBody, err := httpreq.NewRequest(context.Background(), http.MethodGet, ts.URL, nil).
		Call().ExpectStatusCodes(http.StatusOK).Response()
	require.IsType(t, httpreq.UnexpectedStatusCodeError{}, err)
	require.Equal(t, http.StatusInternalServerError, err.(httpreq.UnexpectedStatusCodeError).StatusCode)
	require.Equal(t, []byte("Server error"), respBody)
	require.Equal(t, http.MethodGet, recv.method)
}

func testTLSClient(t *testing.T, ts *httptest.Server, recv *request) {
	respBody, err := httpreq.NewRequest(context.Background(), http.MethodPost, ts.URL, []byte("body")).
		Call().Expect200().Response()
	require.Error(t, err)

	respBody, err = httpreq.NewClient(ts.Client()).NewRequest(context.Background(), http.MethodPost, ts.URL, []byte("body")).
		Header("Header", "h1", "h2").
		Query("query", "q1", "q2").
		Call().Expect200().Response()
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), respBody)
	require.Equal(t, http.MethodPost, recv.method)
	require.Equal(t, []byte("body"), recv.body)
	require.Equal(t, []string{"h1", "h2"}, recv.headers["Header"])
	require.Equal(t, []string{"q1", "q2"}, recv.query["query"])
}
