package main

import (
	"github.com/gin-gonic/gin"
)

func (app *application) routes() *gin.Engine {
	
	router := gin.Default()

	router.Use(app.enableGinCORS())

	router.GET("/", app.Home)

	router.POST("/authenticate", app.authenticate)
	router.GET("/refresh", app.refreshToken)
	router.GET("/logout", app.logout)

	router.GET("/movies", app.AllMovies)
	router.GET("/movies/:id", app.GetMovie)
	router.GET(`/genres`, app.AllGenres)
	router.GET(`/movies/genres/:id`, app.AllMoviesByGenre)

	router.POST("/graph", app.moviesGraphQL)

	adminGroup := router.Group("/admin", app.authRequired())

	adminGroup.GET("/movies", app.MovieCatalog)
	adminGroup.GET("/movies/:id", app.MovieForEdit)
	adminGroup.PUT("/movies/0", app.InsertMovie)
	adminGroup.PATCH("/movies/:id", app.UpdateMovie)
	adminGroup.DELETE("/movies/:id", app.DeleteMovie)
	return router

}
