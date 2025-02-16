package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/global"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/util"
)

func (appContext *AppContext) deleteEnvironment(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	log.Println("Deleting environment: ", vars["id"])

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println(global.ParameterIDError, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	env, err := appContext.Repositories.EnvironmentDAO.GetByID(id)
	if err != nil {
		log.Println("Error retrieving environment by ID: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.Repositories.EnvironmentDAO.DeleteEnvironment(*env); err != nil {
		log.Println("Error deleting environment: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := removeEnvironmentFile(env.Group + "_" + env.Name); err != nil {
		log.Println("Error deleting environment file: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) editEnvironment(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
		return
	}

	var payload model.DataElement

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	env := payload.Data

	result, err := appContext.Repositories.EnvironmentDAO.GetByID(int(env.ID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oldFile := result.Group + "_" + result.Name
	removeEnvironmentFile(oldFile)

	createEnvironmentFile(env.Name, env.Token, appContext.K8sConfigPath+env.Group+"_"+env.Name,
		env.CACertificate, env.ClusterURI, env.Namespace)

	if err := appContext.Repositories.EnvironmentDAO.EditEnvironment(env); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) duplicateEnvironments(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	log.Println("Duplicating environment: ", vars["id"])

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println(global.ParameterIDError, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	variables, err := appContext.Repositories.VariableDAO.GetAllVariablesByEnvironment(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var env model.Environment
	env.Namespace = environment.Namespace
	env.Name = environment.Name + "-Copy"
	env.Group = environment.Group
	env.CACertificate = environment.CACertificate
	env.Token = environment.Token
	env.ClusterURI = environment.ClusterURI
	env.Gateway = environment.Gateway

	createEnvironmentFile(env.Name, env.Token, appContext.K8sConfigPath+env.Group+"_"+env.Name,
		env.CACertificate, env.ClusterURI, env.Namespace)

	var envID int
	if envID, err = appContext.Repositories.EnvironmentDAO.CreateEnvironment(env); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var newVariable *model.Variable
	for _, variable := range variables {
		newVariable = &model.Variable{}
		newVariable.Name = variable.Name
		newVariable.EnvironmentID = envID
		newVariable.Value = variable.Value
		newVariable.Description = variable.Description
		newVariable.Scope = variable.Scope

		if _, _, err := appContext.Repositories.VariableDAO.CreateVariable(*newVariable); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) addEnvironments(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
	}

	var payload model.DataElement

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	env := payload.Data

	createEnvironmentFile(env.Name, env.Token, appContext.K8sConfigPath+env.Group+"_"+env.Name,
		env.CACertificate, env.ClusterURI, env.Namespace)

	if _, err := appContext.Repositories.EnvironmentDAO.CreateEnvironment(env); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) getEnvironments(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)

	envResult := &model.EnvResult{}

	if len(principal.Email) <= 0 {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusMethodNotAllowed)
		return
	}

	var err error
	if envResult.Envs, err = appContext.Repositories.EnvironmentDAO.GetAllEnvironments(principal.Email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(envResult)
	w.Write(data)

}

func (appContext *AppContext) export(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, "text/plain; charset=UTF-8")
	w.Header().Set("Content-Disposition", "attachment; filename=environment.txt")

	vars := mux.Vars(r)
	log.Println("Deleting environment: ", vars["id"])

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println(global.ParameterIDError, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var variables []model.Variable
	if variables, err = appContext.Repositories.VariableDAO.GetAllVariablesByEnvironment(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ibid := bytes.NewBufferString("\n")

	for _, element := range variables {
		ibid.WriteString(element.Scope + " " + element.Name + "=" + element.Value + "\n")
	}

	w.WriteHeader(http.StatusOK)
	w.Write(ibid.Bytes())

}

func (appContext *AppContext) getAllEnvironments(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	envResult := &model.EnvResult{}

	var err error
	if envResult.Envs, err = appContext.Repositories.EnvironmentDAO.GetAllEnvironments(""); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(envResult)
	w.Write(data)
}

func removeEnvironmentFile(fileName string) error {
	log.Println("Removing file: " + fileName)

	if _, err := os.Stat("./" + fileName); err == nil {
		err := os.Remove("./" + fileName)
		if err != nil {
			log.Println("Error removing file", err)
			return err
		}
	}
	return nil
}

//CreateEnvironmentFile export function "createEnvironmentFile" for other packages
func CreateEnvironmentFile(clusterName, clusterUserToken, fileName, ca, server, namespace string) {
	createEnvironmentFile(clusterName, clusterUserToken, fileName, ca, server, namespace)
}

func createEnvironmentFile(clusterName string, clusterUserToken string,
	fileName string, ca string, server string, namespace string) {

	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	ca = strings.TrimSuffix(ca, "\n")
	caBase64 := base64.StdEncoding.EncodeToString([]byte(ca))

	startIndex := strings.Index(clusterUserToken, "kubeconfig-")
	clusterUser := "xpto"
	if startIndex > 0 {
		startIndex = startIndex + 11
		endIndex := strings.Index(clusterUserToken, ":")
		clusterUser = clusterUserToken[startIndex:endIndex]
	}

	file.WriteString("apiVersion: v1\n")
	file.WriteString("clusters:\n")
	file.WriteString("- cluster:\n")
	file.WriteString("    certificate-authority-data: " + caBase64 + "\n")
	file.WriteString("    server: " + server + "\n")
	file.WriteString("  name: " + clusterName + "\n")
	file.WriteString("contexts:\n")
	file.WriteString("- context:\n")
	file.WriteString("    cluster: " + clusterName + "\n")
	file.WriteString("    namespace: " + namespace + "\n")
	file.WriteString("    user: " + clusterUser + "\n")
	file.WriteString("  name: " + clusterName + "\n")
	file.WriteString("current-context: " + clusterName + "\n")
	file.WriteString("kind: Config\n")
	file.WriteString("preferences: {}\n")
	file.WriteString("users:\n")
	file.WriteString("- name: " + clusterUser + "\n")
	file.WriteString("  user:\n")
	file.WriteString("    token: " + clusterUserToken + "\n")

}
