package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/CodyGuo/semaphore/api/projects"
	"github.com/CodyGuo/semaphore/api/sockets"
	"github.com/CodyGuo/semaphore/api/tasks"

	"github.com/CodyGuo/semaphore/util"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
)

var publicAssets = packr.NewBox("../web/public")

//JSONMiddleware ensures that all the routes respond with Json, this is added by default to all routes
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		next.ServeHTTP(w, r)
	})
}

//plainTextMiddleware resets headers to Plain Text if needed
func plainTextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/plain; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func pongHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 not found"))
	fmt.Println(r.Method, ":", r.URL.String(), "--> 404 Not Found")
}

// Route declares all routes
func Route() *mux.Router {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(servePublic)

	webPath := "/"
	if util.WebHostURL != nil {
		webPath = util.WebHostURL.RequestURI()
	}

	r.Use(mux.CORSMethodMiddleware(r), JSONMiddleware)
	r.HandleFunc(webPath, http.HandlerFunc(servePublic))
	r.HandleFunc("/api/auth/login", login).Methods("POST")

	ping := r.PathPrefix(webPath + "api/ping").Subrouter()
	ping.Use(plainTextMiddleware)
	ping.HandleFunc("/api/ping", pongHandler).Methods("GET", "HEAD")

	// set up the namespace
	api := r.PathPrefix(webPath + "api").Subrouter()
	api.Use(authentication)
	api.HandleFunc("/auth/logout", logout).Methods("POST")
	api.HandleFunc("/ws", sockets.Handler).Methods("GET", "HEAD")
	api.HandleFunc("/info", getSystemInfo).Methods("GET", "HEAD")
	api.HandleFunc("/upgrade", checkUpgrade).Methods("GET", "HEAD")
	api.HandleFunc("/upgrade", doUpgrade).Methods("POST")

	api.HandleFunc("/user", getUser).Methods("GET", "HEAD")
	api.HandleFunc("/users", getUsers).Methods("GET", "HEAD")
	api.HandleFunc("/projects", projects.GetProjects).Methods("GET", "HEAD")
	api.HandleFunc("/events", getAllEvents).Methods("GET", "HEAD")
	api.HandleFunc("/events/last", getLastEvents).Methods("GET", "HEAD")

	user := api.PathPrefix("/user").Subrouter()
	user.HandleFunc("/tokens", getAPITokens).Methods("GET", "HEAD")
	user.HandleFunc("/tokens", createAPIToken).Methods("POST")
	user.HandleFunc("/tokens/{token_id}", expireAPIToken).Methods("DELETE")

	users := api.PathPrefix("/users").Subrouter()
	users.Use(getUserMiddleware)
	users.HandleFunc("/users", addUser).Methods("POST")
	users.HandleFunc("/{user_id}", getUser).Methods("GET", "HEAD")
	users.HandleFunc("/{user_id}", updateUser).Methods("PUT")
	users.HandleFunc("/{user_id}/password", updateUserPassword).Methods("POST")
	users.HandleFunc("/{user_id}", deleteUser).Methods("DELETE")

	apiProjects := api.PathPrefix("/projects").Subrouter()
	apiProjects.Use(mustBeAdminMiddleware)
	apiProjects.HandleFunc("", projects.AddProject).Methods("POST")

	apiProject := api.PathPrefix("/project/{project_id}").Subrouter()
	apiProject.Use(projects.ProjectMiddleware)
	apiProject.HandleFunc("", projects.GetProject).Methods("GET", "HEAD")
	apiProject.HandleFunc("/events/last", getLastEvents).Methods("GET", "HEAD")
	apiProject.HandleFunc("/events", getAllEvents).Methods("GET", "HEAD")
	apiProject.HandleFunc("/tasks/last", tasks.GetLastTasks).Methods("GET", "HEAD")
	apiProject.HandleFunc("/tasks", tasks.GetAllTasks).Methods("GET", "HEAD")
	apiProject.HandleFunc("/templates", projects.GetTemplates).Methods("GET", "HEAD")
	apiProject.HandleFunc("/inventory", projects.GetInventory).Methods("GET", "HEAD")
	apiProject.HandleFunc("/environment", projects.GetEnvironment).Methods("GET", "HEAD")
	apiProject.HandleFunc("/keys", projects.GetKeys).Methods("GET", "HEAD")
	apiProject.HandleFunc("/repositories", projects.GetRepositories).Methods("GET", "HEAD")
	apiProject.HandleFunc("/users", projects.GetUsers).Methods("GET", "HEAD")

	projectManagement := apiProject.Path("").Subrouter()
	projectManagement.Use(mustBeAdminMiddleware)
	projectManagement.HandleFunc("", projects.UpdateProject).Methods("PUT")
	projectManagement.HandleFunc("", projects.DeleteProject).Methods("DELETE")

	templates := apiProject.PathPrefix("/templates").Subrouter()
	templates.Use(projects.MustBeAdmin)
	templates.HandleFunc("", projects.AddTemplate).Methods("POST")

	templatesManagement := templates.PathPrefix("/{template_id}").Subrouter()
	templatesManagement.Use(projects.TemplatesMiddleware)
	templatesManagement.HandleFunc("", projects.UpdateTemplate).Methods("PUT")
	templatesManagement.HandleFunc("", projects.RemoveTemplate).Methods("DELETE")

	apiProject.HandleFunc("/tasks", tasks.AddTask).Methods("POST")
	projectTasks := apiProject.PathPrefix("/tasks").Subrouter()
	projectTasks.Use(tasks.GetTaskMiddleware)
	projectTasks.HandleFunc("/{task_id}", tasks.GetTask).Methods("GET", "HEAD")
	projectTasks.HandleFunc("/{task_id}/output", tasks.GetTaskOutput).Methods("GET", "HEAD")

	tasksManagement := projectTasks.PathPrefix("/{task_id}").Subrouter()
	tasksManagement.Use(projects.MustBeAdmin)
	tasksManagement.HandleFunc("", tasks.RemoveTask).Methods("DELETE")

	inventory := apiProject.PathPrefix("/inventory").Subrouter()
	inventory.Use(projects.MustBeAdmin)
	inventory.HandleFunc("", projects.AddInventory).Methods("POST")

	inventoryManagement := inventory.PathPrefix("/{inventory_id}").Subrouter()
	inventoryManagement.Use(projects.InventoryMiddleware)
	inventoryManagement.HandleFunc("", projects.UpdateInventory).Methods("PUT")
	inventoryManagement.HandleFunc("", projects.RemoveInventory).Methods("DELETE")

	environment := apiProject.PathPrefix("/environment").Subrouter()
	environment.Use(projects.MustBeAdmin)
	environment.HandleFunc("", projects.AddEnvironment).Methods("POST")

	environmentManager := environment.PathPrefix("/{environment_id}").Subrouter()
	environmentManager.Use(projects.EnvironmentMiddleware)
	environmentManager.HandleFunc("", projects.UpdateEnvironment).Methods("PUT")
	environmentManager.HandleFunc("", projects.RemoveEnvironment).Methods("DELETE")

	keys := apiProject.PathPrefix("/keys").Subrouter()
	keys.Use(projects.MustBeAdmin)
	keys.HandleFunc("", projects.AddKey).Methods("POST")

	keysManagement := keys.PathPrefix("/{key_id}").Subrouter()
	keysManagement.Use(projects.KeyMiddleware)
	keysManagement.HandleFunc("", projects.UpdateKey).Methods("PUT")
	keysManagement.HandleFunc("", projects.RemoveKey).Methods("DELETE")

	repositories := apiProject.PathPrefix("/repositories").Subrouter()
	repositories.Use(projects.MustBeAdmin)
	repositories.HandleFunc("", projects.AddRepository).Methods("POST")

	repositoriesManagement := repositories.PathPrefix("/{repository_id}").Subrouter()
	repositoriesManagement.Use(projects.RepositoryMiddleware)
	repositoriesManagement.HandleFunc("", projects.UpdateRepository).Methods("PUT")
	repositoriesManagement.HandleFunc("", projects.RemoveRepository).Methods("DELETE")

	projectUsers := apiProject.PathPrefix("/users").Subrouter()
	projectUsers.Use(projects.MustBeAdmin)
	projectUsers.HandleFunc("", projects.AddUser).Methods("POST")

	usersManagement := projectUsers.PathPrefix("/{user_id}").Subrouter()
	usersManagement.Use(projects.UserMiddleware)
	usersManagement.HandleFunc("", projects.RemoveUser).Methods("DELETE")
	usersManagement.HandleFunc("/admin", projects.MakeUserAdmin).Methods("POST")
	usersManagement.HandleFunc("/admin", projects.MakeUserAdmin).Methods("DELETE")

	return r
}

//nolint: gocyclo
func servePublic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, "/api") {
		notFoundHandler(w, r)
		return
	}

	webPath := "/"
	if util.WebHostURL != nil {
		webPath = util.WebHostURL.RequestURI()
	}

	if !strings.HasPrefix(path, webPath+"public") {
		if len(strings.Split(path, ".")) > 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		path = "/html/index.html"
	}

	path = strings.Replace(path, webPath+"public/", "", 1)
	split := strings.Split(path, ".")
	suffix := split[len(split)-1]

	res, err := publicAssets.MustBytes(path)
	if err != nil {
		notFoundHandler(w, r)
		return
	}

	// replace base path
	if util.WebHostURL != nil && path == "/html/index.html" {
		res = []byte(strings.Replace(string(res),
			"<base href=\"/\">",
			"<base href=\""+util.WebHostURL.String()+"\">",
			1))
	}

	contentType := "text/plain"
	switch suffix {
	case "png":
		contentType = "image/png"
	case "jpg", "jpeg":
		contentType = "image/jpeg"
	case "gif":
		contentType = "image/gif"
	case "js":
		contentType = "application/javascript"
	case "css":
		contentType = "text/css"
	case "woff":
		contentType = "application/x-font-woff"
	case "ttf":
		contentType = "application/x-font-ttf"
	case "otf":
		contentType = "application/x-font-otf"
	case "html":
		contentType = "text/html"
	}

	w.Header().Set("content-type", contentType)
	_, err = w.Write(res)
	util.LogWarning(err)
}

func getSystemInfo(w http.ResponseWriter, r *http.Request) {
	body := map[string]interface{}{
		"version": util.Version,
		"update":  util.UpdateAvailable,
		"config": map[string]string{
			"dbHost":  util.Config.MySQL.Hostname,
			"dbName":  util.Config.MySQL.DbName,
			"dbUser":  util.Config.MySQL.Username,
			"path":    util.Config.TmpPath,
			"cmdPath": util.FindSemaphore(),
		},
	}

	if util.UpdateAvailable != nil {
		body["updateBody"] = string(blackfriday.MarkdownCommon([]byte(*util.UpdateAvailable.Body)))
	}

	util.WriteJSON(w, http.StatusOK, body)
}

func checkUpgrade(w http.ResponseWriter, r *http.Request) {
	if err := util.CheckUpdate(util.Version); err != nil {
		util.WriteJSON(w, 500, err)
		return
	}

	if util.UpdateAvailable != nil {
		getSystemInfo(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func doUpgrade(w http.ResponseWriter, r *http.Request) {
	util.LogError(util.DoUpgrade(util.Version))
}
