package http

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTimeOptionAdministrationRequiresAdmin(t *testing.T) {
	for _, role := range []string{"", "user"} {
		e := echo.New()
		req := httptest.NewRequest("PUT", "/admin/time-options/10-seconds", strings.NewReader(`{"label":"X","seconds":10,"enabled":true}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Role", role)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		require.NoError(t, (Handler{}).PutTimeOption(c))
		require.Equal(t, 403, rec.Code)
	}
}
