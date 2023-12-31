package main

import (
	"backend/internal/graph"
	"backend/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func (app *application) Home(ctx *gin.Context) {
	var payLoad = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Go Movies up and running",
		Version: "1.0.0",
	}

	ctx.JSON(http.StatusOK, payLoad)


}

func (app *application) AllMovies(ctx *gin.Context) {

	movies, err := app.DB.AllMovies()

	if err != nil {
		app.resErrorJSON(ctx, err)
		// fmt.Println(err)
		return
	}
	ctx.JSON(http.StatusOK, movies)

}


func (app *application) authenticate(ctx *gin.Context) {
	// read json payload
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := ctx.BindJSON(&requestPayload)
	if err != nil {
		app.resErrorJSON(ctx, err, http.StatusBadRequest)
		return
	}

	//validate user against database
	user, err := app.DB.GetUserByEmail(requestPayload.Email)

	if err != nil {
		app.resErrorJSON(ctx, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	//check password
	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.resErrorJSON(ctx, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	//create jwt user
	u := jwtUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	//generate tokens
	tokens, err := app.auth.GenerateTokenPair(&u)

	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	app.SetCookieToContext(ctx, refreshCookie)
	ctx.JSON(http.StatusAccepted, tokens)
}

func (app *application) refreshToken(ctx *gin.Context) {
	cookieValue, err := ctx.Cookie(app.auth.CookieName)
	cookiews := ctx.Request.Cookies()
	for _, elmt := range cookiews {
		if elmt.Name == app.auth.CookieName {
			fmt.Println("Cookie Name correct : ", elmt.Name)
		} else {
			fmt.Println("Cookie Name in-correct : ", elmt.Name)
		}
	}
	if err != nil {
		app.resErrorJSON(ctx, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}

	claims := &Claims{}
	refreshToken := cookieValue
	println(app.auth.CookieName)

	//parse the token to get the claims
	_, err = jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(app.JWTSecret), nil
	})

	if err != nil {
		app.resErrorJSON(ctx, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}

	//get the user id from the token claims
	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		app.resErrorJSON(ctx, errors.New("unknow user"), http.StatusUnauthorized)
		return
	}

	user, err := app.DB.GetUserByID(userID)
	if err != nil {
		app.resErrorJSON(ctx, errors.New("unknow user"), http.StatusUnauthorized)
		return
	}

	u := jwtUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	tokenPairs, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		app.resErrorJSON(ctx, errors.New("error generating tokens"), http.StatusUnauthorized)
		return
	}

	app.SetCookieToContext(ctx, app.auth.GetRefreshCookie(tokenPairs.RefreshToken))

	ctx.JSON(http.StatusOK, tokenPairs)

}


func (app *application) logout(ctx *gin.Context) {
	app.SetCookieToContext(ctx, app.auth.GetExpiredRefreshCookie())
	ctx.Writer.WriteHeader(http.StatusAccepted)
}


func (app *application) MovieCatalog(ctx *gin.Context) {
	movies, err := app.DB.AllMovies()

	if err != nil {
		app.resErrorJSON(ctx, err)
		// fmt.Println(err)
		return
	}

	ctx.JSON(http.StatusOK, movies)
}


func (app *application) GetMovie(ctx *gin.Context) {

	// id := chi.URLParam(r, "id")
	id := ctx.Param("id")
	movieID, err := strconv.Atoi(id)

	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	movie, err := app.DB.OneMovie(movieID)
	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, movie)

	// _ = app.writeJSON(w, http.StatusOK, movie)

}


func (app *application) MovieForEdit(ctx *gin.Context) {
	id := ctx.Param("id")
	movieID, err := strconv.Atoi(id)

	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	movie, genres, err := app.DB.OneMovieForEdit(movieID)
	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	var payload = struct {
		Movie  *models.Movie   `json:"movie"`
		Genres []*models.Genre `json:"genres"`
	}{
		movie,
		genres,
	}

	ctx.JSON(http.StatusOK, payload)

}


func (app *application) AllGenres(ctx *gin.Context) {
	genres, err := app.DB.AllGenres()

	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, genres)
	// _ = app.writeJSON(w, http.StatusOK, genres)
}


func (app *application) InsertMovie(ctx *gin.Context) {
	var movie models.Movie

	err := ctx.BindJSON(&movie)

	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	//try to get an image
	movie = app.getPoster(movie)

	movie.CreatedAt = time.Now()
	movie.UpdatedAt = time.Now()

	newID, err := app.DB.InsertMovie(movie)
	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	//now handle genres
	err = app.DB.UpdateMovieGenres(newID, movie.GenreArray)
	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	resp := JSONResponse{
		Error:   false,
		Message: "movie updated!",
	}

	ctx.JSON(http.StatusAccepted, resp)
}


func (app *application) getPoster(movie models.Movie) models.Movie {
	type TheMovieDB struct {
		Page    int `json:"page"`
		Results []struct {
			PosterPath string `json:"poster_path"`
		} `json:"results"`

		TotalPages int `json:"total_pages"`
	}

	client := &http.Client{}
	theUrl := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s", app.APIKey)

	req, err := http.NewRequest("GET", theUrl+"&query="+url.QueryEscape(movie.Title), nil)

	if err != nil {
		log.Println(err)
		return movie
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return movie
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return movie
	}

	var responseObject TheMovieDB

	json.Unmarshal(bodyBytes, &responseObject)

	if len(responseObject.Results) > 0 {
		movie.Image = responseObject.Results[0].PosterPath
	}

	return movie

}

func (app *application) UpdateMovie(ctx *gin.Context) {
	var payload models.Movie
	err := ctx.BindJSON(&payload)

	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	movie, err := app.DB.OneMovie(payload.ID)

	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	movie.Title = payload.Title
	movie.ReleaseDate = payload.ReleaseDate
	movie.Description = payload.Description
	movie.MPAARating = payload.MPAARating
	movie.Runtime = payload.Runtime
	movie.UpdatedAt = time.Now()

	err = app.DB.UpdateMovie(*movie)
	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	err = app.DB.UpdateMovieGenres(movie.ID, payload.GenreArray)
	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	resp := JSONResponse{
		Error:   false,
		Message: "movie updated",
	}
	ctx.JSON(http.StatusAccepted, resp)
}


func (app *application) DeleteMovie(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	err = app.DB.DeleteMovie(id)
	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	resp := JSONResponse{
		Error:   false,
		Message: "movie deleted",
	}

	ctx.JSON(http.StatusAccepted, resp)

}


func (app *application) AllMoviesByGenre(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	movies, err := app.DB.AllMovies(id)
	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, movies)

	// app.writeJSON(w, http.StatusOK, movies)
}


func (app *application) moviesGraphQL(ctx *gin.Context) {
	// we need to poulate our Graph type with the movies
	movies, _ := app.DB.AllMovies()

	//get the query from the request

	var bodyContent string

	err := ctx.ShouldBind(&bodyContent)
	if err != nil {
		app.resErrorJSON(ctx, err, http.StatusBadRequest)
		return
	}
	query := bodyContent

	//create a new variable of type *graph.Grap
	g := graph.New(movies)

	//set the query stering on the variable
	g.QueryString = query

	//perform the query
	resp, err := g.Query()
	if err != nil {
		app.resErrorJSON(ctx, err)
		return
	}

	//send the response
	j, _ := json.MarshalIndent(resp, "", "\t")
	ctx.Header("Content-Type", "application/json")
	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Writer.Write(j)
}

