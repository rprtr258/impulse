package service

import (
	"context"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pkg/errors"
	"github.com/rprtr258/fun"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"

	"github.com/impulse-http/local-backend/internal/database"
)

type Service struct {
	DB *database.DB
}

func handlerWrapper[In, Out any](handler func(context.Context, In) (Out, error)) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request In
		if err := c.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, errors.Wrap(err, "read request body").Error())
		}

		response, err := handler(c.Context(), request)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(response)
	}
}

func handler(handlers map[string]func(*fiber.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var route struct {
			Path string `json:"ROUTE"`
		}
		if err := c.BodyParser(&route); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, errors.Wrap(err, "read ROUTE").Error())
		} else if !fun.Has(handlers, route.Path) {
			return fiber.NewError(fiber.StatusNotFound)
		}

		return handlers[route.Path](c)
	}
}

func New(dbFs afero.Fs) (*fiber.App, func()) {
	db := database.New(dbFs)

	s := &Service{DB: db}

	app := fiber.New(fiber.Config{
		Immutable: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			type ErrorMessage struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}

			// Retrieve custom status code if it is *fiber.Error
			var e *fiber.Error
			if errors.As(err, &e) {
				return c.
					Status(e.Code).
					JSON(ErrorMessage{
						Message: "error occured",
						Error:   e.Message,
					})
			}

			return c.
				Status(fiber.StatusInternalServerError).
				JSON(ErrorMessage{
					Message: "internal error",
					Error:   err.Error(),
				})
		},
	})
	app.Use(
		cors.New(cors.Config{
			AllowOrigins: "*",
			AllowHeaders: "Origin, Content-Type, Accept",
		}),
		fiberzerolog.New(fiberzerolog.Config{
			Logger: &log.Logger,
			Fields: []string{fiberzerolog.FieldIP, fiberzerolog.FieldLatency, fiberzerolog.FieldStatus, fiberzerolog.FieldMethod, fiberzerolog.FieldURL, fiberzerolog.FieldError, fiberzerolog.FieldBody},
			SkipBody: func(*fiber.Ctx) bool {
				return false
			},
		}),
	)
	app.Static("/", "./frontend/dist")

	app.Post("/api", handler(map[string]func(*fiber.Ctx) error{
		"/list": handlerWrapper(s.HandlerList),
		// TODO: support operating on dirs also
		"/create":  handlerWrapper(s.HandlerNew),
		"/read":    handlerWrapper(s.HandlerGet),
		"/update":  handlerWrapper(s.HandlerUpdate),
		"/delete":  handlerWrapper(s.HandlerDelete),
		"/perform": handlerWrapper(s.HandlerSend),
	}))

	return app, func() {
		db.Close()
	}
}
