package funcs_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/howlrs/buy-the-dips/funcs"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	godotenv.Load("../.env.local")
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/buy-the-dips", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := funcs.BuyTheDips(c)

	// Assertions
	var response map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&response)

	if assert.NoError(t, h) {
		if rec.Code == http.StatusNoContent {
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Equal(t, "not dip", response["message"])
		} else if rec.Code == http.StatusOK {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "buy the dips", response["message"])
		}
	} else {
		t.Logf("Error: %v", response["message"])
		assert.NotEqual(t, http.StatusOK, rec.Code)
		assert.NotEqual(t, "buy the dips", response["message"])
	}
}
