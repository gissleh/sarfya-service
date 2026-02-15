package webapi

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/emphasis"
	"github.com/gissleh/sarfya-service/textformats"
	"github.com/gissleh/sarfya/sarfyaservice"
	"github.com/labstack/echo/v4"
)

func Examples(group *echo.Group, svc *sarfyaservice.Service, emphasisStorage emphasis.Storage) {
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

		stressAnnotations := make(map[string]emphasis.FitResult)
		if emphasisStorage != nil {
			for _, group := range res {
				for _, example := range group.Examples {
					fitResult, err := emphasisStorage.FindEmphasis(c.Request().Context(), example.ID)
					if err != nil {
						return err
					}

					stressAnnotations[example.ID] = *fitResult
				}
			}
		}

		duration := time.Since(startTime)
		if duration > time.Millisecond*200 {
			log.Printf("Slow! %#+v took %s", search, time.Since(startTime))
		}

		if compactLang != "" {
			compacts := make([]sarfyaservice.FilterMatchGroupCompact, 0, len(res))
			for _, group := range res {
				compacts = append(compacts, *group.ToCompact(compactLang))
			}

			return c.JSON(http.StatusOK, map[string]any{
				"groups":            compacts,
				"executionMs":       duration.Seconds() * 1000.0,
				"lang":              compactLang,
				"stressAnnotations": stressAnnotations,
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"examples":          res,
			"executionMs":       duration.Seconds() * 1000.0,
			"stressAnnotations": stressAnnotations,
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

	group.GET("/:id/discord-quote", func(c echo.Context) error {
		example, err := svc.FindExample(c.Request().Context(), c.Param("id"))
		if err != nil {
			return err
		}

		var match *sarfya.FilterMatch

		if c.QueryParam("filter") != "" {
			filter, resolvedMaps, err := sarfya.ParseFilter(c.Request().Context(), c.QueryParam("filter"), svc.Dictionary)
			if err != nil {
				return err
			}

			filterIndex, err := strconv.Atoi(c.QueryParam("filter_index"))
			if err != nil {
				if c.Param("filter_index") == "" {
					filterIndex = 0
				} else {
					return err
				}
			}

			match = filter.CheckExample(*example, resolvedMaps[filterIndex])
			if match == nil {
				return errors.New("filter is not supposed to fail, parameters must be incorrect")
			}
		} else {
			match = &sarfya.FilterMatch{
				Example:             *example,
				Spans:               nil,
				TranslationAdjacent: nil,
				TranslationSpans:    nil,
				WordMap:             nil,
			}
		}

		var stress *emphasis.FitResult
		if emphasisStorage != nil {
			stress, err = emphasisStorage.FindEmphasis(c.Request().Context(), example.ID)
			if err != nil {
				return err
			}

			if len(stress.Ambiguities) > 0 || len(stress.MissingParts) > 0 {
				stress = nil
			}
		}

		var formatter textformats.Formatter
		switch c.QueryParam("format") {
		case "discord", "":
			formatter = textformats.DiscordFormatter()
		case "bbcode":
			formatter = textformats.BBCodeFormatter()
		default:
			return fmt.Errorf("unknown format: %s", c.QueryParam("format"))
		}

		return c.JSON(http.StatusOK, map[string]any{
			"text": textformats.Generate(formatter, *match, c.Param("lang"), stress),
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
