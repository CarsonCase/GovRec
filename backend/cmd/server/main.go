package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/template"

	"github.com/CarsonCase/GovRec/pkg/sam"
	"github.com/joho/godotenv"
)

type AppWrapper struct {
	app *pocketbase.PocketBase
}

func authErrorHandler(title string, err error, c echo.Context) error {
	myErr := apis.NewBadRequestError(title, err)
	log.Println(myErr)
	c.Redirect(http.StatusTemporaryRedirect, "/")
	return myErr
}

func (wapp *AppWrapper) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("AuthToken")
		if err != nil {
			return authErrorHandler("Cookie Error ", err, c)
		}

		_, err = wapp.app.Dao().FindAuthRecordByToken(cookie.Value, wapp.app.Settings().RecordAuthToken.Secret)

		if err != nil {
			return authErrorHandler("Cookie Authorization Error ", err, c)
		}

		return next(c)
	}

}

func main() {
	app := pocketbase.New()
	rootDir, _ := os.Getwd()

	err := godotenv.Load(rootDir + "/.env")

	if err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	sam := sam.Sam{API_KEY: os.Getenv("SAM_API_KEY")}
	wrappedApp := AppWrapper{app: app}

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		registry := template.NewRegistry()

		e.Router.GET("/", func(c echo.Context) error {

			html, err := registry.LoadFiles(
				"../web/index.html",
			).Render(map[string]any{})

			if err != nil {
				// or redirect to a dedicated 404 HTML page
				return apis.NewNotFoundError("", err)
			}

			return c.HTML(http.StatusOK, html)
		})

		e.Router.GET("home", func(c echo.Context) error {
			listings, err := sam.GetNListings(3)
			fmt.Println(listings)

			if err != nil {
				log.Fatal("Error getting listings: ", err)
			}

			html, err := registry.LoadFiles(
				"../web/home.html",
			).Render(map[string]any{
				"listings": listings,
			})

			if err != nil {
				// or redirect to a dedicated 404 HTML page
				return apis.NewNotFoundError("", err)
			}

			return c.HTML(http.StatusOK, html)
		}, wrappedApp.authMiddleware)

		return nil
	})

	// // fires for every auth collection
	app.OnRecordAuthRequest().Add(func(e *core.RecordAuthEvent) error {
		c := (e.HttpContext)

		cookie := &http.Cookie{
			Name:     "AuthToken",
			Value:    e.Token,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		}
		c.SetCookie(cookie)
		c.Response().Header().Set("HX-Redirect", "/home")
		return c.String(http.StatusOK, "Loading...")
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
