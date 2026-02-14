package templfrontend

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/a-h/templ"
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/emphasis"
	"github.com/gissleh/sarfya/sarfyaservice"
	"github.com/labstack/echo/v4"
)

//go:embed assets/*
var assets embed.FS

func Endpoints(group *echo.Group, svc *sarfyaservice.Service, emphasisStorage emphasis.Storage) {
	outputHtml := func(c echo.Context, code int, component templ.Component) error {
		c.Response().Header().Add("Content-Type", "text/html; charset=utf-8")
		c.Response().WriteHeader(code)
		return component.Render(c.Request().Context(), c.Response())
	}

	go func() {
		examples, err := svc.Storage.FetchExamples(context.Background(), nil, nil)
		if err != nil {
			log.Fatal("Failed to load examples for counting:", err)
		}

		exampleCount = len(examples)
	}()

	demo := createDemo(svc.Dictionary)

	assets, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}

	group.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			lang := c.QueryParam("lang")
			if lang == "" {
				langCookie, err := c.Cookie("lang")
				if err != nil && langCookie != nil {
					lang = langCookie.Value
				} else {
					lang = "en"
				}
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(
				context.WithValue(
					c.Request().Context(),
					langCtxKey,
					lang,
				),
				demoCtxKey,
				demo,
			)))

			return next(c)
		}
	})

	group.StaticFS("/static/", assets)

	group.GET("/", func(c echo.Context) error {
		return outputHtml(c, http.StatusOK, layoutWrapper(fmt.Sprintf("Sarfya"), indexPage()))
	})

	group.GET("/search/:search", func(c echo.Context) error {
		search, err := url.QueryUnescape(c.Param("search"))
		if err != nil {
			return outputHtml(c, http.StatusUnprocessableEntity, layoutWrapper(fmt.Sprintf("Sarfya – %s", search), searchPage(search, err.Error(), nil, nil)))
		}

		startTime := time.Now()
		res, err := svc.QueryExample(c.Request().Context(), search)
		if err != nil {
			return outputHtml(c, http.StatusInternalServerError, layoutWrapper(fmt.Sprintf("Sarfya – %s", search), searchPage(search, err.Error(), nil, nil)))
		}

		stressAnnotations := make(map[string]*emphasis.FitResult)
		if emphasisStorage != nil {
			for _, group := range res {
				for _, example := range group.Examples {
					stress, err := emphasisStorage.FindEmphasis(c.Request().Context(), example.ID)
					if err != nil {
						return outputHtml(c, http.StatusInternalServerError, layoutWrapper(fmt.Sprintf("Sarfya – %s", search), searchPage(search, err.Error(), nil, nil)))
					}

					if len(stress.Ambiguities) == 0 && len(stress.MissingParts) == 0 {
						stressAnnotations[example.ID] = stress
					}
				}
			}
		}

		duration := time.Since(startTime)
		if duration > time.Millisecond*200 {
			log.Printf("Slow! %#+v took %s", search, time.Since(startTime))
		}

		return outputHtml(c, 200, layoutWrapper(fmt.Sprintf("Sarfya – %s", search), searchPage(search, "", res, stressAnnotations)))
	})

	group.GET("/fragments/example/:id", func(c echo.Context) error {
		id, err := url.QueryUnescape(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}

		res, err := svc.FindExample(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"error": err.Error(),
			})
		}

		filter, resolved, err := sarfya.ParseFilter(c.Request().Context(), c.QueryParam("filter"), svc.Dictionary)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"error": err.Error(),
			})
		}

		match := filter.CheckExample(*res, resolved[0])
		if match == nil {
			return c.JSON(http.StatusPreconditionFailed, map[string]string{
				"error": "filter did not match on example",
			})
		}

		return outputHtml(c, 200, example(*match, c.Request().Context().Value(langCtxKey).(string), nil, -1))
	})
}
