package v1

import (
	"family-catering/config"
	handler "family-catering/internal/handler/http"
	"family-catering/internal/repository"
	"family-catering/internal/service"
	"family-catering/pkg/consts"
	"family-catering/pkg/db/postgres"
	"family-catering/pkg/db/redis"
	"family-catering/pkg/utils"
	"family-catering/pkg/web"
	"net/http"
	"time"

	_ "family-catering/docs"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

// @title           Family Catering API
// @version         1.0
// @description     Documentation for Family Catering API.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Family catering API Support
// @contact.url    http://www.family-catering.com/support
// @contact.email  support.family-catering@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:9000
// @BasePath  /api/v1

// @securityDefinitions.apiKey  BearerAuth
// @in header
// @name Authorization
func NewRouter(pg postgres.PostgresClient, redis redis.RedisClient) *chi.Mux {
	cfg := config.Cfg()
	// repositories
	ownerRepository := repository.NewOwnerRepository(pg)
	menuRepository := repository.NewMenuRepository(pg)
	authRepository := repository.NewAuthRepository(pg, redis)
	orderRepository := repository.NewOrderRepository(pg)

	// services
	// mailer
	mailerOpts := service.MailerOption{
		Host:                       cfg.Mailer.Host,
		Port:                       cfg.Mailer.Port,
		Email:                      cfg.Mailer.Email,
		Password:                   cfg.Mailer.Password,
		SupportEmail:               cfg.Mailer.SupportEmail,
		AppName:                    cfg.App.Name,
		TemplateForgotPasswordName: cfg.Mailer.TemplateForgotPassword,
		Identity:                   cfg.Mailer.Identity,
	}
	mailer := service.NewMailer(mailerOpts)

	ownerService := service.NewOwnerService(ownerRepository)
	menuService := service.NewMenuService(menuRepository)
	authService := service.NewAuthService(ownerRepository, authRepository, mailer)
	orderService := service.NewOrderService(orderRepository, menuRepository)

	// handler
	ownerHandler := handler.NewOwnerHandler(ownerService)
	menuHandler := handler.NewMenuHandler(menuService)
	authHandler := handler.NewAuthandler(authService)
	orderHandler := handler.NewOrderHandler(orderService)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Web.AllowedOrigins,
		AllowedMethods:   cfg.Web.AllowedMethods,
		AllowedHeaders:   cfg.Web.AllowedHeaders,
		AllowCredentials: true,
		MaxAge:           cfg.Web.MaxAge,
	}))
	// midsName := [6]string{"CORS", "logger", "rateLimitByIP", "authorizationRequired", "sessionRequired", "rateLimitBySID"}

	r.Use(web.Logger)
	r.Use(httprate.LimitByIP(cfg.Web.GeneralRequestLimit, time.Minute))
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	v1 := r.Route("/api/v1", func(r chi.Router) {})

	v1.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login())
		r.Put("/forgot-password", authHandler.ForgotPassword())

		r.Group(func(r chi.Router) {
			r.Use(authHandler.AuthorizationRequired)
			r.Use(authHandler.SessionRequired)
			r.Delete("/logout", authHandler.Logout())
			r.Get("/renew-access-token", authHandler.RenewAccessToken())
		})
	})

	v1.Route("/owner", func(r chi.Router) {
		r.Post("/", ownerHandler.Create())
		r.Get("/", ownerHandler.List())

		r.Route("/{id:\\d+}", func(r chi.Router) {
			r.Get("/", ownerHandler.Get())

			r.With(authHandler.AuthorizationRequired).Put("/", ownerHandler.Update())

			r.Group(func(r chi.Router) {
				r.Use(authHandler.AuthorizationRequired)
				r.Use(authHandler.SessionRequired)
				r.Use(httprate.Limit(5, 30*time.Second, httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
					return utils.ValueContext(r.Context(), consts.CtxKeySID).(string), nil
				})))
				r.Delete("/", ownerHandler.Delete())
				r.Put("/reset-password", ownerHandler.ResetPasswordById())
				r.Put("/update-email", ownerHandler.UpdateEmailByID())

			})
		})

		r.With(authHandler.AuthorizationRequired).Put("/reset-password/{rpid}", ownerHandler.ResetPasswordByEmail())
	})

	v1.Route("/menu", func(r chi.Router) {
		r.Use(authHandler.AuthorizationRequired)
		r.Get("/", menuHandler.List())
		r.Post("/", menuHandler.Create())

		r.Route("/{id:[0-9]+}", func(r chi.Router) {
			r.Get("/", menuHandler.GetByID())
			r.Put("/", menuHandler.Update())
			r.Delete("/", menuHandler.Delete())
		})
		r.Get("/name/{name}", menuHandler.GetByName())
	})

	v1.Route("/order", func(r chi.Router) {
		r.Use(authHandler.AuthorizationRequired)
		r.Post("/", orderHandler.Create())
		r.Get("/search", orderHandler.Search())
		r.Put("/confirm-payment", orderHandler.ConfirmPayment())
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		// hide the models section
		httpSwagger.UIConfig(map[string]string{"defaultModelsExpandDepth": "-1"}),
		httpSwagger.URL("doc.json")))

	// chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
	// 	fmt.Printf("[%s]: '%s' has %d middlewares: %+v\n", method, route, len(middlewares), midsName[:len(middlewares)])
	// 	return nil
	// })

	return r
}
