package webapi

import (
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya/sarfyaservice"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"net/url"
	"time"
)

func Examples(group *echo.Group, svc *sarfyaservice.Service) {
	group.GET("/:search", func(c echo.Context) error {
		search, err := url.QueryUnescape(c.Param("search"))
		if err != nil {
			return err
		}

		compactLang := c.QueryParam("compact")

		startTime := time.Now()
		res, err := svc.QueryExample(c.Request().Context(), search)
		if err != nil {
			return err
		}

		duration := time.Since(startTime)
		if duration > time.Millisecond*100 {
			log.Printf("Slow! %#+v exectured in %s", search, time.Since(startTime))
		}

		if compactLang != "" {
			compacts := make([]sarfyaservice.FilterMatchGroupCompact, 0, len(res))
			for _, group := range res {
				compacts = append(compacts, *group.ToCompact(compactLang))
			}

			return c.JSON(http.StatusOK, map[string]any{
				"groups":      compacts,
				"executionMs": duration.Seconds() * 1000.0,
				"lang":        compactLang,
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"examples":    res,
			"executionMs": duration.Seconds() * 1000.0,
		})
	})

	group.GET("/:id/input", func(c echo.Context) error {
		example, err := svc.FindExample(c.Request().Context(), c.Param("id"))
		if err != nil {
			return err
		}

		input, err := example.MinimalInput(c.Request().Context(), svc.Dictionary)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]any{
			"input": input,
		})
	})

	group.POST("/", func(c echo.Context) error {
		input := sarfya.Input{}
		if err := c.Bind(&input); err != nil {
			return err
		}
		if input.Text == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Text cannot be left blank",
			})
		}

		if input.Source.ID == "" || input.Source.Date == "" || input.Source.URL == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Missing fields in source",
			})
		}

		res, err := svc.SaveExample(c.Request().Context(), input, c.QueryParam("dry") == "true")
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]any{
			"example": *res,
		})
	})

	group.DELETE("/:id", func(c echo.Context) error {
		example, err := svc.DeleteExample(c.Request().Context(), c.Param("id"))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]any{
			"example": example,
		})
	})

}
