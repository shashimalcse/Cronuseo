package check

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shashimalcse/cronuseo/internal/util"
)

func RegisterHandlers(r *echo.Group, service Service) {
	res := permission{service: service}
	router := r.Group("/:org/permission/check")
	router.POST("", res.check)
}

type permission struct {
	service Service
}

// @Description Check.
// @Tags        Permission
// @Accept      json
// @Param org path string true "Organization"
// @Param request body CheckRequest true "body"
// @Produce     json
// @Success     201
// @failure     400,403,500
// @Router      /{org}/permission/check [post]
func (r permission) check(c echo.Context) error {
	var input CheckRequest
	api_key := c.Request().Header.Get("API_KEY")
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid inputs. Please check your inputs")
	}

	allow, err := r.service.Check(context.Background(), c.Param("org"), input, api_key)
	if err != nil {
		return util.HandleError(err)
	}

	return c.JSON(http.StatusOK, allow)
}
