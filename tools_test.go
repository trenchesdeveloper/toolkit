package toolkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("wrong length random string returned")
	}
}

var uploadTests = []struct{
	name string
	allowedTypes []string
	renameFile bool
	errorExpected bool
}{
	{name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: false, errorExpected: false},
	{name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: true, errorExpected: false},
	{name: "not allowed", allowedTypes: []string{"image/jpeg"}, renameFile: true, errorExpected: true},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests{
		// set up a pipe to avoid buffering
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			/// create the form data field 'file'
			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/img.png")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error decoding image", err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}

		}()

		// read from the pipe which receives data
		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.renameFile)
		if err != nil  && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
			}

			// clean up
			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error expected but none received", e.name)
		}

		wg.Wait()
	}
}

func TestTools_UploadOneFile(t *testing.T) {
		// set up a pipe to avoid buffering
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)


		go func() {
			defer writer.Close()

			/// create the form data field 'file'
			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Error(err)
			}

			file, err := os.Open("./testdata/img.png")
			if err != nil {
				t.Error(err)
			}
			defer file.Close()

			img, _, err := image.Decode(file)
			if err != nil {
				t.Error("error decoding image", err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}

		}()

		// read from the pipe which receives data
		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools

		uploadedFile, err := testTools.UploadOneFile(request, "./testdata/uploads/", true)
		if err != nil  {
			t.Error(err)
		}


			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFile.NewFileName)); os.IsNotExist(err) {
				t.Errorf(" expected file to exist: %s", err.Error())
			}

			// clean up
			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFile.NewFileName))
		}



func TestTools_CreateDirIfNotExist(t *testing.T) {
	var testTools Tools

	err := testTools.CreateDirIfNotExist("./testdata/myDir")

	if err != nil {
		t.Error(err)
	}

	err = testTools.CreateDirIfNotExist("./testdata/myDir")

	if err != nil {
		t.Error(err)
	}

	// clean up
	_ = os.Remove("./testdata/myDir")


}

var slugTests = []struct{
	name string
	s string
	expected string
	errorExpected bool
}{
	{name: "valid string", s: "This is a test", expected: "this-is-a-test", errorExpected: false},
	{name: "", s: "This is a test", expected: "", errorExpected: true},
	{name: "complex string", s: "Now is the time for all men! + fish & such &^123", expected: "now-is-the-time-for-all-men-fish-such-123", errorExpected: false},
	{name: "japanese string", s: "こんにちはテスト", expected: "", errorExpected: true},
	{name: "chinese string and english", s: "hello world你好测试", expected: "hello-world", errorExpected: false},
}

func TestTools_Slugify(t *testing.T) {
	for _, e := range slugTests{
		var testTools Tools

		slug, err := testTools.Slugify(e.s)
		if err != nil  && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected && slug != e.expected {
			t.Errorf("%s: expected %s, got %s", e.name, e.expected, slug)
		}
	}
}

func TestTools_DownloadStaticFile(t *testing.T) {
	// create a req
	rr := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/", nil)

	var testTools Tools

	 testTools.DownloadStaticFile(rr, req, "./testdata", "pic.jpg", "puppy.jpg")

	 res := rr.Result()

	 defer res.Body.Close()

	 if res.Header["Content-Length"][0] != "98827" {
		 t.Error("wrong content length of", res.Header["Content-Length"][0])
	 }

	 if res.Header["Content-Disposition"][0] != "attachment; filename=puppy.jpg" {
		 t.Error("wrong content disposition of", res.Header["Content-Disposition"][0])
	 }

	 _, err := ioutil.ReadAll(res.Body)

	 if err != nil {
		 t.Error(err)
	 }


}

var jsonTests = []struct{
	name string
	json string
	errorExpected bool
	maxSize int
	allowedUnknown bool
}{
	{name: "valid json", json: `{"name": "test"}`, errorExpected: false, maxSize: 1024, allowedUnknown: false},
	{name: "badly formarted json", json: `{"name": "test"`, errorExpected: true, maxSize: 1024, allowedUnknown: false},
	{name: "incorrect type", json: `{"name": 123}`, errorExpected: true, maxSize: 1024, allowedUnknown: false},
	{name: "two json objects", json: `{"name": "test"}{"name": "test"}`, errorExpected: true, maxSize: 1024, allowedUnknown: false},
	{name: "empty json", json: ``, errorExpected: true, maxSize: 1024, allowedUnknown: false},
	{name: "syntax error", json: `{"name": "test",}`, errorExpected: true, maxSize: 1024, allowedUnknown: false},
	{name: "unknown field", json: `{"name": "test", "unknown": "test"}`, errorExpected: true, maxSize: 1024, allowedUnknown: false},
	{name: "allow unknown field", json: `{"name": "test", "unknown": "test"}`, errorExpected: false, maxSize: 1024, allowedUnknown: true},
	{name: "missing field name", json: `{"name": "test", "unknown": "test"}`, errorExpected: true, maxSize: 1024, allowedUnknown: false},
	{name: "file too large", json: `{"name": "test"}`, errorExpected: true, maxSize: 1, allowedUnknown: false},
	{name: "not json", json: `not json`, errorExpected: true, maxSize: 1024, allowedUnknown: false},
}
func TestTools_ReadJSON(t *testing.T){
	var testTools Tools

	for _, e := range jsonTests{
		// set the MaxSize
		testTools.MaxJSONSize = e.maxSize

		// allow/ disallow unknown fields
		testTools.AllowUnknownFields = e.allowedUnknown

		// variable to read decoded json into
		var decodedJSON struct {
			Name string `json:"name"`
		}

		// create a request with the body
		req, err := http.NewRequest("POST", "/", bytes.NewReader([]byte(e.json)))

		if err != nil {
			t.Log(err)
		}

		// create a recorder
		rr := httptest.NewRecorder()

		// read the json
		err = testTools.ReadJSON(rr, req, &decodedJSON)

		if err == nil && e.errorExpected {
			t.Error("expected error but got none")
		}

		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		req.Body.Close()
	}
}

func TestTools_WriteJSON(t *testing.T){

	var testTools Tools

	// create a recorder
	rr := httptest.NewRecorder()

	// create a payload
	payload := JSONResponse {
		Message: "test",
		Error: false,
	}

	// create a header
	header := http.Header{}
	header.Add("FOO", "BAR")

	// write the json
	err := testTools.WriteJSON(rr, http.StatusOK, payload, header)

	if err != nil {
		t.Errorf("error writing json: %s", err)
	}


}

func TestTools_ErrorJSON(t *testing.T){

	var testTools Tools

	// create a RESPONSE recorder
	rr := httptest.NewRecorder()

	// write the json
	err := testTools.ErrorJSON(rr, errors.New("some error"), http.StatusBadRequest)

	if err != nil {
		t.Errorf("error writing json: %s", err)
	}

	var payload JSONResponse

	// read the json
	decoded := json.NewDecoder(rr.Body)

	err = decoded.Decode(&payload)

	if err != nil {
		t.Errorf("error decoding json: %s", err)
	}

	if !payload.Error {
		t.Error("error not set")
	}

	if payload.Message != "some error" {
		t.Error("wrong message")
	}

	if rr.Code != http.StatusBadRequest {
		t.Error("wrong status code")
	}

}