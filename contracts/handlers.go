package contracts

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"techunicorn.com/udc-core/gettingStarted/pkg/app/server/contracts"
	domctr "techunicorn.com/udc-core/gettingStarted/pkg/domain/contracts"
)

type handlerContext interface {
	context.Context
	GetIdentity() (typ string, id string, roles []string, features []string)
}

type Handler struct {
	client *contracts.HealthzClient
	conn   *grpc.ClientConn
}

func NewHandler(
	prefix string,
	address string,
	opts []grpc.DialOption,
) ([]HandlerPackage, error) {

	conn, err := grpc.Dial(address, opts...)
	if err != nil {

	}
	client := contracts.NewHealthzClient(conn)

	packages := []HandlerPackage{
		{
			Method: http.MethodGet,
			Path:   "/health",
			Handler: func(c *fiber.Ctx) error {
				body := domctr.HealthQuery{}

				raw := c.Body()
				if err := protojson.Unmarshal(raw, &body); err != nil {
					return err
				}
				craw := c.UserContext()
				if craw == nil {

				}
				ctx, ok := craw.(handlerContext)
				if !ok {
					return fmt.Errorf("invalid context provided")
				}

				typ, id, roles, feats := ctx.GetIdentity()
				body.UserContext = &domctr.UserContext{
					UserType: typ,
					Id:       id,
					Roles:    roles,
					Features: feats,
				}

				res, err := client.GetHealthStatus(ctx, &body, nil)
				if err != nil {
					return err
				}

				json, err := protojson.Marshal(res)
				if err != nil {
					return err
				}
				c.Set("Content-type", "application/json; charset=utf-8")
				c.Status(fiber.StatusOK)
				c.Write(json)
				return nil
			},
		},
	}

	return packages, nil
}

type HandlerPackage struct {
	Method  string
	Handler fiber.Handler
	Path    string
}
