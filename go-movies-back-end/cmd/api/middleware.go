package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (app *application) enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// w.Header().Set("Access-Control-Allow-Origin", "https://movies-mm.ddns.net")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		// w.Header().Set("Access-Control-Allow-Credentials", "true")
		// w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		// w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization")
		// header := w.Header()
		// header.Add("Access-Control-Allow-Origin", "http://localhost:3000")
		// fmt.Println(r.Method)
		if r.Method == "OPTIONS" {

			// header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

			// w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization")
			return
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func (app *application) enableGinCORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://movies-mm.ddns.net"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Origin", "Access-Control-Allow-Headers", "Authorization", "Access-Control-Allow-Origin", "X-CSRF-Token", "Accept", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})
}

func (app *application) authRequired() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		_, _, err := app.auth.GetTokenFromHeaderAndVerify(ctx)

		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)

			return
		}
		ctx.Next()
		// next.ServeHTTP(w, r)
	}
}

// func (app *application) authRequired(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		_, _, err := app.auth.GetTokenFromHeaderAndVerify(w, r)

// 		if err != nil {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }
