package projects

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/CodyGuo/semaphore/db"

	"github.com/CodyGuo/semaphore/util"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

// TemplatesMiddleware ensures a template exists and loads it to the context
func TemplatesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		templateID, err := util.GetIntParam("template_id", w, r)
		if err != nil {
			return
		}

		var template db.Template
		if err := db.Mysql.SelectOne(&template, "select * from project__template where project_id=? and id=?", project.ID, templateID); err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			panic(err)
		}

		context.Set(r, "template", template)
		next.ServeHTTP(w, r)
	})
}

// GetTemplates returns all templates for a project in a sort order
func GetTemplates(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var templates []db.Template

	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	if order != asc && order != desc {
		order = asc
	}

	q := squirrel.Select("pt.id",
		"pt.ssh_key_id",
		"pt.project_id",
		"pt.inventory_id",
		"pt.repository_id",
		"pt.environment_id",
		"pt.alias",
		"pt.playbook",
		"pt.arguments",
		"pt.override_args").
		From("project__template pt")

	switch sort {
	case "alias", "playbook":
		q = q.Where("pt.project_id=?", project.ID).
			OrderBy("pt." + sort + " " + order)
	case "ssh_key":
		q = q.LeftJoin("access_key ak ON (pt.ssh_key_id = ak.id)").
			Where("pt.project_id=?", project.ID).
			OrderBy("ak.name " + order)
	case "inventory":
		q = q.LeftJoin("project__inventory pi ON (pt.inventory_id = pi.id)").
			Where("pt.project_id=?", project.ID).
			OrderBy("pi.name " + order)
	case "environment":
		q = q.LeftJoin("project__environment pe ON (pt.environment_id = pe.id)").
			Where("pt.project_id=?", project.ID).
			OrderBy("pe.name " + order)
	case "repository":
		q = q.LeftJoin("project__repository pr ON (pt.repository_id = pr.id)").
			Where("pt.project_id=?", project.ID).
			OrderBy("pr.name " + order)
	default:
		q = q.Where("pt.project_id=?", project.ID).
			OrderBy("pt.alias " + order)
	}

	query, args, err := q.ToSql()
	util.LogWarning(err)

	if _, err := db.Mysql.Select(&templates, query, args...); err != nil {
		panic(err)
	}

	util.WriteJSON(w, http.StatusOK, templates)
}

// AddTemplate adds a template to the database
func AddTemplate(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	user := context.Get(r, "user").(*db.User)

	var tpl db.Template
	if err := util.Bind(w, r, &tpl); err != nil {
		return
	}

	res, err := db.Mysql.Exec("insert into project__template set ssh_key_id=?, project_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=?", tpl.SSHKeyID, project.ID, tpl.InventoryID, tpl.RepositoryID, tpl.EnvironmentID, tpl.Alias, tpl.Playbook, tpl.Arguments, tpl.OverrideArguments)
	if err != nil {
		panic(err)
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	tpl.ID = int(insertID)

	objType := "template"
	desc := "Template Name " + tpl.Alias + "(" + strconv.Itoa(tpl.ID) + ") created By " + user.Name
	if err := (db.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &tpl.ID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	util.WriteJSON(w, http.StatusCreated, tpl)
}

func GetTemplate(w http.ResponseWriter, r *http.Request) {
	util.WriteJSON(w, http.StatusOK, context.Get(r, "template"))
}

// UpdateTemplate writes a template to an existing key in the database
func UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	oldTpl := context.Get(r, "template").(db.Template)
	user := context.Get(r, "user").(*db.User)

	var tpl db.Template
	if err := util.Bind(w, r, &tpl); err != nil {
		return
	}

	if tpl.Arguments != nil && *tpl.Arguments == "" {
		tpl.Arguments = nil
	}

	if _, err := db.Mysql.Exec("update project__template set ssh_key_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=? where id=?", tpl.SSHKeyID, tpl.InventoryID, tpl.RepositoryID, tpl.EnvironmentID, tpl.Alias, tpl.Playbook, tpl.Arguments, tpl.OverrideArguments, oldTpl.ID); err != nil {
		panic(err)
	}

	objType := "template"
	desc := "Template Name " + tpl.Alias + "(" + strconv.Itoa(tpl.ID) + ") updated By " + user.Name
	if err := (db.Event{
		ProjectID:   &oldTpl.ProjectID,
		Description: &desc,
		ObjectID:    &oldTpl.ID,
		ObjectType:  &objType,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveTemplate deletes a template from the database
func RemoveTemplate(w http.ResponseWriter, r *http.Request) {
	tpl := context.Get(r, "template").(db.Template)
	user := context.Get(r, "user").(*db.User)

	if _, err := db.Mysql.Exec("delete from project__template where id=?", tpl.ID); err != nil {
		panic(err)
	}

	desc := "Template Name " + tpl.Alias + "(" + strconv.Itoa(tpl.ID) + ") deleted By " + user.Name
	if err := (db.Event{
		ProjectID:   &tpl.ProjectID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
